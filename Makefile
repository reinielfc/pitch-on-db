APP_NAME := pitch-on-db

COMPOSE  := docker compose --env-file .env -f docker-compose.yml -f .docker/monitoring.docker-compose.yml

gobin    := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
SQLC	 := $(gobin)/sqlc
GOOSE	 := $(gobin)/goose
MOCKERY  := $(gobin)/mockery
MOCKERY_CONFIG ?= .mockery.yaml

MOCK_NAME ?=

-include .env
export

POSTGRES_USER ?= pitchondb
POSTGRES_PASSWORD ?= pitchondb
POSTGRES_PORT ?= 5432
POSTGRES_DB ?= pitchondb
POSTGRES_SSLMODE ?= disable

DATABASE_URL ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)"

.PHONY: setup sqlc-generate mock-generate build migrate run docker-up docker-down clean

setup:
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/vektra/mockery/v2@latest
	@echo "Installed sqlc, goose, and mockery binaries"

mock-generate:
	@echo "Generating all configured mocks using $(MOCKERY_CONFIG)"
	@$(MOCKERY) --config "$(MOCKERY_CONFIG)"
	@echo "Generated all configured mocks using $(MOCKERY_CONFIG)"

sqlc-generate:
	@echo "Generating SQL code using sqlc with configuration from database/sqlc.yml"
	@$(SQLC) generate --file database/sqlc.yml
	@echo "Generated SQL code successfully"


test:
	@go test -v ./internal/services

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
