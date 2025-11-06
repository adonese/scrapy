# Wave 1: Foundation (Iterative Approach)

## Philosophy
- Build → Test → Verify → Next
- Each step produces testable output
- Natural stopping points for review
- Fast feedback loops over brute forcing

## Iteration 1.1: Minimal Project Structure
**Goal**: Get a hello-world Go server running
**Agent**: Project Scaffolding (minimal)
**Deliverables**:
```
- go.mod initialized
- Basic directory structure (cmd/api only)
- Simple HTTP handler that returns "OK"
- Dockerfile that builds and runs
- Can run: make run && curl localhost:8080/health
```
**Test**: Server responds to health check
**Stop Point**: Review structure before continuing

---

## Iteration 1.2: Database Connection
**Goal**: Connect to PostgreSQL, run one migration
**Agent**: Database Setup (minimal)
**Dependencies**: 1.1 complete
**Deliverables**:
```
- docker-compose.yml with PostgreSQL only
- One migration file (create one table)
- Database connection utility
- Can run: make db-up && make migrate
```
**Test**: Connection successful, table created
**Stop Point**: Verify DB schema before adding more

---

## Iteration 1.3: First Data Model
**Goal**: Create and query one cost_data_point
**Agent**: Database Setup (continuation)
**Dependencies**: 1.2 complete
**Deliverables**:
```
- Complete cost_data_points table migration
- Repository pattern implementation
- CRUD operations for cost data
- Seed script with 10 sample records
```
**Test**: Insert and query sample cost data
**Stop Point**: Validate data model works as expected

---

## Iteration 1.4: API Endpoint
**Goal**: Expose first REST endpoint
**Agent**: API Framework (minimal)
**Dependencies**: 1.3 complete
**Deliverables**:
```
- Echo/Fiber framework setup
- GET /api/v1/costs endpoint
- JSON serialization
- Basic error handling
```
**Test**: curl localhost:8080/api/v1/costs returns data
**Stop Point**: Review API design before expanding

---

## Iteration 1.5: Temporal Hello World
**Goal**: Run simplest possible workflow
**Agent**: Temporal Foundation (minimal)
**Dependencies**: 1.1 complete (parallel with 1.2-1.4)
**Deliverables**:
```
- docker-compose updated with Temporal + SQLite
- One workflow: HelloWorkflow that prints message
- One activity: LogActivity
- Worker that executes workflow
```
**Test**: Start worker, trigger workflow, see logs
**Stop Point**: Verify Temporal setup before complex workflows

---

## Iteration 1.6: Basic Observability
**Goal**: See one metric, one trace
**Agent**: Monitoring Stack (minimal)
**Dependencies**: 1.4 complete
**Deliverables**:
```
- OpenTelemetry SDK integrated
- Prometheus metrics endpoint
- One metric: http_requests_total
- One traced endpoint
- docker-compose updated with Prometheus
```
**Test**: curl /metrics, see counter increment
**Stop Point**: Validate monitoring works before full stack

---

## Iteration 1.7: First Scraper Skeleton
**Goal**: Scrape one page from one site
**Agent**: Scraper Architecture (minimal)
**Dependencies**: 1.3 complete
**Deliverables**:
```
- Scraper interface defined
- One implementation: BayutScraper (stub)
- Fetches one listing page
- Parses one field (price)
- Stores in database
```
**Test**: Run scraper, verify one record in DB
**Stop Point**: Review scraper pattern before scaling

---

## Iteration 1.8: Workflow + Scraper
**Goal**: Run scraper via Temporal workflow
**Agent**: Integration (1.5 + 1.7)
**Dependencies**: 1.5, 1.7 complete
**Deliverables**:
```
- ScraperWorkflow that calls scraper activity
- Error handling and retries
- Schedule workflow to run on demand
```
**Test**: Trigger workflow, data appears in DB
**Stop Point**: System working end-to-end (minimal)

---

## Success Criteria for Wave 1
After iteration 1.8, we should have:
- ✅ Running Go API server
- ✅ PostgreSQL with one table
- ✅ One REST endpoint returning data
- ✅ Temporal executing workflows
- ✅ Basic metrics and tracing
- ✅ One scraper collecting data
- ✅ End-to-end data flow working

**Total Build Time**: ~4-6 hours of focused work
**Lines of Code**: ~1000-1500 (not thousands)
**Complexity**: Minimal but complete

## What We DON'T Do Yet
- ❌ All migrations (just core tables)
- ❌ All scrapers (just one skeleton)
- ❌ Full monitoring stack (just basics)
- ❌ Complete API (just one endpoint)
- ❌ Frontend (that's Wave 2)
- ❌ Complex workflows (just simple ones)

## Testing Between Iterations

After each iteration, run:
```bash
# Health check
make health-check

# Integration test
make test-integration

# Manual verification
make verify-iteration
```

## Rollback Strategy
Each iteration is atomic:
- Git commit after each successful iteration
- Can rollback to any iteration
- Iterations don't break previous work

## Context for Sub-Agents

Each agent receives:
1. **data_models.md** - Data structure reference
2. **plan.md** - Overall vision
3. **Previous iteration output** - What's already built
4. **Current iteration spec** - Specific deliverables
5. **Test requirements** - How to verify success

This enables agents to:
- Understand the bigger picture
- Build on previous work
- Know exactly what to deliver
- Verify their work independently