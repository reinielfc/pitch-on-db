APP_NAME := pitch-on-db

# Docker Compose commands with environment variables
COMPOSE     := docker compose --env-file .env -f docker-compose.yml
COMPOSE_DEV := $(COMPOSE) -f docker-compose.dev.yml

-include .env
export

# Go binaries installation path
gobin       := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
SQLC_BIN	?= $(gobin)/sqlc
GOOSE_BIN	?= $(gobin)/goose
MOCKERY_BIN ?= $(gobin)/mockery

# Configuration Paths
SQLC_CONFIG := database/sqlc.yml

# Postgres connection parameters with defaults
POSTGRES_USER     ?= pitchondb
POSTGRES_PASSWORD ?= pitchondb
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_DB       ?= pitchondb
POSTGRES_SSLMODE  ?= disable

pg_creds := $(POSTGRES_USER):$(POSTGRES_PASSWORD)
pg_addr  := $(POSTGRES_HOST):$(POSTGRES_PORT)
pg_path	 := /$(POSTGRES_DB)
pg_query := ?sslmode=$(POSTGRES_SSLMODE)

POSTGRES_URL := postgres://$(pg_creds)@$(pg_addr)$(pg_path)$(pg_query)

# Formatting
c_reset  := $(shell tput sgr0)
c_bold   := $(shell tput bold)
c_bwhite := $(c_bold)$(shell tput setaf 7)
c_green  := $(shell tput setaf 2)

fmt_info = $(c_bwhite)%s:$(c_reset) %s...\n
fmt_done = $(c_bwhite)%s:$(c_reset) %s $(c_green)(done)$(c_reset)\n

.PHONY: \
	setup \
	generate generate-sqlc generate-mocks \
	test build \
	migrate dev dev-logs dev-stop dev-down

setup:
	@printf "$(fmt_info)" "setup" "Setting up the development environment"
	
	@printf "$(fmt_info)" "setup" "Installing sqlc for SQL code generation"
	@go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	
	@printf "$(fmt_info)" "setup" "Installing goose for database migrations"
	@go install -v github.com/pressly/goose/v3/cmd/goose@latest

	@printf "$(fmt_info)" "setup" "Installing mockery for generating mocks"
	@go install -v github.com/vektra/mockery/v2@latest

	@printf "$(fmt_info)" "setup" "Installing pnpm for frontend package management"
	@command -v pnpm >/dev/null 2>&1 || (curl -fsSL https://get.pnpm.io/install.sh | sh -)

	@printf "$(fmt_done)" "setup" "Installed sqlc, goose, mockery, and pnpm successfully"

generate: generate-sqlc generate-mocks
	@printf "$(fmt_done)" "generate" "All code generation tasks completed successfully"

generate-mocks:
	@printf "$(fmt_info)" "generate" "Generating mocks for API services using Mockery"
	@cd ./apps/api && $(MOCKERY_BIN)

generate-sqlc:
	@printf "$(fmt_info)" "generate" "Generating SQL code using SQLC with configuration from $(SQLC_CONFIG)"
	@$(SQLC_BIN) generate --file $(SQLC_CONFIG)

test:
	@printf "$(fmt_info)" "test" "Running unit tests for API services"
	@go test -v ./apps/api/services

build:
	@printf "$(fmt_info)" "build" "Building the '$(APP_NAME)' binary for API"
	@mkdir -vp bin
	@go build -o bin/$(APP_NAME) apps/api/main.go
	@printf "$(fmt_done)" "build" "Built '$(APP_NAME)' binary at 'bin/$(APP_NAME)'"

migrate:
	@printf "$(fmt_info)" "migrate" "Applying database migrations using Goose"
	@$(GOOSE_BIN) -dir database/migrations postgres "$(POSTGRES_URL)" up

start-db:
	@printf "$(fmt_info)" "start-db" "Starting the PostgreSQL database service using Docker Compose"
	@$(COMPOSE_DEV) up -d postgres
	@printf "$(fmt_done)" "start-db" "PostgreSQL database service started successfully"

start-api:
	@printf "$(fmt_info)" "start-api" "Starting the API service using Docker Compose"
	@$(COMPOSE_DEV) up -d api
	@printf "$(fmt_done)" "start-api" "API service started successfully"

dev:
	@$(COMPOSE_DEV) watch

dev-logs:
	@$(COMPOSE_DEV) logs -f

dev-stop:
	@$(COMPOSE_DEV) stop

dev-down:
	@$(COMPOSE_DEV) down
