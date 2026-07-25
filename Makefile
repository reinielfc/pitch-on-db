## Config

ENV_FILE := .env

ifneq ("$(wildcard $(ENV_FILE))","")
	include $(ENV_FILE)
	export
endif

GO := go

MIGRATIONS   := ./db/migrations
SQLC_CONFIG  := ./db/sqlc.yml

POSTGRES_USER     ?= pitchondb
POSTGRES_PASSWORD ?= pitchondb
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_DB       ?= pitchondb
POSTGRES_SSLMODE  ?= disable

##@ Docker

DOCKER_COMMANDS := up stop down logs ps

.PHONY: $(DOCKER_COMMANDS) $(addprefix docker-,$(DOCKER_COMMANDS))

up: docker-up ## Start Docker services
docker-up:
	$(DOCKER_COMPOSE) up -d

stop: docker-stop ## Stop Docker services
docker-stop:
	$(DOCKER_COMPOSE) stop

down: docker-down ## Stop and remove Docker services
docker-down:
	$(DOCKER_COMPOSE) down

logs: docker-logs ## Follow Docker logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

ps: docker-ps ## List Docker services
docker-ps:
	$(DOCKER_COMPOSE) ps

##@ Services

DOCKER_SERVICES := web api postgres grafana

.PHONY: $(DOCKER_SERVICES) \
	$(addprefix start-,$(DOCKER_SERVICES)) $(addprefix service-,$(addsuffix -start,$(DOCKER_SERVICES))) \
	$(addprefix follow-,api postgres)      $(addprefix service-,$(addsuffix -follow,api postgres)) \
	$(addprefix reset-,postgres)           $(addprefix service-,$(addsuffix -reset,postgres))

start-web: service-web-start ## Start Web service
service-web-start:
	$(DOCKER_COMPOSE) build web
	$(DOCKER_COMPOSE) up web

start-api: service-api-start ## Start API service
service-api-start:
	$(DOCKER_COMPOSE) build api
	$(DOCKER_COMPOSE) up api

start-postgres: service-postgres-start ## Start Postgres service
service-postgres-start:
	$(DOCKER_COMPOSE) up postgres

start-grafana: service-grafana-start ## Start Grafana service
service-grafana-start:
	$(DOCKER_COMPOSE) up grafana

follow-api: ## Follow API service logs
service-api-follow:
	$(DOCKER_COMPOSE) logs -f api

follow-postgres: ## Follow Postgres service logs
service-postgres-follow:
	$(DOCKER_COMPOSE) logs -f postgres

reset-postgres: ## Reset Postgres service (stop, remove, and start)
service-postgres-reset:
	$(DOCKER_COMPOSE) down -v postgres
	$(DOCKER_COMPOSE) up -d postgres

DOCKER_COMPOSE := $(or $(shell command -v docker-compose),docker compose) \
	-f docker-compose.yml -f docker-compose.dev.yml

##@ Code Generation

.PHONY: sqlc $(addprefix generate-,sqlc)

sqlc: generate-sqlc ## Generate SQL code (requires: sqlc)
generate-sqlc:
	$(GO) tool sqlc generate -f $(SQLC_CONFIG)

schema: generate-schema ## Generate OpenAPI spec and Schema
generate-schema:
	$(MAKE) -C apps/api generate-spec
	$(MAKE) -C apps/web generate-schema

##@ Postgres

POSTGRES_URL := host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) \
	user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) \
	dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSLMODE)

GOOSE_FLAGS := -dir $(MIGRATIONS) postgres '$(POSTGRES_URL)'

.PHONY: migrate rollback $(addprefix goose-,up down)

migrate: goose-up ## Apply database migrations
goose-up:
	$(GO) tool goose $(GOOSE_FLAGS) up

rollback: goose-down ## Rollback database migrations
goose-down:
	$(GO) tool goose $(GOOSE_FLAGS) down

##@ Development Tools

.PHONY: tools

tools: ## Install required tools
	cd tools && $(GO) install tool

##@ Housekeeping

.PHONY: help

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z_.-]+:.*?## / { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
		/^##@ / { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)
	@echo
