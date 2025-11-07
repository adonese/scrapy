# Test Infrastructure Guide

This directory contains comprehensive testing infrastructure for the UAE Cost of Living project, including HTML fixtures, test helpers, and integration tests.

## Directory Structure

```
test/
├── fixtures/          # HTML fixtures for all scrapers
│   ├── bayut/        # Bayut real estate fixtures (5 files)
│   ├── dubizzle/     # Dubizzle housing fixtures (4 files)
│   ├── dewa/         # DEWA utility rates fixture
│   ├── sewa/         # SEWA utility rates fixture
│   ├── aadc/         # AADC utility rates fixture
│   └── rta/          # RTA transportation fares fixture
├── helpers/          # Test helper utilities
│   ├── fixtures.go   # Fixture loading functions
│   ├── mock_server.go # HTTP mock server
│   └── assertions.go # Custom test assertions
└── integration/      # Integration tests (created by Agent 2)
```

## Quick Start

### Loading Fixtures

```go
import "github.com/adonese/cost-of-living/test/helpers"

// Load a fixture
html, err := helpers.LoadFixture("bayut", "dubai_listings.html")

// Or panic if it fails (useful in test setup)
html := helpers.MustLoadFixture("bayut", "dubai_listings.html")

// Check if fixture exists
if helpers.FixtureExists("bayut", "dubai_listings.html") {
    // ...
}

// List all fixtures in a directory
files, err := helpers.ListFixtures("bayut")
```

### Using Mock Server

```go
import "github.com/adonese/cost-of-living/test/helpers"

// Create a mock server with specific fixtures
mockServer, err := helpers.NewBayutMockServer()
defer mockServer.Close()

// Or build a custom mock server
mockServer := helpers.NewMockServerBuilder().
    WithFixture("/path", "bayut", "dubai_listings.html").
    WithContent("/other", "<html>custom content</html>").
    MustBuild()
defer mockServer.Close()

// Get the server URL
url := mockServer.URL()
```

### Custom Assertions

```go
import (
    "github.com/adonese/cost-of-living/test/helpers"
    "github.com/adonese/cost-of-living/internal/models"
)

func TestScraper(t *testing.T) {
    // ... scraping logic ...

    // Assert a single data point
    helpers.AssertHousingDataPoint(t, cdp)

    // Assert with specific expectations
    helpers.AssertCostDataPoint(t, cdp, helpers.CostDataPointExpectations{
        MinPrice: 10000,
        MaxPrice: 1000000,
        Category: "Housing",
        Source: "bayut",
    })

    // Assert data point count
    helpers.AssertDataPointCount(t, dataPoints, 5, 10) // between 5 and 10

    // Assert all data points are valid
    helpers.AssertAllDataPointsValid(t, dataPoints)

    // Assert no duplicates
    helpers.AssertNoDuplicates(t, dataPoints)

    // Assert location is valid
    helpers.AssertLocationValid(t, location)

    // Assert price is in range
    helpers.AssertPriceInRange(t, price, "yearly_rent")

    // Assert source URL contains domain
    helpers.AssertSourceURL(t, url, "bayut.com")

    // Assert tags contain required values
    helpers.AssertTagsContain(t, tags, "rent", "apartment")

    // Print summary (useful for debugging)
    helpers.PrintDataPointSummary(t, dataPoints)
}
```

## Available Fixtures

### Bayut Fixtures (Real Estate)
- `dubai_listings.html` - 10 property listings across Dubai (studios to 4BR, various areas)
- `sharjah_listings.html` - 5 property listings in Sharjah
- `ajman_listings.html` - 4 property listings in Ajman
- `abudhabi_listings.html` - 6 property listings in Abu Dhabi
- `empty_results.html` - Empty search results page

**Price Range**: AED 18,000 - 450,000/year
**Property Types**: Studio, 1BR, 2BR, 3BR, 4BR
**Areas**: Dubai Marina, Downtown, JLT, Business Bay, Sharjah, Ajman, Abu Dhabi, etc.

### Dubizzle Fixtures (Housing)
- `apartments.html` - 8 apartment listings (full apartments for rent)
- `bedspace.html` - 6 bedspace listings (shared room accommodations)
- `roomspace.html` - 8 roomspace listings (private room in shared apartment)
- `error_page.html` - Anti-bot error page (Incapsula/Cloudflare)

**Price Range**:
- Apartments: AED 32,000 - 155,000/year
- Bedspace: AED 450 - 700/month
- Roomspace: AED 900 - 2,200/month

### Utility Provider Fixtures
- `dewa/rates_table.html` - DEWA electricity and water tariff tables
- `sewa/tariff_page.html` - SEWA electricity and water rates
- `aadc/rates.html` - AADC utility rates for Abu Dhabi
- `rta/fare_calculator.html` - RTA public transport fares (Metro, Bus, Tram, Taxi)

## Writing Tests with Fixtures

### Example: Testing a Scraper with Fixtures

```go
package myscraper_test

import (
    "strings"
    "testing"

    "github.com/PuerkitoBio/goquery"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/adonese/cost-of-living/test/helpers"
)

func TestMyScraperWithFixture(t *testing.T) {
    // Load fixture
    html := helpers.MustLoadFixture("bayut", "dubai_listings.html")

    // Parse HTML
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    require.NoError(t, err)

    // Create scraper and extract data
    scraper := NewMyScraper(config)
    dataPoints := scraper.extractListings(doc, "https://example.com")

    // Validate results
    helpers.AssertDataPointCount(t, dataPoints, 5, 15)
    helpers.AssertAllDataPointsValid(t, dataPoints)
    helpers.AssertNoDuplicates(t, dataPoints)

    // Print summary for debugging
    if testing.Verbose() {
        helpers.PrintDataPointSummary(t, dataPoints)
    }
}
```

### Example: Testing with Mock Server

```go
func TestScraperWithMockServer(t *testing.T) {
    // Create mock server
    mockServer, err := helpers.NewBayutMockServer()
    require.NoError(t, err)
    defer mockServer.Close()

    // Note: To use mock server, your scraper needs to accept
    // a base URL parameter, or you need to test extraction logic separately

    t.Logf("Mock server running at: %s", mockServer.URL())
}
```

## Fixture Design Principles

All fixtures follow these principles:

1. **Realistic**: Based on actual website HTML structures
2. **Comprehensive**: Cover multiple scenarios (normal, edge cases, errors)
3. **Multi-emirate**: Include data from Dubai, Sharjah, Ajman, Abu Dhabi
4. **Price Variety**: Range from budget to luxury options
5. **Property Types**: Studios, 1-4 bedrooms, shared accommodation
6. **Parseable**: Validated with actual scraper code
7. **Maintainable**: Clean, well-formatted HTML

## Adding New Fixtures

When adding new fixtures:

1. Create directory: `test/fixtures/my_scraper/`
2. Add realistic HTML files based on actual website structure
3. Include edge cases (empty results, error pages)
4. Update this README with fixture details
5. Add helper functions to `test/helpers/` if needed
6. Write tests to validate fixtures work with your parser

## Integration Tests

Integration tests using these fixtures are located in:
- `internal/scrapers/bayut/bayut_integration_test.go`
- `internal/scrapers/dubizzle/dubizzle_integration_test.go`
- `test/integration/` (comprehensive suite by Agent 2)

## Running Tests

```bash
# Run all tests
go test ./...

# Run only helper tests
go test ./test/helpers/...

# Run integration tests
go test ./internal/scrapers/bayut/... ./internal/scrapers/dubizzle/...

# Run with verbose output
go test -v ./test/...

# Run specific test
go test -run TestLoadFixture ./test/helpers/...
```

## Notes for Future Agents

### Agent 2 (Integration Testing)
- All fixtures are ready to use in `test/fixtures/`
- Helper functions available in `test/helpers/`
- Examples in `*_integration_test.go` files
- Mock server is ready for end-to-end testing

### Agents 4-8 (New Scrapers)
- Follow the pattern in `test/fixtures/bayut/` for your fixtures
- Use `helpers.LoadFixture()` to load your HTML
- Use `helpers.NewMockServer()` for HTTP testing
- Reference `bayut_integration_test.go` as an example
- Utility provider fixtures (DEWA, SEWA, AADC, RTA) are ready for your use

## Troubleshooting

**Problem**: Fixture not found
**Solution**: Use absolute paths or `helpers.GetFixturePath()` function

**Problem**: Tests fail in CI
**Solution**: Ensure fixtures are committed to git (they are)

**Problem**: Need different HTML structure
**Solution**: Create new fixture file or use `mock_server.WithContent()`

**Problem**: Price/location validation fails
**Solution**: Check if data matches expected format, adjust assertions or fixture

## Resources

- Bayut Integration Tests: `/internal/scrapers/bayut/bayut_integration_test.go`
- Dubizzle Integration Tests: `/internal/scrapers/dubizzle/dubizzle_integration_test.go`
- Orchestration Document: `/AGENT_ORCHESTRATION.md`
- CI/CD Guide: `/CI_CD_GUIDE.md`

---

Created by Agent 1 - Wave 1: Testing Infrastructure
Last Updated: 2025-11-07
