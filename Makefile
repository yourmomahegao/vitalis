.PHONY: run build docker-build docker-up docker-down docker-logs

run:
	go run ./cmd/api

build:
	go build -o bin/vitalis ./cmd/api

docker-build:
	docker build -f internal/deployments/Dockerfile -t vitalis .

docker-build-host:
	docker build -f internal/deployments/Dockerfile -t yourmomahegao/vitalis:latest .

docker-push-host:
	docker push yourmomahegao/vitalis:latest

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f vitalis