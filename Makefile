.PHONY: run build

run:
	go run ./cmd/api

build:
	go build -o bin/vitalis ./cmd/api