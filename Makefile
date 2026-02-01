DEFAULT_GOAL := build

include .env
export $(shell sed 's/=.*//' .env)

.PHONY: .build, fmt, clean, migrate, up, down, test, test-integration, cover

build:
	go build -o ./bin/app ./cmd/app/main.go

clean:
	rm -rf ./bin

fmt:
	go fmt ./...

up:
	docker compose up -d

down:
	docker compose down

migrate:
	goose -dir ./migrations postgres postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@127.0.0.1:$(POSTGRES_PORT)/$(POSTGRES_DB) up

test:
	go test ./...

test-integration:
	go test ./internal/integration -run TestIntegration -v

cover:
	go tool cover -html=coverage.out

