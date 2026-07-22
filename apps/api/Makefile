# ---- Config ------------------------------------------------------------------

APP_NAME    := pitchondb
BIN_DIR     := ./bin
CMD_PATH    := ./cmd/service
BINARY_PATH := $(BIN_DIR)/$(APP_NAME)

GO ?= go
GOFLAGS :=
LDFLAGS_STRIP := -s -w
LDFLAGS_VERSION := \
	-X main.Version=$(VERSION) \
	-X main.GitCommit=$(GIT_COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)

GOLANGCI_LINT_VERSION := v2.11.4

MOCKERY_CONFIG := ./.mockery.yaml
OPENAPI_SPEC   := ./openapi.yml

# --- Environment --------------------------------------------------------------

GOBIN := $(or $(shell $(GO) env GOBIN),$(shell $(GO) env GOPATH)/bin)
PATH  := $(PATH):$(GOBIN)

VERSION    := $(shell git describe --tags --always --dirty || echo dev)
GIT_COMMIT := $(shell git rev-parse --short HEAD || echo unknown)
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# --- Tooling ------------------------------------------------------------------

.PHONY: tools

tools: ## Install development tools
	curl -sSfL $(GOLANGCI_LINT_URL) | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)

GOLANGCI_LINT_URL := https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh

# ---- Build / Run -------------------------------------------------------------

.PHONY: build run dev healthcheck

build: ## Compile service binary
	$(GO) build $(BUILD_FLAGS) -o $(BINARY_PATH) $(CMD_PATH)

run: ## Run service locally
	$(GO) run $(CMD_PATH)

dev: export SERVICE_APP_PORT := 8079
dev: ## Run service with live reload
	$(GO) tool air $(DEV_AIRFLAGS)

healthcheck: ## Perform health check on service
	$(GO) run $(CMD_PATH) healthcheck -r 3 -i 2s

BUILD_FLAGS := $(GOFLAGS) -ldflags "$(LDFLAGS_STRIP) $(LDFLAGS_VERSION)"

DEV_AIRFLAGS := \
	--build.cmd '$(GO) build $(BUILD_FLAGS) -o ./tmp/main $(CMD_PATH)' \
	--build.include_ext go \
	\
	--misc.startup_banner '' \
	--misc.clean_on_exit true \
	\
	--proxy.enabled true \
	--proxy.proxy_port 8080 \
	--proxy.app_port 8079

# ---- Test / Lint -------------------------------------------------------------

.PHONY: test test-cover lint fmt vet

test: ## Run unit tests
	$(GO) test -v ./... -race -count=1

test-cover: ## Run unit tests with coverage report
	$(GO) test -v ./... -race -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

lint: ## Run linter (requires: golangci-lint)
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	$(GO) tool goimports -w .

vet: ## Vet code
	$(GO) vet ./...

# ---- Codegen -----------------------------------------------------------------

.PHONY: generate $(addprefix generate-,enums mocks spec)

generate: enums mocks ## Generate code (enums, mocks)

enums: generate-enums ## Generate code for enums
generate-enums:
	$(GO) generate -x -run="github.com/dmarkham/enumer@latest" ./...

mocks: generate-mocks ## Generate mocks (requires: mockery)
generate-mocks:
	$(GO) tool mockery --config=$(MOCKERY_CONFIG)

spec: generate-spec ## Generate OpenAPI spec
generate-spec:
	$(GO) run $(CMD_PATH) openapi -o $(OPENAPI_SPEC)

# ---- Docker ------------------------------------------------------------------

.PHONY: docker-build

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) .

# ---- Housekeeping ------------------------------------------------------------

.PHONY: sync clean help

sync: ## Synchronize dependencies
	$(GO) mod tidy

clean: ## Clean up build artifacts
	rm -rvf tmp; rm -vf $(BINARY_PATH) coverage.out

help: ## Show this help message
	@echo -e "Usage: make [target]\n"
	@grep -hE '^[a-zA-Z_.-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\t%-16s %s\n", $$1, $$2}'
	@echo
