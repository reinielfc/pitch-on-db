APP_NAME := pitch-on-db

-include .env
export

# Go binaries installation path
GOBIN := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
PATH  := $(PATH):$(GOBIN)

# Configuration Paths
SQLC_CONFIG := database/sqlc.yml
GOOSE_MIGRATIONS := database/migrations

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

.PHONY: \
	install-all \
		install-gotools \
			install-sqlc \
			install-mockery \
			install-goose \
		install-pnpm \
	generate \
		generate-sqlc \
		generate-mocks \
	migrate \
		rollback \
	docker-up up \
		start-db \
		start-api \
		start-web \
		start-monitoring \
	docker-down down \
	docker-stop stop \
	docker-ps ps
	

# === Logging ===

c_reset  := $(shell tput sgr0)
c_bold   := $(shell tput bold)
c_bwhite := $(c_bold)$(shell tput setaf 7)
c_red    := $(shell tput setaf 1)
c_green  := $(shell tput setaf 2)

fmt_info := $(c_bwhite)%s:$(c_reset) %s...\n
log_info = printf "$(fmt_info)" "$@" "$(1)"

fmt_done := $(c_bwhite)%s:$(c_reset) %s $(c_green)(done)$(c_reset)\n
log_done = printf "$(fmt_done)" "$@" "$(1)"

fmt_fail := $(c_bwhite)%s:$(c_reset) %s $(c_red)(failed)$(c_reset)\n
log_fail = printf "$(fmt_fail)" "$@" "$(1)"

# === Development Environment ===

ensure_file_exists = \
	if [ ! -f "$(1)" ]; then \
		$(call log_fail,File does not exist: $(1)); \
		exit 1; \
	fi

ensure_cmd_exists = \
	if ! command -v "$(1)" &> /dev/null; then \
		$(call log_fail,Command not found: $(1)); \
		exit 1; \
	fi

fmt_bin_not_installed := Binary '%s' not found, please run 'make install-%s' to install it
ensure_bin_installed = \
	if [ ! -f "$(1)" ]; then \
		$(call log_fail,$(shell printf "$(fmt_bin_not_installed)" "$(1)" "$(2)")); \
		exit 1; \
	fi

define install_go_tool
	@$(call ensure_cmd_exists,go)
	@go install -v $(2)
	@$(call log_done,Go tool '$(1)' installed successfully)
endef

install-all: install-gotools install-pnpm
	@$(call log_done,Setup completed successfully)

install-gotools: install-sqlc install-mockery install-goose
	@$(call log_done,All Go tools installed successfully)

install-sqlc:
	$(call install_go_tool,sqlc,github.com/sqlc-dev/sqlc/cmd/sqlc@latest)

install-mockery:
	$(call install_go_tool,mockery,github.com/vektra/mockery/v2@latest)

install-goose:
	$(call install_go_tool,goose,github.com/pressly/goose/v3/cmd/goose@latest)

install-pnpm:
	@$(call ensure_cmd_exists,curl)
	@curl -fsSL https://get.pnpm.io/install.sh | sh -
	@$(call log_done,pnpm installed successfully)

# === Lifecycle Commands ===

sync:
	@go work sync
	@cd apps/api && go mod tidy
	@$(call log_done,Dependencies synced successfully)

test:
	@go test -v ./apps/api/services
	@$(call log_done,All tests passed successfully)

build:
	@go build -o bin/$(APP_NAME) ./apps/api
	@$(call log_done,Application built successfully)

# === Code Generation ===

generate: generate-sqlc generate-mocks
	@$(call log_done,All code generation tasks completed successfully)

generate-sqlc:
	@$(call ensure_cmd_exists,sqlc)
	@$(call ensure_file_exists,$(SQLC_CONFIG))
	@sqlc generate --file $(SQLC_CONFIG)
	@$(call log_done,SQL code generated successfully)

generate-mocks:
	@$(call ensure_cmd_exists,mockery)
	@cd ./apps/api && mockery
	@$(call log_done,Mocks generated successfully)

# === Database Migration ===

goose_migrate := goose -dir $(GOOSE_MIGRATIONS) postgres "$(POSTGRES_URL)"

migrate:
	@$(call ensure_cmd_exists,goose)
	@$(goose_migrate) up
	@$(call log_done,Database migrations applied successfully)

rollback:
	@$(call ensure_cmd_exists,goose)
	@$(goose_migrate) down
	@$(call log_done,Database migration rolled back successfully)

# === Docker ===

compose 	:= docker-compose -f docker-compose.yml
compose_dev := $(compose) -f docker-compose.dev.yml

docker-up up:
	@$(compose_dev) build api web
	@$(compose_dev) up -d
	@$(call log_done,All services started successfully)

start-db:
	@$(compose_dev) up -d postgres
	@$(call log_done,Postgres database started successfully)

start-api:
	@$(compose_dev) build api
	@$(compose_dev) up -d api
	@$(call log_done,API service started successfully)

start-web:
	@$(compose_dev) build web
	@$(compose_dev) up -d web
	@$(call log_done,Web service started successfully)

start-monitoring:
	@$(compose_dev) up -d grafana loki
	@$(call log_done,Monitoring services started successfully)

docker-down down:
	@$(compose_dev) down
	@$(call log_done,All services stopped and removed successfully)

docker-stop stop:
	@$(compose_dev) stop
	@$(call log_done,All services stopped successfully)

docker-clean:
	@$(compose_dev) down --volumes
	@$(call log_done,All services stopped and cleaned up successfully)

docker-purge:
	@$(compose_dev) down --rmi all --volumes --remove-orphans
	@$(call log_done,All services purged successfully)

docker-ps ps:
	@$(call log_info,Listing all running services)
	@$(compose_dev) ps
