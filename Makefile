# ---- Config ------------------------------------------------------------------

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

POSTGRES_URL := host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) \
	user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) \
	dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSLMODE)

DOCKER_COMPOSE := $(or $(shell command -v docker-compose),$(shell command -v docker) compose,docker compose)\
	-f docker-compose.yml -f docker-compose.dev.yml

# ---- Docker ------------------------------------------------------------------

DOCKER_TARGETS := up stop down logs ps

.PHONY: $(addprefix docker-,$(DOCKER_TARGETS)) $(DOCKER_TARGETS)

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

.PHONY: $(addprefix start-,web api postgres grafana)

start-web: ## Start Web service
	$(DOCKER_COMPOSE) build web
	$(DOCKER_COMPOSE) up web

start-api: ## Start API service
	$(DOCKER_COMPOSE) build api
	$(DOCKER_COMPOSE) up api

start-postgres: ## Start Postgres service
	$(DOCKER_COMPOSE) up postgres

start-grafana: ## Start Grafana service
	$(DOCKER_COMPOSE) up grafana

.PHONY: $(addprefix follow-,api postgres)

follow-api: ## Follow API service logs
	$(DOCKER_COMPOSE) logs -f api

follow-postgres: ## Follow Postgres service logs
	$(DOCKER_COMPOSE) logs -f postgres

.PHONY: reset-postgres

reset-postgres: ## Reset Postgres service (stop, remove, and start)
	$(DOCKER_COMPOSE) down -v postgres
	$(DOCKER_COMPOSE) up -d postgres

# ---- Codegen -----------------------------------------------------------------

.PHONY: sqlc $(addprefix generate-,sqlc)

sqlc: generate-sqlc ## Generate SQL code
generate-sqlc:
	$(GO) tool sqlc generate -f $(SQLC_CONFIG)

# ---- Postgres / Migrations ---------------------------------------------------

.PHONY: migrate rollback $(addprefix goose-,up down)

migrate: goose-up ## Apply database migrations
goose-up:
	$(GO) tool goose $(GOOSE_FLAGS) up

rollback: goose-down ## Rollback database migrations
goose-down:
	$(GO) tool goose $(GOOSE_FLAGS) down

GOOSE_FLAGS := -dir $(MIGRATIONS) postgres '$(POSTGRES_URL)'

# ---- API ---------------------------------------------------------------------

.PHONY: $(addprefix api.,build test run healthcheck clean)

api.build: ## (API) Build the application
	$(MAKE) -C ./apps/api build

api.test: ## (API) Run tests for the application
	$(MAKE) -C ./apps/api test

api.run: ## (API) Run the application
	$(MAKE) -C ./apps/api run

api.healthcheck: ## (API) Perform health check on the application
	$(MAKE) -C ./apps/api healthcheck

api.clean: ## (API) Clean build artifacts for the application
	$(MAKE) -C ./apps/api clean

# ---- Tools -------------------------------------------------------------------

.PHONY: tools

tools: ## Install required tools
	cd tools && $(GO) install tool

# ---- Housekeeping ------------------------------------------------------------

.PHONY: help

help: ## Show this help message
	@echo -e "Usage: make [target]\n"
	@grep -hE '^[a-zA-Z_.-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\t%-20s %s\n", $$1, $$2}'
	@echo
