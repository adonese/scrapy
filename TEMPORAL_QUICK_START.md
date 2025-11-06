# Temporal Quick Start Guide

## Quick Setup (3 Steps)

### 1. Start Temporal Server
```bash
make temporal-up
```

This starts:
- Temporal server on localhost:7233
- Web UI on http://localhost:8233

### 2. Verify Temporal is Running
```bash
docker logs cost-of-living-temporal
```

You should see:
```
Temporal server: 0.0.0.0:7233
Web UI:          http://0.0.0.0:8233
```

### 3. Run Tests
```bash
make test
```

All tests should pass, including:
- TestHelloWorkflow
- TestHelloActivity

## Workflow Components

### Hello World Workflow
Located in `/internal/workflow/`:

**hello.go** - Defines the workflow logic:
```go
func HelloWorkflow(ctx workflow.Context, input HelloWorkflowInput) (*HelloWorkflowResult, error)
```

**activities.go** - Defines activities:
```go
func HelloActivity(ctx context.Context, name string) (string, error)
```

### Worker
Executes workflows and activities:
```bash
make worker
```

### Client
Triggers workflow execution:
```bash
make run-workflow
```

## Architecture

```
┌─────────────┐
│  Client     │  Triggers workflow execution
│  (Go code)  │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│  Temporal Server    │  Manages workflow state
│  (Docker container) │  Stores execution history
└──────┬──────────────┘
       │
       ▼
┌─────────────┐
│   Worker    │  Executes workflow code
│  (Go code)  │  Runs activities
└─────────────┘
```

## Task Queue

All workflows use: `cost-of-living-task-queue`

Workers poll this queue for:
- Workflow tasks (decisions)
- Activity tasks (work to be done)

## Temporal UI

Access at http://localhost:8233

View:
- Workflow executions
- Activity results
- Execution history
- Task queue status
- Namespace information

## Useful Commands

```bash
# View Temporal logs
docker logs cost-of-living-temporal

# Access Temporal CLI inside container
docker exec -it cost-of-living-temporal temporal --help

# List workflows
docker exec cost-of-living-temporal temporal workflow list

# List namespaces
docker exec cost-of-living-temporal temporal operator namespace list

# Stop Temporal
make temporal-down

# Restart Temporal
docker-compose restart temporal
```

## Testing

### Unit Tests (Recommended)
```bash
go test -v ./internal/workflow/...
```

Uses Temporal test suite:
- No server needed
- Fast execution
- Tests logic in isolation

### Manual Testing
1. Start Temporal: `make temporal-up`
2. Start Worker: `make worker` (Terminal 1)
3. Run Workflow: `make run-workflow` (Terminal 2)
4. Check UI: http://localhost:8233

## Troubleshooting

### Temporal Won't Start
```bash
# Check logs
docker logs cost-of-living-temporal

# Restart
docker-compose down temporal
make temporal-up
```

### Worker Can't Connect
**Note**: Known WSL2 networking issue. Worker/client from host may not connect.

**Solutions**:
1. Run tests instead (they work fine)
2. Use Docker exec to test inside container
3. Deploy to non-WSL environment

### Port Already in Use
```bash
# Check what's using port 7233
lsof -i :7233

# Or port 8233 for UI
lsof -i :8233
```

## Next Steps

After Iteration 1.5, you can:
1. Add logging/metrics (Iteration 1.6)
2. Build scraping workflows (Iteration 1.7)
3. Add more complex workflow patterns
4. Implement retry policies
5. Add workflow versioning

## File Structure

```
/home/adonese/src/cost-of-living/
├── cmd/worker/main.go              # Worker process
├── examples/workflow_client.go     # Example client
├── internal/workflow/
│   ├── hello.go                    # Workflow definition
│   ├── activities.go               # Activity implementations
│   └── hello_test.go               # Tests
└── docker-compose.yml              # Temporal server config
```

## Key Concepts

**Workflow**: Durable function that coordinates activities
**Activity**: Individual unit of work (can be retried)
**Worker**: Process that executes workflows and activities
**Task Queue**: Queue where tasks are distributed to workers

## Resources

- [Temporal Docs](https://docs.temporal.io/)
- [Go SDK](https://docs.temporal.io/dev-guide/go)
- [Workflow Patterns](https://docs.temporal.io/encyclopedia/workflow-message-passing)
