# Observability - Iteration 1.6

This document describes the observability infrastructure added in Iteration 1.6.

## Features

### 1. Structured Logging

- Uses Go's built-in `log/slog` package for structured JSON logging
- Located in `/pkg/logger/logger.go`
- All API startup and runtime logs are in JSON format
- Includes request details: method, path, status, latency, etc.

Example log output:
```json
{"time":"2025-11-06T20:42:18.691994889+04:00","level":"INFO","msg":"Starting UAE Cost of Living API"}
{"time":"2025-11-06T20:42:43.757423033+04:00","id":"","remote_ip":"::1","host":"localhost:8080","method":"GET","uri":"/health","status":200,"latency":5103555,"latency_human":"5.103555ms"}
```

### 2. Prometheus Metrics

- Single metric: `http_requests_total` - counts HTTP requests by method, endpoint, and status
- Located in `/pkg/metrics/metrics.go`
- Middleware in `/internal/middleware/metrics.go` automatically tracks all requests
- Metrics exposed at `/metrics` endpoint in Prometheus format

Example metrics output:
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/health",method="GET",status="200"} 6
http_requests_total{endpoint="/api/v1/cost-data-points",method="GET",status="200"} 2
```

### 3. Prometheus Server

- Prometheus server configured in `docker-compose.yml`
- Configuration file at `/monitoring/prometheus.yml`
- Prometheus UI available at http://localhost:9090
- Currently configured to scrape Temporal (containerized service)

## Usage

### Starting Prometheus

```bash
# Start Prometheus
make prom-up

# Open Prometheus UI
make prom-ui
# Or manually: http://localhost:9090

# Stop Prometheus
make prom-down
```

### Accessing Metrics

```bash
# Start the API
make run

# View metrics directly
curl http://localhost:8080/metrics

# Make some requests to generate metrics
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/cost-data-points?limit=5

# Check metrics again
curl http://localhost:8080/metrics | grep http_requests_total
```

### Testing Observability

Run the automated test script:

```bash
./scripts/test-metrics.sh
```

This will:
1. Make several API requests
2. Check that metrics are being recorded
3. Display sample metrics

## Architecture

```
API Server (Port 8080)
  ├── /metrics endpoint (Prometheus format)
  ├── Structured JSON logs (stdout)
  └── Metrics middleware (tracks all requests)

Prometheus (Port 9090)
  ├── Scrapes /metrics endpoint
  ├── Stores time-series data
  └── Provides query interface
```

## Implementation Details

### Metrics Middleware

All HTTP requests are automatically tracked via middleware:
- Method (GET, POST, etc.)
- Endpoint path (e.g., /health, /api/v1/cost-data-points)
- Response status code (200, 404, 500, etc.)

### Logger Usage

```go
import "github.com/adonese/cost-of-living/pkg/logger"

// Initialize (done in main.go)
logger.Init()

// Log messages
logger.Info("Server starting", "port", port)
logger.Error("Failed to connect", "error", err)
logger.Debug("Processing request", "id", requestID)
logger.Warn("Deprecated endpoint used", "endpoint", path)
```

## Docker Networking Note

**WSL2/Linux Consideration**: Prometheus runs in Docker and needs to access the API server running on the host. The current configuration uses the WSL2 gateway IP (172.29.64.1). If you're running on a different environment, you may need to adjust the target in `monitoring/prometheus.yml`:

- **macOS/Windows Docker Desktop**: Use `host.docker.internal:8080`
- **Linux native Docker**: Use `172.17.0.1:8080` or the bridge IP
- **WSL2**: Use the gateway IP (find with: `ip route show default | awk '/default/ {print $3}'`)

Alternatively, you can run the API in Docker with the same network as Prometheus.

## What's Next

Future iterations can add:
- More metrics (request duration histogram, database query metrics, etc.)
- Grafana dashboards for visualization
- Distributed tracing with OpenTelemetry
- Log aggregation (e.g., Loki)
- Alerting rules
- Custom exporters

## Files Created/Modified

```
pkg/metrics/metrics.go              # Prometheus metrics definitions
pkg/logger/logger.go                # Structured logging package
internal/middleware/metrics.go      # Metrics tracking middleware
cmd/api/main.go                     # Updated to use logger and metrics
docker-compose.yml                  # Added Prometheus service
monitoring/prometheus.yml           # Prometheus configuration
Makefile                           # Added prom-up, prom-down, prom-ui targets
scripts/test-metrics.sh            # Automated metrics testing
docs/OBSERVABILITY.md              # This documentation
```
