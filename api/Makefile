## Config

APP_NAME    := pitchondb
BIN_DIR     := ./bin
CMD_PATH    := ./cmd/service
BINARY_PATH := $(BIN_DIR)/$(APP_NAME)

GO ?= go
GOFLAGS :=
LDFLAGS_STRIP := -s -w
GOLANGCI_LINT_VERSION := v2.11.4

MOCKERY_CONFIG := ./.mockery.yaml
OPENAPI_SPEC   := ./openapi.yml

##@ Environment

GOBIN := $(or $(shell $(GO) env GOBIN),$(shell $(GO) env GOPATH)/bin)
PATH  := $(PATH):$(GOBIN)

LDFLAGS_VERSION := \
	-X main.version=$(shell git describe --tags --always --dirty || echo dev) \
	-X main.gitCommit=$(shell git rev-parse --short HEAD || echo unknown) \
	-X main.buildTime=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

GOLANGCI_LINT_URL := https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh

.PHONY: tools

tools: ## Install development tools
	curl -sSfL $(GOLANGCI_LINT_URL) | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)

##@ Build / Run

BUILD_FLAGS = $(GOFLAGS) -ldflags '$(LDFLAGS_STRIP) $(LDFLAGS_VERSION)'

.PHONY: build run dev healthcheck

build: ## Compile service binary
	$(GO) build $(BUILD_FLAGS) -o $(BINARY_PATH) $(CMD_PATH)

run: .env ## Run service locally
	$(GO) run $(BUILD_FLAGS) $(CMD_PATH)

dev: .env.development ## Run service with live reload
	$(GO) tool air -c .air.toml \
		--build.cmd="$(GO) build $(BUILD_FLAGS) \
		-o ./tmp/main $(CMD_PATH)"

doc: ## Run godoc server
	$(GO) run golang.org/x/pkgsite/cmd/pkgsite@latest -http=localhost:6060

healthcheck: ## Perform health check on service
	$(GO) run $(BUILD_FLAGS) $(CMD_PATH) healthcheck -r 3 -i 2s

##@ Quality Assurance

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

##@ Code Generation

.PHONY: generate $(addprefix generate-,enums mocks spec)

enums: generate-enums ## [generate] Code for enums
generate-enums:
	$(GO) generate -x -run="go tool enumer" ./...

mocks: generate-mocks ## [generate] Mocks (requires: mockery)
generate-mocks:
	$(GO) tool mockery --config=$(MOCKERY_CONFIG)

spec: generate-spec ## [generate] OpenAPI spec
generate-spec:
	$(GO) run $(BUILD_FLAGS) $(CMD_PATH) openapi -o $(OPENAPI_SPEC)

##@ Housekeeping

.PHONY: sync clean help

sync: ## Synchronize dependencies
	$(GO) mod tidy

clean: ## Clean up build artifacts
	rm -rvf tmp; rm -vf $(BINARY_PATH) coverage.out

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z_.-]+:.*?## / { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
		/^##@ / { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)
	@echo
