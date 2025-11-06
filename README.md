# UAE Cost of Living Calculator

A comprehensive UAE cost of living calculator with Go backend, PostgreSQL/TimescaleDB, and Templ + HTMX frontend.

## Current Status: Iteration 1.2

This iteration includes:
- Go HTTP server with health check endpoint
- PostgreSQL 15 with TimescaleDB extension
- First database migration (cost_data_points table)
- Database connection pooling and health checks

## Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- Make (optional, but recommended)

## Quick Start

### 1. Set up environment variables (optional)

```bash
# Copy example env file
cp .env.example .env

# Edit .env if you want to change database credentials
```

### 2. Start the database

```bash
# Start PostgreSQL with TimescaleDB
make db-up

# Wait a few seconds for the database to be ready
```

### 3. Run migrations

```bash
# Apply database migrations
make migrate

# Check migration status
make migrate-version
```

### 4. Start the API server

```bash
# Using make
make run

# Or directly with go
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

### 5. Test the health endpoint

```bash
curl localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2025-11-06T19:33:27+04:00"
}
```

## Build

```bash
# Build binary
make build

# Run the binary
./bin/api
```

## Docker

```bash
# Build image
docker build -t cost-of-living:latest .

# Run container
docker run -p 8080:8080 cost-of-living:latest
```

## Project Structure

```
.
├── cmd/
│   ├── api/
│   │   └── main.go          # Application entry point
│   └── migrate/
│       └── main.go          # Migration runner CLI
├── internal/
│   └── handlers/
│       ├── health.go        # Health check handler
│       └── health_test.go   # Health check tests
├── pkg/
│   └── database/
│       ├── database.go      # Database connection package
│       └── database_test.go # Database tests
├── migrations/
│   ├── 001_create_cost_data_points.up.sql   # Create table migration
│   └── 001_create_cost_data_points.down.sql # Rollback migration
├── docker-compose.yml       # PostgreSQL with TimescaleDB
├── Dockerfile               # Multi-stage Docker build
├── Makefile                 # Common commands
├── .env.example             # Example environment variables
├── go.mod                   # Go module definition
└── README.md               # This file
```

## Available Endpoints

- `GET /health` - Health check endpoint (returns status, database connection, and timestamp)

## Development

### Database Commands

```bash
# Start database
make db-up

# Stop database
make db-down

# View database logs
make db-logs

# Run migrations
make migrate

# Rollback last migration
make migrate-down

# Check migration version
make migrate-version
```

### Running tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run integration tests (requires running database)
go test -v ./pkg/database
```

### Clean build artifacts

```bash
make clean
```

### Direct Database Access

```bash
# Connect to PostgreSQL
docker exec -it cost-of-living-db psql -U postgres -d cost_of_living

# View tables
\dt

# Describe cost_data_points table
\d cost_data_points

# View hypertables
SELECT * FROM timescaledb_information.hypertables;
```

## Database Schema

### cost_data_points Table

The main table for storing cost data points with TimescaleDB hypertable partitioning:

- **id**: UUID primary key
- **category**: Category (Housing, Food, etc.)
- **sub_category**: Subcategory
- **item_name**: Specific item name
- **price**: Main price value
- **min_price, max_price, median_price**: Price range statistics
- **sample_size**: Number of data points
- **location**: JSONB with emirate, city, area, coordinates
- **recorded_at**: Timestamp (used for hypertable partitioning)
- **valid_from, valid_to**: Validity period
- **source**: Data source identifier
- **source_url**: URL reference
- **confidence**: Confidence score (0.0 to 1.0)
- **unit**: Currency unit (default: AED)
- **tags**: Array of tags
- **attributes**: JSONB for flexible additional data

**Indexes:**
- Primary key on (id, recorded_at)
- GIN indexes on location, tags, and attributes
- B-tree indexes on category, sub_category, item_name, recorded_at
- Specialized indexes on location fields (emirate, city)

**TimescaleDB Features:**
- Hypertable with 1-month chunk intervals
- Optimized for time-series queries

## Next Steps (Iteration 1.3)

- Implement CRUD operations for cost data points
- Add repository pattern
- Create REST API endpoints for data management
- Add validation and error handling

## Design Principles

- Keep it simple
- Readable and debuggable code
- Pragmatic solutions over perfect architecture
- No unnecessary complexity
