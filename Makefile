# Starterpack — single entrypoint for the whole monorepo.
# JS/TS workspaces are driven through bun + turbo; the Go backend through the go
# toolchain. Run `make help` for the list of targets.

SHELL := /bin/bash
API_DIR := apps/api

# Prefer a project-local DATABASE_URL from apps/api/.env.local when present.
-include $(API_DIR)/.env.local
export

.DEFAULT_GOAL := help

## ──────────────────────────────────────────────────────────────────────────
## Help
## ──────────────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ──────────────────────────────────────────────────────────────────────────
## Setup
## ──────────────────────────────────────────────────────────────────────────
.PHONY: setup
setup: install tools go-deps generate ## Install everything and generate code
	@echo "✓ setup complete. Configure apps/*/.env.local then run 'make dev'."

.PHONY: install
install: ## Install JS dependencies (bun)
	bun install

.PHONY: go-deps
go-deps: ## Download Go module dependencies
	cd $(API_DIR) && go mod download

.PHONY: tools
tools: ## Install Go CLIs (sqlc, dbmate, swag)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/amacneil/dbmate/v2@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## ──────────────────────────────────────────────────────────────────────────
## Code generation
## ──────────────────────────────────────────────────────────────────────────
.PHONY: generate
generate: sqlc openapi gen-client ## Run all code generators

.PHONY: sqlc
sqlc: ## Generate type-safe Go from SQL (sqlc)
	cd $(API_DIR) && sqlc generate

.PHONY: openapi
openapi: ## Generate the OpenAPI spec from Go annotations (swag)
	cd $(API_DIR) && swag init -g cmd/api/main.go -o docs --parseInternal

.PHONY: gen-client
gen-client: ## Generate the typed TS API client from the OpenAPI spec
	cd packages/api-client && bun run generate

## ──────────────────────────────────────────────────────────────────────────
## Database (dbmate)
## ──────────────────────────────────────────────────────────────────────────
DBMATE := dbmate --migrations-dir $(API_DIR)/db/migrations --schema-file $(API_DIR)/db/schema.sql

.PHONY: migrate
migrate: ## Apply pending migrations (dbmate up)
	$(DBMATE) up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	$(DBMATE) down

.PHONY: migrate-new
migrate-new: ## Create a migration: make migrate-new name=create_widgets
	$(DBMATE) new $(name)

## ──────────────────────────────────────────────────────────────────────────
## Development
## ──────────────────────────────────────────────────────────────────────────
.PHONY: dev
dev: ## Run everything (Go API + app/web/storybook/email) concurrently
	@$(MAKE) -j2 --no-print-directory dev-js dev-api

.PHONY: dev-js
dev-js: ## Run all JS apps via turbo (app:3000 web:3001 storybook:6006 email:3003)
	bun run dev

.PHONY: dev-api
dev-api: ## Run the Go API with live env (port 3002)
	cd $(API_DIR) && go run ./cmd/api

## ──────────────────────────────────────────────────────────────────────────
## Build / quality
## ──────────────────────────────────────────────────────────────────────────
.PHONY: build
build: build-js build-api ## Build everything

.PHONY: build-js
build-js: ## Build all JS apps/packages (turbo)
	bun run build

.PHONY: build-api
build-api: ## Compile the Go API binary
	cd $(API_DIR) && go build -o bin/api ./cmd/api

.PHONY: lint
lint: ## Lint JS (ultracite) and vet Go
	bun run check
	cd $(API_DIR) && go vet ./...

.PHONY: test
test: ## Run JS and Go tests
	bun run test
	cd $(API_DIR) && go test ./...

.PHONY: clean
clean: ## Remove build artifacts and node_modules
	bun run clean || true
	rm -rf $(API_DIR)/bin
