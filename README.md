# Vitalis

Vitalis is a lightweight, self-hosted system monitoring agent with a REST API, written in Go (Gin).

A background worker (asynq + Redis) periodically samples host metrics — CPU, RAM, network, and disk — via `gopsutil` and stores them in PostgreSQL. A Gin-based HTTP API lets you query this data with filtering and pagination, protected by tokens issued against a shared secret.

## Contents

- [How it works](#how-it-works)
- [Requirements](#requirements)
- [Quick start (Docker Compose)](#quick-start-docker-compose)
- [Running locally without Docker](#running-locally-without-docker)
- [Environment variables](#environment-variables)
- [Deployment](#deployment)
- [API](#api)
  - [Authentication](#authentication)
  - [Health check](#health-check)
  - [Metrics endpoints](#metrics-endpoints)
  - [Filters](#filters)
- [Data model](#data-model)
- [Makefile](#makefile)
- [Development](#development)

## How it works

1. On startup, `internal/enviroment` loads `.env` (`godotenv`). If `SECRET_KEY` is empty, a new 32-byte hex secret is generated, appended to `.env`, logged, and the application **exits** (`os.Exit(0)`). This is a first-run bootstrap step — you need to restart the app for the secret to take effect.
2. `internal/tasks` starts an asynq scheduler and worker (connected to Redis) that collect CPU/RAM/network/disk metrics every `COLLECT_*_INFO_INTERVAL_SECONDS` seconds, and also runs an initial collection immediately at startup.
3. `internal/database` opens a PostgreSQL connection and creates tables via `CREATE TABLE IF NOT EXISTS` (there's no separate migration tool — the schema is idempotently applied on every startup).
4. A Gin server is started on `RUN_ADDRESS`, serving the collected metrics through a REST API protected by bearer tokens.

Each collection cycle gets its own `group_id` (from a dedicated Postgres sequence per metric type), so rows from a single collection run (e.g. all disk partitions in one "file" snapshot) can be grouped together and pruned together once `MAX_INFO_GROUPS_AMOUNT` is exceeded.

## Requirements

- Go 1.26+ (for local build/development)
- PostgreSQL 16+
- Redis 7+
- Docker and Docker Compose (for containerized deployment)

## Quick start (Docker Compose)

```bash
# DONT FORGET TO CHANGE "DATABASE_PASSWORD" INSIDE .env FILE
mkdir vitalis-server
cd vitalis-server
wget https://raw.githubusercontent.com/yourmomahegao/vitalis/refs/heads/main/.env.example
wget https://raw.githubusercontent.com/yourmomahegao/vitalis/refs/heads/main/docker-compose.yml
cp .env.example .env
chmod 666 .env
docker compose up -d
```

Compose brings up three services:

| Service | Image | Port (host:container) | Purpose |
|---|---|---|---|
| `vitalis-postgres` | `postgres:16-alpine` | `${DATABASE_PORT:-5432}:5432` | database |
| `vitalis-redis` | `redis:7-alpine` | `${REDIS_PORT:-6379}:6379` | asynq task queue |
| `vitalis` | built from `internal/deployments/Dockerfile` | `8080:8080` | the API server itself |

The `vitalis` service waits (`depends_on: condition: service_healthy`) for Postgres and Redis to become healthy. Inside the compose network, the DB/Redis addresses are overridden to `vitalis-postgres`/`vitalis-redis`, regardless of what's in `.env`.

⚠️ Since the app auto-generates a secret and exits on first run without `SECRET_KEY`, the `vitalis` container may restart once on the **first** stack startup — this is expected (Docker restarts the container, which then picks up the secret just written to `.env`). The secret is written to the host file through the mounted `./.env:/app/.env:ro` volume, so it persists and will show up in your local `.env` after the restart.

Application logs:

```bash
make docker-logs
```

## Running locally without Docker

```bash
cp .env.example .env
# point DATABASE_ADDRESS/PORT, REDIS_ADDRESS/PORT at your local Postgres/Redis instances
go run ./cmd/api
```

For hot-reload during development, [air](https://github.com/air-verse/air) is used (config in `.air.toml`, builds the binary to `tmp/main`):

```bash
air
```

## Environment variables

Defined in `.env.example` and parsed in `internal/enviroment/enviroment.go`. If a variable is unset, the code-level default (in parentheses) is used.

| Variable | Purpose | Default in code |
|---|---|---|
| `GIN_DEBUG` | `true`/`false` — Gin mode (Debug/Release) | `false` |
| `REDIS_ADDRESS` | Redis host for the asynq queue | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `DATABASE_ADDRESS` | PostgreSQL host | `localhost` |
| `DATABASE_PORT` | PostgreSQL port | `5432` |
| `DATABASE_NAME` | database name | `vitalis` |
| `DATABASE_USER` | database user | `root` |
| `DATABASE_PASSWORD` | database password | `""` |
| `RUN_ADDRESS` | address:port the HTTP server listens on | `0.0.0.0:8080` |
| `SECRET_KEY` | shared secret used to issue access tokens (`/auth/token/`); auto-generated if left empty | auto-generated |
| `COLLECT_CPU_INFO_INTERVAL_SECONDS` | CPU metric collection interval, seconds | `30` |
| `COLLECT_RAM_INFO_INTERVAL_SECONDS` | RAM metric collection interval, seconds | `30` |
| `COLLECT_NET_INFO_INTERVAL_SECONDS` | network metric collection interval, seconds | `30` |
| `COLLECT_FILE_INFO_INTERVAL_SECONDS` | disk metric collection interval, seconds | `30` |
| `MAX_INFO_GROUPS_AMOUNT` | how many recent "snapshots" (group_id) to retain per metric type; older ones are pruned | `100` |
| `MAX_SESSION_KEYS_AMOUNT` | how many live access tokens to keep at once; older ones are pruned | `60` |
| `ACCESS_TOKEN_LIFETIME_MINUTES` | access token lifetime, minutes | `60` |

`.env` is already listed in `.gitignore` (along with `.env.*`, except `.env.example`), so secrets never end up in the repository.

## Deployment

The production image is built in multiple stages from `internal/deployments/Dockerfile`:

1. **builder** (`golang:1.26-alpine`) — `go mod download` + a static build (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) of the `./cmd/api` binary.
2. **runtime** (`alpine:3.20`) — just `ca-certificates` and the binary, run as the unprivileged `vitalis` user, `EXPOSE 8080`.

Build the image manually:

```bash
make docker-build   # docker build -f internal/deployments/Dockerfile -t vitalis .
```

Full deployment cycle via Compose:

```bash
make docker-up     # docker compose up --build -d
make docker-logs   # tail logs of the vitalis container
make docker-down   # stop and tear down the stack
```

To deploy outside of Compose (e.g. on a bare host or in Kubernetes), the app needs:
- a reachable PostgreSQL instance with the target database already created (the schema is created automatically at startup);
- a reachable Redis instance;
- a `.env` file mounted/present next to the working directory (or equivalent environment variables — but loading specifically goes through `godotenv.Load()`, which looks for `.env` in the current directory).

## API

All responses are JSON wrapped in a single envelope:

```json
{
  "status": true,
  "message": "...",
  "data": { }
}
```

`data` is not present in every response (`omitempty`).

### Authentication

Access to protected endpoints requires an `Authorization: Bearer <access_token>` header.

**`GET /auth/token/`** — issue a new access token.

The `secret_key` parameter (a form-encoded field in the request body) must match `SECRET_KEY` from `.env` (compared using constant-time comparison).

```bash
curl -X GET http://localhost:8080/auth/token/ \
  --data-urlencode "secret_key=<your SECRET_KEY>"
```

200 response:
```json
{
  "status": true,
  "message": "Generated new access-token",
  "data": {
    "access_token": "a1b2c3...", 
    "valid_until": "2026-07-17T13:00:00Z"
  }
}
```

Error codes: `400` — `secret_key` missing; `401` — secret doesn't match; `500` — internal token generation error.

**`GET /auth/token/check/`** — validate an existing token.

```bash
curl http://localhost:8080/auth/token/check/ \
  -H "Authorization: Bearer <access_token>"
```

`200` — token is valid; `400` — header missing or malformed; `401` — token unknown or expired.

### Health check

**`GET /worker/status/`** — no authentication required, always responds with:
```json
{ "status": true, "message": "All systems online" }
```
(this is a static liveness endpoint — it does not actually check the asynq worker's real health).

### Metrics endpoints

All four are `POST`, require `Authorization: Bearer <access_token>`, accept filters as `multipart/form-data` or `x-www-form-urlencoded` fields (not a JSON body), and return an array of rows from the corresponding table.

| Method | Path | Table | Metric |
|---|---|---|---|
| POST | `/info/system/cpu/` | `info_cpu` | CPU load/frequency/cores |
| POST | `/info/system/ram/` | `info_ram` | RAM total/used |
| POST | `/info/system/net/` | `info_net` | network traffic, connections |
| POST | `/info/system/file/` | `info_file` | disk partition usage |

Example:
```bash
curl -X POST http://localhost:8080/info/system/cpu/ \
  -H "Authorization: Bearer <access_token>" \
  -F "limit=50" \
  -F "start_datetime=2026-07-01T00:00:00Z"
```

200 response:
```json
{
  "status": true,
  "message": "Ok",
  "data": [
    {
      "id": 1, "group_id": 5, "name": "AMD Ryzen ...",
      "physical_cores": 8, "logical_cores": 16,
      "utilization": 12.3, "current_speed_mhz": 3800.0, "base_speed_mhz": 3600.0,
      "processes_amount": 240, "threads_amount": 3200, "handles_amount": 51200,
      "uptime": 86400000000000, "insertion_datetime": "2026-07-17T12:00:00Z"
    }
  ]
}
```

`400` — form/filter parsing error; `500` — database query error.

### Filters

Common to all four endpoints:

`start_datetime`, `end_datetime`, `limit_group`, `offset_group`, `limit`, `offset`

**CPU** (`/info/system/cpu/`):
`cpu_name`, `cpu_physical_cores_min/max`, `cpu_logical_cores_min/max`, `cpu_utilization_min/max`, `cpu_current_speed_mhz_min/max`, `cpu_base_speed_mhz_min/max`, `cpu_processes_amount_min/max`, `cpu_threads_amount_min/max`, `cpu_handles_amount_min/max`, `cpu_uptime_min/max`

**RAM** (`/info/system/ram/`):
`ram_total_min/max`, `ram_used_min/max`, `ram_free_min/max`, `ram_commited_min/max`, `ram_cached_min/max`

**Net** (`/info/system/net/`):
`net_bytes_sent_min/max`, `net_bytes_recv_min/max`, `net_packets_sent_min/max`, `net_packets_recv_min/max`, `net_err_in_min/max`, `net_err_out_min/max`, `net_connections_min/max`

**File** (`/info/system/file/`):
`file_path`, `file_total_min/max`, `file_used_min/max`, `file_free_min/max`, `file_used_percent_min/max`

`limit_group`/`offset_group` paginate over `group_id` (whole snapshots), while `limit`/`offset` paginate over individual rows.

## Data model

The schema is created automatically at startup (`database.Initialize()`, `CREATE TABLE IF NOT EXISTS`) — there is no separate migration tool.

- **`auth_session_keys`** — `id`, `session_key`, `creation_datetime`, `valid_until`.
- **`info_cpu`** — `id`, `group_id`, `name`, `physical_cores`, `logical_cores`, `utilization`, `current_speed_mhz`, `base_speed_mhz`, `processes_amount`, `threads_amount`, `handles_amount`, `uptime`, `insertion_datetime`.
- **`info_ram`** — `id`, `group_id`, `total`, `used`, `free`, `commited`, `cached`, `insertion_datetime`.
- **`info_net`** — `id`, `group_id`, `bytes_sent`, `bytes_recv`, `packets_sent`, `packets_recv`, `err_in`, `err_out`, `connections`, `insertion_datetime`.
- **`info_file`** — `id`, `group_id`, `path`, `total`, `used`, `free`, `used_percent`, `insertion_datetime`.

## Makefile

| Target | What it does |
|---|---|
| `make run` | `go run ./cmd/api` — run locally (requires reachable Postgres/Redis) |
| `make build` | build the binary to `bin/vitalis` |
| `make docker-build` | build the `vitalis` Docker image from `internal/deployments/Dockerfile` |
| `make docker-up` | `docker compose up --build -d` — bring up the whole stack |
| `make docker-down` | stop and remove the stack |
| `make docker-logs` | tail logs of the `vitalis` container |

## Development

Tests (`go-sqlmock`, no real Postgres required):

```bash
go test ./...
```

Covered: `internal/handlers` (auth, system-info endpoints) and part of `internal/services`. Not yet covered: `internal/database`, `internal/tasks`, `internal/enviroment`.
