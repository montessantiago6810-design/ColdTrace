# ColdTrace developer commands. Run `make help` for the list.
#
# Recipes load .env (if present) so the Go binaries and the SvelteKit dev
# server see the same COLDTRACE_* variables as Docker Compose.

SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help

GO            ?= go
NPM           ?= npm
DOCKER        ?= docker
GOLANGCI_LINT ?= golangci-lint
SQLC          ?= sqlc

LOAD_ENV := set -a; if [ -f .env ]; then . ./.env; fi; set +a;

.PHONY: help
help: ## List the available targets
	@grep -hE '^[a-z][a-z-]*:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-18s %s\n", $$1, $$2}'

## ---- Local stack ---------------------------------------------------------

.PHONY: dev
dev: infra migrate run ## Start Compose, migrate, then run the API and dashboard

.env:
	@echo "No .env found: copy .env.example to .env and set POSTGRES_PASSWORD." >&2
	@exit 1

.PHONY: infra
infra: .env ## Start TimescaleDB and Mosquitto and wait until healthy
	$(DOCKER) compose up -d --wait

.PHONY: infra-down
infra-down: ## Stop the Compose stack (data volumes are kept)
	$(DOCKER) compose down

.PHONY: migrate
migrate: ## Apply pending database migrations
	@$(LOAD_ENV) $(GO) run ./cmd/migrate up

.PHONY: migrate-status
migrate-status: ## Show which migrations are applied
	@$(LOAD_ENV) $(GO) run ./cmd/migrate status

.PHONY: run
run: web-deps ## Run the API and the dashboard dev server (Ctrl-C stops both)
	@$(LOAD_ENV) \
	$(GO) build -o bin/coldtrace-api ./cmd/api; \
	trap 'kill $$(jobs -p) 2>/dev/null || true' EXIT INT TERM; \
	bin/coldtrace-api & \
	(cd web && exec node_modules/.bin/vite dev) & \
	wait -n

.PHONY: api
api: ## Run only the Go API
	@$(LOAD_ENV) $(GO) run ./cmd/api

.PHONY: web
web: web-deps ## Run only the SvelteKit dev server
	@$(LOAD_ENV) cd web && $(NPM) run dev

## ---- Checks (the same ones CI runs) --------------------------------------

.PHONY: check
check: vet lint test sqlc-check web-check web-test ## Run every check except integration tests

.PHONY: vet
vet: ## go vet, including integration-tagged files
	$(GO) vet ./...
	$(GO) vet -tags integration ./...

.PHONY: lint
lint: ## golangci-lint (config in .golangci.yml)
	$(GOLANGCI_LINT) run ./...

.PHONY: fmt
fmt: ## Format Go code with gofmt and goimports
	$(GOLANGCI_LINT) fmt ./...

.PHONY: test
test: ## Go unit tests
	$(GO) test ./...

.PHONY: test-race
test-race: ## Go unit tests with the race detector (needs cgo)
	CGO_ENABLED=1 $(GO) test -race ./...

.PHONY: test-integration
test-integration: ## Store tests on TimescaleDB (Docker, or COLDTRACE_TEST_DATABASE_URL)
	$(GO) test -tags integration -count=1 ./internal/store/...

.PHONY: sqlc
sqlc: ## Regenerate internal/store/db from the SQL queries
	$(SQLC) generate

.PHONY: sqlc-check
sqlc-check: ## Fail if internal/store/db is out of date
	$(SQLC) diff

## ---- Dashboard -----------------------------------------------------------

.PHONY: web-deps
web-deps:
	@if [ ! -d web/node_modules ]; then cd web && $(NPM) ci; fi

.PHONY: web-check
web-check: web-deps ## svelte-check with warnings as errors
	cd web && $(NPM) run check

.PHONY: web-test
web-test: web-deps ## Vitest
	cd web && $(NPM) test

.PHONY: web-build
web-build: web-deps ## Production build of the dashboard
	cd web && $(NPM) run build
