# UAE Cost of Living Calculator

A comprehensive UAE cost of living calculator with Go backend, PostgreSQL/TimescaleDB, and Templ + HTMX frontend.

## Current Status: Iteration 1.1

This iteration provides a minimal Go HTTP server with health check endpoint.

## Prerequisites

- Go 1.23 or later
- Make (optional, but recommended)

## Quick Start

### Running the server

```bash
# Using make
make run

# Or directly with go
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

### Testing the health endpoint

```bash
curl localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
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
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   └── handlers/
│       └── health.go        # Health check handler
├── Dockerfile               # Multi-stage Docker build
├── Makefile                 # Common commands
├── go.mod                   # Go module definition
└── README.md               # This file
```

## Available Endpoints

- `GET /health` - Health check endpoint

## Development

### Running tests

```bash
make test
```

### Clean build artifacts

```bash
make clean
```

## Next Steps (Iteration 1.2)

- Add PostgreSQL with TimescaleDB
- Implement data models
- Add database migrations
- Create data collection workflows

## Design Principles

- Keep it simple
- Readable and debuggable code
- Pragmatic solutions over perfect architecture
- No unnecessary complexity
