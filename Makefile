.PHONY: run build test clean db-up db-down db-logs migrate migrate-down migrate-version

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

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
