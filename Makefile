APP_NAME := pitch-on-db

COMPOSE  := docker compose --env-file .env -f docker-compose.yml -f .docker/monitoring.docker-compose.yml

gobin    := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
SQLC	 := $(gobin)/sqlc
GOOSE	 := $(gobin)/goose

-include .env
export

DATABASE_URL ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)"

.PHONY: setup generate build migrate run docker-up docker-down clean

setup:
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Installed sqlc and goose binaries"

generate:
	@$(SQLC) generate --file database/sqlc.yml

build: clean
	@mkdir -vp bin
	@go build -o bin/$(APP_NAME) cmd/server/main.go
	@echo "Built $(APP_NAME) binary at bin/$(APP_NAME)"

migrate:
	@$(GOOSE) -dir database/migrations postgres "$(DATABASE_URL)" up

run: docker-up migrate

docker-up: 
	@$(COMPOSE) up -d --build

docker-down:
	@$(COMPOSE) down

clean:
	@rm -rvf bin
