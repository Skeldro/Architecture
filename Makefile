# CollabDocs — Phase 1
# `make dev` and `make test` need no container runtime: scripts/localpg.sh runs
# a userland PostgreSQL, downloading it on first use.

GO ?= go
PG := ./scripts/localpg.sh

.PHONY: help dev test loadtest build fmt vet check db-start db-stop db-reset docker-build

help:
	@echo "dev        run the application against a local database (http://localhost:8080)"
	@echo "test       run the FR1-FR4 test suite against a separate test database"
	@echo "loadtest   100 virtual users x 1 write/sec for 60s against a running instance"
	@echo "check      fmt + vet + test"
	@echo "db-start   start the local database    db-stop / db-reset also available"

db-start:
	@$(PG) start >/dev/null

db-stop:
	@$(PG) stop

db-reset:
	@$(PG) reset

dev: db-start
	DATABASE_URL=$$($(PG) url) $(GO) run .

test: db-start
	TEST_DATABASE_URL=$$($(PG) testurl) $(GO) test -count=1 ./...

loadtest:
	$(GO) run ./cmd/loadtest

build:
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/collabdocs .

fmt:
	gofmt -w .

vet:
	$(GO) vet ./...

check: fmt vet test

docker-build:
	docker build -t collabdocs:local .
