.PHONY: build test test-cover test-integration test-e2e test-all ch-up ch-down docs-serve docs-build docs-deploy

build:
	go build -o drover-ch ./cmd/drover-ch

# Unit tests only (default; no Docker required)
test:
	go test ./... -count=1

test-cover:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

# Requires: docker compose -f docker-compose.test.yml up -d --wait
ch-up:
	docker compose -f docker-compose.test.yml up -d --wait

ch-down:
	docker compose -f docker-compose.test.yml down -v

test-integration: ch-up
	CLICKHOUSE_PORT=19000 go test -tags=integration ./... -count=1 -timeout=5m

test-e2e: ch-up
	CLICKHOUSE_PORT=19000 go test -tags=e2e ./test/e2e/... -count=1 -timeout=10m

test-all: test test-integration test-e2e

docs-serve:
	mkdocs serve

docs-build:
	mkdocs build --strict

docs-deploy:
	mkdocs gh-deploy --force
