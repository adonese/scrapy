.PHONY: run build test test-unit test-repo clean db-up db-down db-logs migrate migrate-down migrate-version temporal-up temporal-down temporal-ui worker run-workflow trigger-scrape trigger-scheduled prom-up prom-down prom-ui scrape-bayut scrape-all e2e-test

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

test-unit:
	go test -v ./pkg/... ./internal/models/...

test-repo:
	./scripts/test-repository.sh

clean:
	rm -rf bin/

# Database commands
db-up:
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 3

db-down:
	docker-compose down

db-logs:
	docker-compose logs -f postgres

# Migration commands
migrate:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-version:
	go run cmd/migrate/main.go version

# Temporal commands
temporal-up:
	@echo "Starting Temporal..."
	docker-compose up -d temporal temporal-ui
	@echo "Waiting for Temporal to be ready..."
	@sleep 5

temporal-down:
	docker-compose stop temporal temporal-ui

temporal-ui:
	@echo "Temporal UI: http://localhost:8233"
	@which open > /dev/null && open http://localhost:8233 || which xdg-open > /dev/null && xdg-open http://localhost:8233 || echo "Open http://localhost:8233 in your browser"

worker:
	go run cmd/worker/main.go

run-workflow:
	go run examples/workflow_client.go

# Workflow trigger commands
trigger-scrape:
	go run cmd/trigger-scrape/main.go -scraper bayut

trigger-scheduled:
	go run cmd/trigger-scrape/main.go -scheduled

# Observability commands
prom-up:
	docker-compose up -d prometheus
	@echo "Prometheus UI: http://localhost:9090"

prom-down:
	docker-compose stop prometheus

prom-ui:
	@which open > /dev/null && open http://localhost:9090 || which xdg-open > /dev/null && xdg-open http://localhost:9090 || echo "Open http://localhost:9090 in your browser"

# Scraper commands
scrape-bayut:
	go run cmd/scraper/main.go -scraper bayut

scrape-all:
	go run cmd/scraper/main.go -scraper all

# Setup for development
setup: db-up migrate
	@echo "Development environment ready!"

# Complete end-to-end test
e2e-test:
	@echo "Starting end-to-end test..."
	@echo "1. Starting Temporal..."
	@make temporal-up
	@echo "2. Starting database..."
	@make db-up
	@sleep 3
	@echo "3. Running migrations..."
	@make migrate
	@echo "4. Starting worker in background..."
	@go run cmd/worker/main.go > /tmp/worker.log 2>&1 &
	@echo $$! > /tmp/worker.pid
	@sleep 3
	@echo "5. Triggering scraper workflow..."
	@go run cmd/trigger-scrape/main.go -scraper bayut || true
	@echo "6. Cleaning up worker..."
	@kill `cat /tmp/worker.pid` 2>/dev/null || true
	@rm -f /tmp/worker.pid
	@echo "E2E test complete!"
