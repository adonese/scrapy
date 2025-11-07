# Integration Tests

Comprehensive integration tests for the UAE Cost of Living scrapers.

## Overview

This directory contains integration tests that verify:
- Scraper functionality with mock HTTP servers
- HTML parsing and data extraction
- Error handling and retry logic
- Rate limiting and timeout behavior
- Multi-emirate support
- Temporal workflow integration
- Database integration (when available)

## Test Structure

```
test/integration/
├── README.md                          # This file
├── setup_test.go                      # Test setup and initialization
├── helpers_test.go                    # Reusable test helpers
├── mock_html.go                       # Mock HTML fixtures
├── bayut_integration_test.go          # Bayut scraper tests
├── dubizzle_integration_test.go       # Dubizzle scraper tests
└── workflow_integration_test.go       # Workflow integration tests
```

## Running Tests

### All Integration Tests
```bash
go test -v ./test/integration/...
```

### With Coverage
```bash
go test -v -coverprofile=coverage.out ./test/integration/...
go tool cover -html=coverage.out
```

### Using Test Scripts
```bash
# Run integration tests
./scripts/test-integration.sh

# Run scraper tests specifically
./scripts/test-scrapers.sh

# Generate coverage reports
./scripts/test-coverage.sh
```

### Short Mode (Skips Long-Running Tests)
```bash
go test -short ./test/integration/...
```

## Test Categories

### Bayut Scraper Tests
- **TestBayutScraperIntegration**: Basic scraper functionality
- **TestBayutScraperWithMockHTML**: HTML parsing scenarios
- **TestBayutScraperRateLimiting**: Rate limit enforcement
- **TestBayutScraperTimeout**: Timeout and context handling
- **TestBayutScraperMultiEmirate**: Multi-emirate support
- **TestBayutScraperDataValidation**: Data quality validation
- **TestBayutScraperConcurrency**: Concurrent scraper execution
- **TestBayutScraperWithContext**: Context cancellation

### Dubizzle Scraper Tests
- **TestDubizzleScraperIntegration**: Basic scraper functionality
- **TestDubizzleScraperWithMockHTML**: HTML parsing and bot detection
- **TestDubizzleScraperRetryLogic**: Retry mechanism testing
- **TestDubizzleScraperBotDetection**: Anti-bot page detection
- **TestDubizzleScraperMultiEmirate**: Multi-emirate support
- **TestDubizzleScraperSharedAccommodation**: Bedspace/roomspace categories
- **TestDubizzleScraperHeadersForAntiBot**: Browser-like headers
- **TestDubizzleScraperTimeout**: Timeout handling
- **TestDubizzleScraperConcurrency**: Concurrent execution

### Workflow Tests
- **TestScraperWorkflowIntegration**: Basic workflow execution
- **TestScraperWorkflowWithRetry**: Retry logic
- **TestScraperWorkflowWithCompensation**: Compensation activities
- **TestScheduledScraperWorkflowIntegration**: Scheduled workflows
- **TestScraperServiceIntegration**: Service-level integration
- **TestBatchScrapingWorkflow**: Batch processing
- **TestWorkflowWithDatabaseIntegration**: Database operations
- **TestWorkflowTimeouts**: Timeout handling

## Mock Data

### Mock HTML Fixtures
Located in `mock_html.go`:
- `MockBayutHTML`: Sample Bayut search results page
- `MockBayutEmptyHTML`: Empty results page
- `MockBayutMalformedHTML`: Malformed HTML for error testing
- `MockDubizzleHTML`: Sample Dubizzle listings
- `MockDubizzleEmptyHTML`: Empty results
- `MockDubizzleBotDetectedHTML`: Bot detection page (Incapsula)
- `MockDubizzleCloudflareHTML`: Cloudflare challenge page
- `MockDubizzleMalformedHTML`: Malformed HTML

### Test Helpers
Located in `helpers_test.go`:
- `setupTestDB()`: Creates test database connection
- `setupTestRepository()`: Creates test repository
- `cleanupTestData()`: Removes test data
- `NewMockHTTPServer()`: Creates mock HTTP server
- `ValidateCostDataPoint()`: Validates data point structure
- `CreateTestDataPoint()`: Creates test data points

## Coverage Goals

- **Overall**: > 51% (achieved)
- **Scrapers**: > 50% (Bayut: 56.4%, Dubizzle: 57.7%)
- **Workflows**: > 59% (achieved)
- **Repository**: > 86% (achieved)

## CI/CD Integration

These tests are automatically run by the CI/CD pipeline on:
- Every pull request
- Every push to main
- Scheduled runs (every 6 hours for scraper validation)

See `.github/workflows/test.yml` for pipeline configuration.

## Writing New Integration Tests

1. Add test file: `[component]_integration_test.go`
2. Follow naming convention: `Test[Component][Feature]`
3. Use test helpers from `helpers_test.go`
4. Add mock data to `mock_html.go` if needed
5. Ensure tests are idempotent and can run in parallel
6. Use `-short` flag for long-running tests

### Example Test
```go
func TestMyScraperIntegration(t *testing.T) {
    mockServer := NewMockHTTPServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(MockHTML))
    })
    defer mockServer.Close()

    scraper := myscraper.NewScraper(config)
    dataPoints, err := scraper.Scrape(context.Background())

    require.NoError(t, err)
    ValidateDataPoints(t, dataPoints, 1)
}
```

## Troubleshooting

### Tests Fail with "connection refused"
- Tests are trying to hit real websites instead of mocks
- Check that HTTP client is using mock server URL

### Tests Timeout
- Increase timeout: `go test -timeout 5m`
- Check for infinite loops in retry logic
- Verify rate limiting isn't too aggressive

### Coverage Not Updating
- Run with `-coverprofile` flag
- Delete old coverage files: `rm coverage*.out`
- Rebuild test binaries: `go test -c`

## Future Improvements

- [ ] Add more edge case scenarios
- [ ] Test database integration with real database
- [ ] Add performance benchmarks
- [ ] Test concurrent scraper execution with real load
- [ ] Add integration with Agent 1's fixture files when ready

## References

- [Testing Guide](../../docs/testing.md)
- [CI/CD Guide](../../CI_CD_GUIDE.md)
- [Scraper Documentation](../../docs/scrapers.md)
