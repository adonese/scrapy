# Testing Guide - Iteration 1.2

## Quick Start

```bash
# 1. Start database
make db-up

# 2. Run migrations
make migrate

# 3. Run tests
make test

# 4. Start API server
make run

# 5. Test health endpoint (in another terminal)
curl localhost:8080/health
# Expected: {"status":"ok","database":"connected","timestamp":"..."}
```

## Detailed Testing

### Database Setup

```bash
# Start PostgreSQL with TimescaleDB
make db-up

# Verify container is healthy
docker ps | grep cost-of-living-db

# Check database logs
make db-logs

# Connect to database directly
docker exec -it cost-of-living-db psql -U postgres -d cost_of_living
```

### Migrations

```bash
# Run migrations
make migrate

# Check migration status
make migrate-version
# Expected: Version: 1, Dirty: false

# Rollback migration (for testing)
make migrate-down

# Re-apply migration
make migrate
```

### Database Schema Verification

```bash
# View all tables
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "\dt"

# Describe cost_data_points table
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "\d cost_data_points"

# View hypertables (TimescaleDB)
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "SELECT * FROM timescaledb_information.hypertables;"

# Check indexes
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "SELECT indexname FROM pg_indexes WHERE tablename = 'cost_data_points';"
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./pkg/database
go test -v ./internal/handlers

# Run tests in short mode (skip integration tests)
go test -short ./...

# Run integration tests only
go test -v -run Integration ./pkg/database
```

### API Testing

```bash
# Start server
make run

# Test health endpoint
curl localhost:8080/health

# Test with pretty JSON (if jq installed)
curl -s localhost:8080/health | jq

# Test with verbose output
curl -v localhost:8080/health
```

### Manual Data Testing

```bash
# Insert test data
docker exec cost-of-living-db psql -U postgres -d cost_of_living << 'SQL'
INSERT INTO cost_data_points (
    category, item_name, price, location, source
) VALUES (
    'Food', 'Milk (1L)', 6.50, 
    '{"emirate": "Dubai", "city": "Dubai", "area": "JLT"}'::jsonb,
    'manual_test'
);
SQL

# Query test data
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "SELECT id, item_name, price, location->>'emirate' as emirate FROM cost_data_points;"

# Clean up test data
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "DELETE FROM cost_data_points WHERE source = 'manual_test';"
```

## Troubleshooting

### Database won't start

```bash
# Check if port 5432 is already in use
lsof -i:5432

# Stop any existing PostgreSQL
brew services stop postgresql  # macOS
sudo systemctl stop postgresql  # Linux

# Remove old container and try again
docker rm -f cost-of-living-db
make db-up
```

### Migration fails with "dirty database"

```bash
# Check migration status
make migrate-version

# Force migration to version 0
go run cmd/migrate/main.go force 0

# Drop tables and start fresh
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "DROP TABLE IF EXISTS cost_data_points CASCADE; DROP TABLE IF EXISTS schema_migrations CASCADE;"

# Re-run migrations
make migrate
```

### API server won't start (port in use)

```bash
# Find process using port 8080
lsof -i:8080

# Kill the process
kill -9 <PID>

# Or kill all go processes
pkill -9 go

# Start server again
make run
```

### Tests fail

```bash
# Ensure database is running
docker ps | grep cost-of-living-db

# Ensure migrations are applied
make migrate

# Run tests with verbose output to see error
go test -v ./...

# Check if it's a connection issue
go test -v ./pkg/database
```

## Expected Test Results

### All Tests Pass
```
?   	github.com/adonese/cost-of-living/cmd/api	[no test files]
?   	github.com/adonese/cost-of-living/cmd/migrate	[no test files]
=== RUN   TestHealth
--- PASS: TestHealth (0.00s)
PASS
ok  	github.com/adonese/cost-of-living/internal/handlers	0.003s
=== RUN   TestNewConfigFromEnv
--- PASS: TestNewConfigFromEnv (0.00s)
=== RUN   TestNewConfigFromEnvWithCustomValues
--- PASS: TestNewConfigFromEnvWithCustomValues (0.00s)
=== RUN   TestConnectIntegration
--- PASS: TestConnectIntegration (0.01s)
=== RUN   TestDatabaseNilConnection
--- PASS: TestDatabaseNilConnection (0.00s)
PASS
ok  	github.com/adonese/cost-of-living/pkg/database	0.013s
```

### Health Endpoint Response
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2025-11-06T19:33:27+04:00"
}
```

### Database Schema
```
Table "public.cost_data_points"
    Column    |           Type           | Nullable |         Default          
--------------+--------------------------+----------+--------------------------
 id           | uuid                     | not null | uuid_generate_v4()
 category     | character varying(100)   | not null | 
 ...
Indexes:
    "cost_data_points_pkey" PRIMARY KEY, btree (id, recorded_at)
    "idx_cost_data_points_attributes" gin (attributes)
    "idx_cost_data_points_category" btree (category)
    ...
```

## Clean Up

```bash
# Stop API server
# Press Ctrl+C or:
pkill -f "go run cmd/api/main.go"

# Stop database
make db-down

# Remove volumes (WARNING: deletes all data)
docker-compose down -v
```

## Performance Testing

```bash
# Simple load test with curl
for i in {1..100}; do
    curl -s localhost:8080/health > /dev/null
done

# With timing
time for i in {1..100}; do
    curl -s localhost:8080/health > /dev/null
done

# Monitor database connections
docker exec cost-of-living-db psql -U postgres -d cost_of_living -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'cost_of_living';"
```

## Next Steps

After verifying everything works:
1. Commit all changes
2. Move to Iteration 1.3 (CRUD operations)
3. Keep database running for development
