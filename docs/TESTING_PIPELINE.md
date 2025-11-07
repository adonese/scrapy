# Testing Pipeline Implementation
## Comprehensive Testing Strategy for UAE Cost of Living Scrapers

**Date:** November 7, 2025
**Purpose:** Establish robust testing to validate all scrapers before production deployment

---

## 🎯 Testing Philosophy

### Core Principles
1. **Test First**: Write tests before implementing scrapers
2. **Mock External**: Never hit real websites in unit tests
3. **Validate Everything**: Every scraper output must pass validation
4. **Fail Fast**: Catch issues in CI before deployment
5. **Monitor Continuously**: Production validation on schedule

---

## 📁 Testing Structure

```
cost-of-living/
├── test/
│   ├── fixtures/                 # Mock HTML/JSON responses
│   │   ├── bayut/
│   │   │   ├── dubai_listings.html
│   │   │   ├── sharjah_listings.html
│   │   │   └── empty_results.html
│   │   ├── dubizzle/
│   │   │   ├── apartments.html
│   │   │   ├── bedspace.html
│   │   │   └── error_page.html
│   │   ├── dewa/
│   │   │   └── rates_table.html
│   │   ├── sewa/
│   │   │   └── tariff_page.html
│   │   └── rta/
│   │       └── fare_calculator.html
│   │
│   ├── integration/              # End-to-end tests
│   │   ├── scraper_test.go
│   │   ├── workflow_test.go
│   │   └── api_test.go
│   │
│   ├── validation/               # Data quality tests
│   │   ├── price_validation_test.go
│   │   ├── location_validation_test.go
│   │   └── duplicate_detection_test.go
│   │
│   └── helpers/                  # Test utilities
│       ├── database.go
│       ├── fixtures.go
│       └── assertions.go
```

---

## 🧪 Testing Layers

### Layer 1: Unit Tests (Fastest)
**Purpose:** Test individual functions and parsers
**Location:** Next to source files (*_test.go)
**Execution Time:** <1 second
**Coverage Target:** 80%

```go
// internal/scrapers/bayut/parser_test.go
func TestParsePrice(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected float64
    }{
        {"Price with comma", "AED 85,000/year", 85000},
        {"Price without currency", "120,000", 120000},
        {"Monthly price", "7,500/month", 90000}, // Annualized
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := parsePrice(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Layer 2: Integration Tests (With Mocks)
**Purpose:** Test scraper against mock HTML
**Location:** test/integration/
**Execution Time:** <10 seconds
**Coverage Target:** 70%

```go
// test/integration/bayut_integration_test.go
func TestBayutScraperWithMockHTML(t *testing.T) {
    // Load fixture
    html := fixtures.Load("bayut/dubai_listings.html")

    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(html))
    }))
    defer server.Close()

    // Configure scraper with mock URL
    scraper := bayut.NewScraper(bayut.Config{
        BaseURL: server.URL,
    })

    // Run scraper
    results, err := scraper.Scrape(context.Background())

    // Validate results
    assert.NoError(t, err)
    assert.Len(t, results, 25) // Expected listings
    assert.Greater(t, results[0].Price, 0.0)
    assert.NotEmpty(t, results[0].Location.Emirate)
}
```

### Layer 3: Workflow Tests
**Purpose:** Test Temporal workflows
**Location:** internal/workflow/*_test.go
**Execution Time:** <30 seconds
**Coverage Target:** 60%

```go
// internal/workflow/scraper_workflow_test.go
func TestScraperWorkflowEndToEnd(t *testing.T) {
    suite := &testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()

    // Mock activities
    env.OnActivity(RunScraperActivity, mock.Anything, "dewa").
        Return(&ScraperResult{
            ItemsScraped: 15,
            ItemsSaved:   15,
        }, nil)

    // Execute workflow
    env.ExecuteWorkflow(ScraperWorkflow, ScraperInput{
        ScraperName: "dewa",
    })

    // Verify completion
    assert.True(t, env.IsWorkflowCompleted())
    assert.NoError(t, env.GetWorkflowError())
}
```

### Layer 4: Validation Tests
**Purpose:** Ensure data quality
**Location:** test/validation/
**Execution Time:** <5 seconds
**Coverage Target:** 90%

```go
// test/validation/price_validation_test.go
func TestPriceValidation(t *testing.T) {
    validator := validation.NewPriceValidator()

    tests := []struct {
        name     string
        category string
        price    float64
        valid    bool
    }{
        {"Valid apartment price", "Housing", 50000, true},
        {"Too high apartment", "Housing", 5000000, false},
        {"Negative price", "Housing", -1000, false},
        {"Valid utility", "Utilities", 300, true},
        {"Too high utility", "Utilities", 10000, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.category, tt.price)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Layer 5: Live Validation (Production)
**Purpose:** Validate real scraped data
**Location:** Scripts and monitoring
**Execution Time:** Continuous
**Alert Threshold:** <95% valid

```bash
#!/bin/bash
# scripts/validate-live-data.sh

echo "Validating last 24 hours of scraped data..."

# Check for price outliers
psql -d cost_of_living -c "
SELECT source, COUNT(*) as outliers
FROM cost_data_points
WHERE recorded_at > NOW() - INTERVAL '24 hours'
  AND (price < 0 OR price > 1000000)
GROUP BY source;"

# Check for missing required fields
psql -d cost_of_living -c "
SELECT source, COUNT(*) as missing_location
FROM cost_data_points
WHERE recorded_at > NOW() - INTERVAL '24 hours'
  AND (location->>'emirate' IS NULL)
GROUP BY source;"

# Check for duplicates
psql -d cost_of_living -c "
WITH duplicates AS (
  SELECT item_name, price, location, source, COUNT(*) as dup_count
  FROM cost_data_points
  WHERE recorded_at > NOW() - INTERVAL '24 hours'
  GROUP BY item_name, price, location, source
  HAVING COUNT(*) > 1
)
SELECT source, SUM(dup_count) as total_duplicates
FROM duplicates
GROUP BY source;"
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow
```yaml
# .github/workflows/test.yml
name: Test Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: timescale/timescaledb:latest-pg16
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: cost_of_living_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: |
          go mod download
          go install gotest.tools/gotestsum@latest

      - name: Run migrations
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/cost_of_living_test?sslmode=disable
        run: |
          go run cmd/migrate/main.go up

      - name: Run unit tests
        run: |
          gotestsum --format testname -- -v -cover -race ./...

      - name: Run integration tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/cost_of_living_test?sslmode=disable
        run: |
          go test -v ./test/integration/...

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage.html

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go test -cover ./... | grep -oP '\d+\.\d+(?=%)')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "Coverage below 70% threshold"
            exit 1
          fi
```

### Scraper Validation Workflow
```yaml
# .github/workflows/scraper-validation.yml
name: Validate Scrapers

on:
  schedule:
    - cron: '0 */6 * * *' # Every 6 hours
  workflow_dispatch:

jobs:
  validate:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Test Bayut Scraper
        run: |
          go run cmd/scraper/main.go -scraper bayut -test-mode
        continue-on-error: true

      - name: Test Dubizzle Scraper
        run: |
          go run cmd/scraper/main.go -scraper dubizzle -test-mode
        continue-on-error: true

      - name: Validate scraped data
        run: |
          ./scripts/validate-test-data.sh

      - name: Send alerts on failure
        if: failure()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Scraper validation failed!'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

---

## 📊 Test Data Fixtures

### Creating Realistic Fixtures
```go
// test/helpers/fixtures.go
package helpers

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type FixtureLoader struct {
    basePath string
}

func NewFixtureLoader() *FixtureLoader {
    return &FixtureLoader{
        basePath: "test/fixtures",
    }
}

func (f *FixtureLoader) LoadHTML(path string) (string, error) {
    fullPath := filepath.Join(f.basePath, path)
    data, err := os.ReadFile(fullPath)
    return string(data), err
}

func (f *FixtureLoader) LoadJSON(path string, v interface{}) error {
    fullPath := filepath.Join(f.basePath, path)
    data, err := os.ReadFile(fullPath)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, v)
}

// Generate fixture from real data
func (f *FixtureLoader) SaveFixture(path string, data interface{}) error {
    fullPath := filepath.Join(f.basePath, path)

    // Ensure directory exists
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Marshal and save
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(fullPath, jsonData, 0644)
}
```

### Sample Fixture: DEWA Rates
```html
<!-- test/fixtures/dewa/rates_table.html -->
<div class="rates-table">
    <h2>Electricity Tariff</h2>
    <table>
        <thead>
            <tr>
                <th>Consumption Slab (kWh)</th>
                <th>Fils/kWh</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>0 - 2,000</td>
                <td>23</td>
            </tr>
            <tr>
                <td>2,001 - 4,000</td>
                <td>28</td>
            </tr>
            <tr>
                <td>4,001 - 6,000</td>
                <td>32</td>
            </tr>
            <tr>
                <td>Above 6,000</td>
                <td>38</td>
            </tr>
        </tbody>
    </table>

    <h2>Water Tariff</h2>
    <table>
        <thead>
            <tr>
                <th>Consumption Slab (IG)</th>
                <th>Fils/IG</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>0 - 6,000</td>
                <td>3.5</td>
            </tr>
            <tr>
                <td>6,001 - 12,000</td>
                <td>4.0</td>
            </tr>
            <tr>
                <td>Above 12,000</td>
                <td>4.6</td>
            </tr>
        </tbody>
    </table>
</div>
```

---

## 🚨 Testing Commands

### Quick Test Commands
```bash
# Run all tests
make test-all

# Run unit tests only
go test ./internal/...

# Run integration tests
go test ./test/integration/...

# Run with coverage
go test -cover ./...

# Run specific scraper tests
go test ./internal/scrapers/bayut/...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Validation Commands
```bash
# Validate fixture data
./scripts/validate-fixtures.sh

# Check for duplicates
./scripts/check-duplicates.sh

# Validate price ranges
./scripts/validate-prices.sh

# Run all validations
make validate-all
```

---

## 📈 Testing Metrics

### Coverage Goals
- **Overall**: 75% minimum
- **Scrapers**: 80% minimum
- **Parsers**: 90% minimum
- **Workflows**: 60% minimum
- **API Handlers**: 70% minimum

### Performance Benchmarks
- **Unit Tests**: < 1 second total
- **Integration Tests**: < 30 seconds total
- **Full Test Suite**: < 2 minutes
- **CI Pipeline**: < 5 minutes

### Quality Gates
- [ ] All tests passing
- [ ] Coverage above threshold
- [ ] No race conditions
- [ ] No memory leaks
- [ ] Fixtures up to date
- [ ] Validation rules passing

---

## 🔍 Debugging Failed Tests

### Common Issues and Solutions

#### 1. Selector Changes
```go
// Before: Brittle selector
price := doc.Find(".price-tag span").Text()

// After: Multiple fallbacks
price := doc.Find(".price-tag span, .price-amount, [data-price]").First().Text()
```

#### 2. Timing Issues
```go
// Add retries for flaky tests
func TestWithRetry(t *testing.T, maxRetries int, testFunc func() error) {
    var err error
    for i := 0; i < maxRetries; i++ {
        if err = testFunc(); err == nil {
            return
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    t.Fatalf("Test failed after %d retries: %v", maxRetries, err)
}
```

#### 3. Database State
```go
// Clean database before each test
func TestWithCleanDB(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Clean relevant tables
    db.Exec("TRUNCATE cost_data_points CASCADE")

    // Run test
    // ...
}
```

---

## 🎯 Next Steps

1. **Implement Wave 1 Agents** - Set up testing infrastructure
2. **Create Mock Fixtures** - Build comprehensive test data
3. **Write Integration Tests** - Cover all existing scrapers
4. **Set Up CI/CD** - Automate testing on every commit
5. **Add Validation Rules** - Ensure data quality
6. **Monitor Production** - Continuous validation

---

**Ready to implement robust testing? Start with Wave 1! 🧪**