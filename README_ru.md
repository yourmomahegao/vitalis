# Vitalis

Vitalis — лёгкий self-hosted агент мониторинга системы с REST API, написанный на Go (Gin).

Фоновый воркер (asynq + Redis) периодически снимает метрики хоста — CPU, RAM, сеть и диски — через `gopsutil` и сохраняет их в PostgreSQL. HTTP API поверх Gin позволяет получать эти данные с фильтрацией и пагинацией, а доступ защищён токенами, выдаваемыми по общему секрету.

## Содержание

- [Как это устроено](#как-это-устроено)
- [Требования](#требования)
- [Быстрый старт (Docker Compose)](#быстрый-старт-docker-compose)
- [Локальный запуск без Docker](#локальный-запуск-без-docker)
- [Переменные окружения](#переменные-окружения)
- [Деплой](#деплой)
- [API](#api)
  - [Аутентификация](#аутентификация)
  - [Health-check](#health-check)
  - [Эндпоинты метрик](#эндпоинты-метрик)
  - [Фильтры](#фильтры)
- [Модель данных](#модель-данных)
- [Makefile](#makefile)
- [Разработка](#разработка)

## Как это устроено

1. При старте `internal/enviroment` загружает `.env` (`godotenv`). Если `SECRET_KEY` пуст — генерируется новый 32-байтный hex-секрет, дописывается в `.env`, выводится в лог, и приложение **завершает работу** (`os.Exit(0)`). Это первичная инициализация — приложение нужно перезапустить, чтобы секрет подхватился.
2. `internal/tasks` поднимает asynq-scheduler и asynq-worker (подключение к Redis), которые каждые `COLLECT_*_INFO_INTERVAL_SECONDS` секунд собирают метрики CPU/RAM/сети/дисков и сразу же запускают первый сбор при старте.
3. `internal/database` открывает соединение с PostgreSQL и создаёт таблицы через `CREATE TABLE IF NOT EXISTS` (отдельного инструмента миграций нет — схема идемпотентно накатывается при каждом запуске).
4. Поднимается Gin-сервер на `RUN_ADDRESS`, который отдаёт накопленные метрики через REST API, защищённый bearer-токенами.

Каждый цикл сбора метрик получает свой `group_id` (из отдельной Postgres-последовательности на тип метрики), что позволяет группировать строки одного снятия (например, все разделы диска в одном "file"-снятии) и удалять их пачками при превышении `MAX_INFO_GROUPS_AMOUNT`.

## Требования

- Go 1.26+ (для локальной сборки/разработки)
- PostgreSQL 16+
- Redis 7+
- Docker и Docker Compose (для контейнерного деплоя)

## Быстрый старт (Docker Compose)

```bash
# НЕ ЗАБУДЬТЕ ИЗМЕНИТЬ "DATABASE_PASSWORD" ВЕТРИ .env ФАЙЛА
mkdir vitalis-server
cd vitalis-server
wget https://raw.githubusercontent.com/yourmomahegao/vitalis/refs/heads/main/.env.example
wget https://raw.githubusercontent.com/yourmomahegao/vitalis/refs/heads/main/docker-compose.yml
cp .env.example .env
chmod 666 .env
docker compose up -d
```

Compose поднимает три сервиса:

| Сервис | Образ | Порт (host:container) | Назначение |
|---|---|---|---|
| `vitalis-postgres` | `postgres:16-alpine` | `${DATABASE_PORT:-5432}:5432` | база данных |
| `vitalis-redis` | `redis:7-alpine` | `${REDIS_PORT:-6379}:6379` | очередь задач для asynq |
| `vitalis` | сборка из `internal/deployments/Dockerfile` | `8080:8080` | сам API-сервер |

Сервис `vitalis` ждёт (`depends_on: condition: service_healthy`) готовности Postgres и Redis, а адреса БД/Redis внутри сети compose переопределены на `vitalis-postgres`/`vitalis-redis` независимо от значений в `.env`.

⚠️ Т.к. приложение при первом запуске без `SECRET_KEY` сам генерирует секрет и завершает процесс, при **первом** поднятии стека контейнер `vitalis` может один раз перезапуститься — это ожидаемо (Docker сам перезапустит контейнер, и он подхватит только что записанный в `.env` секрет). Секрет пишется в примонтированный volume `./.env:/app/.env:ro`... файл на хосте, так что после перезапуска он сохранится и будет виден в вашем `.env`.

Логи приложения:

```bash
make docker-logs
```

## Локальный запуск без Docker

```bash
cp .env.example .env
# укажите DATABASE_ADDRESS/PORT, REDIS_ADDRESS/PORT для локальных инстансов Postgres/Redis
go run ./cmd/api
```

Для hot-reload при разработке используется [air](https://github.com/air-verse/air) (конфиг `.air.toml`, собирает бинарь в `tmp/main`):

```bash
air
```

## Переменные окружения

Определены в `.env.example` и парсятся в `internal/enviroment/enviroment.go`. Если переменная не задана — используется значение по умолчанию из кода (в скобках).

| Переменная | Назначение | Значение по умолчанию в коде |
|---|---|---|
| `GIN_DEBUG` | `true`/`false` — режим Gin (Debug/Release) | `false` |
| `REDIS_ADDRESS` | адрес Redis для очереди asynq | `localhost` |
| `REDIS_PORT` | порт Redis | `6379` |
| `DATABASE_ADDRESS` | адрес PostgreSQL | `localhost` |
| `DATABASE_PORT` | порт PostgreSQL | `5432` |
| `DATABASE_NAME` | имя базы данных | `vitalis` |
| `DATABASE_USER` | пользователь БД | `root` |
| `DATABASE_PASSWORD` | пароль БД | `""` |
| `RUN_ADDRESS` | адрес:порт, на котором слушает HTTP-сервер | `0.0.0.0:8080` |
| `SECRET_KEY` | общий секрет для выдачи access-токенов (`/auth/token/`); при пустом значении генерируется автоматически | автогенерация |
| `COLLECT_CPU_INFO_INTERVAL_SECONDS` | интервал сбора метрик CPU, сек | `30` |
| `COLLECT_RAM_INFO_INTERVAL_SECONDS` | интервал сбора метрик RAM, сек | `30` |
| `COLLECT_NET_INFO_INTERVAL_SECONDS` | интервал сбора сетевых метрик, сек | `30` |
| `COLLECT_FILE_INFO_INTERVAL_SECONDS` | интервал сбора метрик дисков, сек | `30` |
| `MAX_INFO_GROUPS_AMOUNT` | сколько последних "снятий" (group_id) хранить на тип метрики, старые удаляются | `100` |
| `MAX_SESSION_KEYS_AMOUNT` | сколько живых access-токенов хранить одновременно, старые удаляются | `60` |
| `ACCESS_TOKEN_LIFETIME_MINUTES` | время жизни access-токена, минут | `60` |

`.env` уже добавлен в `.gitignore` (вместе с `.env.*`, кроме `.env.example`) — секреты в репозиторий не попадают.

## Деплой

Продовый образ собирается многоступенчато из `internal/deployments/Dockerfile`:

1. **builder** (`golang:1.26-alpine`) — `go mod download` + статическая сборка (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) бинаря `./cmd/api`.
2. **runtime** (`alpine:3.20`) — только `ca-certificates` и бинарь, запуск от непривилегированного пользователя `vitalis`, `EXPOSE 8080`.

Собрать образ вручную:

```bash
make docker-build   # docker build -f internal/deployments/Dockerfile -t vitalis .
```

Полный цикл деплоя через Compose:

```bash
make docker-up     # docker compose up --build -d
make docker-logs   # хвост логов контейнера vitalis
make docker-down   # остановить и снести стек
```

Для деплоя вне Compose (например, на голый хост или в Kubernetes) приложению нужны:
- доступный PostgreSQL с уже созданной БД (схему создаст сам при старте);
- доступный Redis;
- смонтированный/присутствующий рядом с рабочей директорией файл `.env` (или переменные окружения, эквивалентные ему — но подгрузка идёт именно через `godotenv.Load()`, ищущий `.env` в текущей директории).

## API

Все ответы — JSON в едином конверте:

```json
{
  "status": true,
  "message": "...",
  "data": { }
}
```

`data` присутствует не во всех ответах (`omitempty`).

### Аутентификация

Доступ к защищённым эндпоинтам — по заголовку `Authorization: Bearer <access_token>`.

**`GET /auth/token/`** — выдать новый access-токен.

Параметр `secret_key` (form-encoded поле в теле запроса) должен совпадать с `SECRET_KEY` из `.env` (сравнение constant-time).

```bash
curl -X GET http://localhost:8080/auth/token/ \
  --data-urlencode "secret_key=<ваш SECRET_KEY>"
```

Ответ 200:
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

Коды ошибок: `400` — не передан `secret_key`; `401` — секрет не совпал; `500` — внутренняя ошибка генерации токена.

**`GET /auth/token/check/`** — проверить валидность токена.

```bash
curl http://localhost:8080/auth/token/check/ \
  -H "Authorization: Bearer <access_token>"
```

`200` — токен валиден; `400` — заголовок отсутствует/некорректного формата; `401` — токен неизвестен или истёк.

### Health-check

**`GET /worker/status/`** — не требует авторизации, всегда отвечает:
```json
{ "status": true, "message": "All systems online" }
```
(это статичный liveness-эндпоинт, реального состояния asynq-воркера он не проверяет).

### Эндпоинты метрик

Все четыре — `POST`, требуют `Authorization: Bearer <access_token>`, принимают фильтры как `multipart/form-data` или `x-www-form-urlencoded` поля (не JSON-тело), возвращают массив строк из соответствующей таблицы.

| Метод | Путь | Таблица | Метрика |
|---|---|---|---|
| POST | `/info/system/cpu/` | `info_cpu` | загрузка/частота/ядра CPU |
| POST | `/info/system/ram/` | `info_ram` | объём/использование RAM |
| POST | `/info/system/net/` | `info_net` | сетевой трафик, соединения |
| POST | `/info/system/file/` | `info_file` | использование дисковых разделов |

Пример:
```bash
curl -X POST http://localhost:8080/info/system/cpu/ \
  -H "Authorization: Bearer <access_token>" \
  -F "limit=50" \
  -F "start_datetime=2026-07-01T00:00:00Z"
```

Ответ 200:
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

`400` — ошибка парсинга формы/фильтров; `500` — ошибка запроса к БД.

### Фильтры

Общие для всех четырёх эндпоинтов:

`start_datetime`, `end_datetime`, `limit_group`, `offset_group`, `limit`, `offset`

**CPU** (`/info/system/cpu/`):
`cpu_name`, `cpu_physical_cores_min/max`, `cpu_logical_cores_min/max`, `cpu_utilization_min/max`, `cpu_current_speed_mhz_min/max`, `cpu_base_speed_mhz_min/max`, `cpu_processes_amount_min/max`, `cpu_threads_amount_min/max`, `cpu_handles_amount_min/max`, `cpu_uptime_min/max`

**RAM** (`/info/system/ram/`):
`ram_total_min/max`, `ram_used_min/max`, `ram_free_min/max`, `ram_commited_min/max`, `ram_cached_min/max`

**Net** (`/info/system/net/`):
`net_bytes_sent_min/max`, `net_bytes_recv_min/max`, `net_packets_sent_min/max`, `net_packets_recv_min/max`, `net_err_in_min/max`, `net_err_out_min/max`, `net_connections_min/max`

**File** (`/info/system/file/`):
`file_path`, `file_total_min/max`, `file_used_min/max`, `file_free_min/max`, `file_used_percent_min/max`

`limit_group`/`offset_group` пагинируют по `group_id` (снятиям целиком), `limit`/`offset` — по отдельным строкам.

## Модель данных

Схема создаётся автоматически при старте (`database.Initialize()`, `CREATE TABLE IF NOT EXISTS`), отдельного инструмента миграций нет.

- **`auth_session_keys`** — `id`, `session_key`, `creation_datetime`, `valid_until`.
- **`info_cpu`** — `id`, `group_id`, `name`, `physical_cores`, `logical_cores`, `utilization`, `current_speed_mhz`, `base_speed_mhz`, `processes_amount`, `threads_amount`, `handles_amount`, `uptime`, `insertion_datetime`.
- **`info_ram`** — `id`, `group_id`, `total`, `used`, `free`, `commited`, `cached`, `insertion_datetime`.
- **`info_net`** — `id`, `group_id`, `bytes_sent`, `bytes_recv`, `packets_sent`, `packets_recv`, `err_in`, `err_out`, `connections`, `insertion_datetime`.
- **`info_file`** — `id`, `group_id`, `path`, `total`, `used`, `free`, `used_percent`, `insertion_datetime`.

## Makefile

| Команда | Что делает |
|---|---|
| `make run` | `go run ./cmd/api` — локальный запуск (нужны доступные Postgres/Redis) |
| `make build` | сборка бинаря в `bin/vitalis` |
| `make docker-build` | сборка Docker-образа `vitalis` из `internal/deployments/Dockerfile` |
| `make docker-up` | `docker compose up --build -d` — поднять весь стек |
| `make docker-down` | остановить и удалить стек |
| `make docker-logs` | хвост логов контейнера `vitalis` |

## Разработка

Тесты (`go-sqlmock`, без реального Postgres):

```bash
go test ./...
```

Покрыты `internal/handlers` (auth, system-info эндпоинты) и частично `internal/services`. Пакеты `internal/database`, `internal/tasks`, `internal/enviroment` тестами пока не покрыты.
