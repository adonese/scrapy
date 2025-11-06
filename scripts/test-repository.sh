#!/bin/bash

# Script to run repository integration tests
# Requires a running PostgreSQL database with TimescaleDB

set -e

echo "Repository Integration Test Runner"
echo "==================================="
echo ""

# Check if database is running
echo "Checking database connection..."
if ! docker ps | grep -q postgres; then
    echo "Error: PostgreSQL container is not running"
    echo "Please run: make db-up"
    exit 1
fi

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 2

# Check if migrations have been run
echo "Checking if migrations are applied..."
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-cost_of_living}

export PGPASSWORD=$DB_PASSWORD
TABLE_EXISTS=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'cost_data_points');" 2>/dev/null || echo "f")

if [[ "$TABLE_EXISTS" =~ "f" ]]; then
    echo "Error: cost_data_points table does not exist"
    echo "Please run: make migrate"
    exit 1
fi

echo "Database is ready!"
echo ""

# Run repository tests
echo "Running repository tests..."
echo ""
go test -v ./internal/repository/postgres/... -count=1

echo ""
echo "==================================="
echo "Tests completed!"
