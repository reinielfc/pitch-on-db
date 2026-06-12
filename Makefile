APP_NAME=pitch-on-db
DB_DSN=$(or $(DATABASE_URL),postgres://pitchondb:pitchondb@localhost:5432/pitchondb?sslmode=disable)

GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif
SQLC  := $(GOBIN)/sqlc
GOOSE := $(GOBIN)/goose

.PHONY: build run clean generate migrate docker-up docker-down

build: generate
	@mkdir -vp bin
	@go build -v -o bin/$(APP_NAME) cmd/server/main.go
	@echo "Build complete: bin/$(APP_NAME)"

run:
	@set -a && . ./.env && set +a && go run cmd/server/main.go

generate:
	@$(SQLC) generate --file database/sqlc.yml

migrate:
	@$(GOOSE) -dir migrations postgres "$(DB_DSN)" up

docker-up:
	@docker compose up -d

docker-down:
	@docker compose down

clean:
	@rm -vrf bin
