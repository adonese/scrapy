.PHONY: run build templ test test-unit test-repo test-ci test-integration test-coverage test-bench validate-scrapers lint security-scan clean db-up db-down db-logs migrate migrate-down migrate-version temporal-up temporal-down temporal-ui worker run-workflow trigger-scrape trigger-scheduled prom-up prom-down prom-ui scrape-bayut scrape-all e2e-test ci-setup ci-validate css css-build css-watch install-tailwind dev

TEMPL_VERSION ?= v0.3.960

templ:
	go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate

# CSS/Frontend commands
install-tailwind: ## Download and install Tailwind CSS standalone CLI
	@echo "Installing Tailwind CSS standalone CLI..."
	@if [ "$$(uname -s)" = "Linux" ]; then \
		if [ "$$(uname -m)" = "x86_64" ]; then \
			curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64; \
			chmod +x tailwindcss-linux-x64; \
			sudo mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss; \
		elif [ "$$(uname -m)" = "aarch64" ]; then \
			curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64; \
			chmod +x tailwindcss-linux-arm64; \
			sudo mv tailwindcss-linux-arm64 /usr/local/bin/tailwindcss; \
		else \
			echo "Unsupported architecture: $$(uname -m)"; exit 1; \
		fi \
	elif [ "$$(uname -s)" = "Darwin" ]; then \
		if [ "$$(uname -m)" = "arm64" ]; then \
			curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64; \
			chmod +x tailwindcss-macos-arm64; \
			sudo mv tailwindcss-macos-arm64 /usr/local/bin/tailwindcss; \
		else \
			curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64; \
			chmod +x tailwindcss-macos-x64; \
			sudo mv tailwindcss-macos-x64 /usr/local/bin/tailwindcss; \
		fi \
	else \
		echo "Unsupported OS. Please download manually from:"; \
		echo "https://github.com/tailwindlabs/tailwindcss/releases/latest"; \
		exit 1; \
	fi
	@echo "✓ Tailwind CSS CLI installed successfully!"
	@tailwindcss --help > /dev/null && echo "✓ Verification successful"

css: css-build ## Build CSS (alias for css-build)

css-build: ## Build Tailwind CSS once
	@echo "Building Tailwind CSS..."
	@tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify
	@echo "✓ CSS built successfully"

css-watch: ## Build Tailwind CSS in watch mode for development
	@echo "Starting Tailwind CSS in watch mode..."
	@tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch

dev: ## Start development mode (run templ generate, css watch, and show instructions)
	@echo "🚀 Development environment setup:"
	@echo ""
	@echo "Terminal 1: make css-watch"
	@echo "Terminal 2: make run"
	@echo ""
	@echo "Starting CSS watch mode now..."
	@make css-watch

run: templ
	go run cmd/api/main.go

build: templ
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

# CI/CD targets
ci-setup: ## Set up CI environment
	@echo "Setting up CI environment..."
	@./scripts/ci/setup.sh

test-ci: ## Run full CI test suite
	@echo "Running CI test suite..."
	@./scripts/ci/run-tests.sh

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -race -timeout 10m -tags=integration ./...

test-coverage: ## Generate coverage report
	@echo "Generating coverage report..."
	@./scripts/ci/coverage.sh

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^$$ ./... | tee benchmarks.txt

validate-scrapers: ## Validate all scrapers
	@echo "Validating scrapers..."
	@./scripts/ci/validate.sh

lint: ## Run linters
	@echo "Running linters..."
	@go vet ./...
	@gofmt -l .
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed"

security-scan: ## Run security scan
	@echo "Running security scan..."
	@which gosec > /dev/null && gosec ./... || echo "gosec not installed, run: go install github.com/securego/gosec/v2/cmd/gosec@latest"

ci-validate: ## Run all CI validation checks
	@echo "Running CI validation..."
	@make lint
	@make security-scan
	@make test-ci
	@make test-coverage
	@make validate-scrapers

# Docker Compose for testing
test-env-up: ## Start test environment with docker-compose
	@echo "Starting test environment..."
	@docker-compose -f docker-compose.test.yml up -d postgres-test

test-env-down: ## Stop test environment
	@echo "Stopping test environment..."
	@docker-compose -f docker-compose.test.yml down -v

# Help target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
