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
tools: ## Install Go CLIs (sqlc, swag)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
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
## Local dependencies (Docker Compose)
## ──────────────────────────────────────────────────────────────────────────
.PHONY: deps-up
deps-up: ## Start core local services (postgres)
	docker compose up -d

.PHONY: deps-up-all
deps-up-all: ## Start all local services (postgres + redis + mailpit)
	docker compose --profile all up -d

.PHONY: deps-down
deps-down: ## Stop local services (keep data volumes)
	docker compose down

.PHONY: deps-reset
deps-reset: ## Stop local services and DELETE their data volumes
	docker compose down -v

.PHONY: deps-logs
deps-logs: ## Tail local service logs
	docker compose logs -f

## ──────────────────────────────────────────────────────────────────────────
## Database (Atlas)
## ──────────────────────────────────────────────────────────────────────────
ATLAS := npx @ariga/atlas@0.37.0

.PHONY: db-diff db-apply db-lint db-status db-reset migrate migrate-new
db-diff: ## Generate migration: make db-diff name=add_users
	cd $(API_DIR) && $(ATLAS) migrate diff $(name) --env local

db-apply: ## Apply pending migrations
	cd $(API_DIR) && $(ATLAS) migrate apply --env local

db-lint: ## Lint pending migrations
	cd $(API_DIR) && $(ATLAS) migrate lint --env local --latest=1

db-status: ## Show migration status
	cd $(API_DIR) && $(ATLAS) migrate status --env local

db-hash: ## Re-calculate migration directory hash (fixes checksum mismatch)
	cd $(API_DIR) && $(ATLAS) migrate hash --env local

db-reset: ## DESTRUCTIVE — drop volume and start over
	docker compose down -v
	docker compose up -d --wait
	cd $(API_DIR) && $(ATLAS) migrate apply --env local

migrate: db-apply ## Alias for db-apply
migrate-new: db-diff ## Alias for db-diff

## ──────────────────────────────────────────────────────────────────────────
## Development
## ──────────────────────────────────────────────────────────────────────────
.PHONY: dev
dev: ## Run everything (Go API + JS apps) concurrently (bootstrapping)
	@$(MAKE) -j2 --no-print-directory client server

.PHONY: client
client: ## Run all JS apps via turbo (app:3000 web:3001 storybook:6006 email:3003) in TUI
	bun run dev

.PHONY: server
server: ## Run the Go API with live env (port 3002) for clean log visibility
	cd $(API_DIR) && go run ./cmd/api

.PHONY: dev-js
dev-js: client ## Alias for client

.PHONY: dev-api
dev-api: server ## Alias for server

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

.PHONY: lint lint-fix
lint: ## Lint JS (ultracite) and vet Go
	bun run check
	cd $(API_DIR) && go vet ./...

lint-fix: ## Auto-fix JS/TS formatting and linting issues (ultracite)
	bun run fix

.PHONY: test
test: ## Run JS and Go tests
	bun run test
	cd $(API_DIR) && go test ./...

.PHONY: clean
clean: ## Remove build artifacts and node_modules
	bun run clean || true
	rm -rf $(API_DIR)/bin
