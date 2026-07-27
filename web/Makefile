## Config

APP_NAME := pitchondb
PNPM     ?= pnpm

OPENAPI_SPEC := ../api/openapi.yml

##@ Environment

VERSION := $(shell git describe --tags --always --dirty || echo dev)

.PHONY: tools install

tools: ## Install development tools
	curl -fsSL https://get.pnpm.io/install.sh | sh -

install: ## Install project dependencies
	$(PNPM) install

##@ Build / Run

.PHONY: dev build preview

dev: ## Run dev server preview
	$(PNPM) dev --strictPort

build: ## Production build
	$(PNPM) build

preview: ## Preview the production build locally
	$(PNPM) preview

##@ Test / Lint

.PHONY: test lint fmt check

test: ## Run tests
	$(PNPM) test

lint: ## Lint source
	$(PNPM) lint

format: ## Format with prettier
	$(PNPM) format

check: ## svelte-check + tsc type checking
	$(PNPM) check

##@ Housekeeping

.PHONY: help

clean: ## Clean up build artifacts
	rm -rf .svelte-kit build node_modules/.vite

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z_.-]+:.*?## / { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
		/^##@ / { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)
	@echo
