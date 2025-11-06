# Scrapers

This directory contains web scrapers for collecting cost of living data from various sources.

## Overview

The scraper system is designed to:
- Fetch real-world pricing data from websites
- Parse and normalize data into our `CostDataPoint` model
- Handle rate limiting, errors, and retries gracefully
- Provide metrics and logging for observability
- Save data to TimescaleDB for time-series analysis

## Architecture

### Scraper Interface

All scrapers implement the `Scraper` interface:

```go
type Scraper interface {
    Name() string
    Scrape(ctx context.Context) ([]*models.CostDataPoint, error)
    CanScrape() bool
}
```

### Configuration

Scrapers use a common `Config` struct:

```go
type Config struct {
    UserAgent  string
    RateLimit  int    // requests per second
    Timeout    int    // seconds
    MaxRetries int
    ProxyURL   string // optional
}
```

## Available Scrapers

### Bayut Scraper

Scrapes housing rental data from Bayut.com.

**Data Collected:**
- Category: Housing
- SubCategory: Rent
- Item: Property title
- Price: Rental price (AED/year)
- Location: Emirate, City, Area
- Attributes: Bedrooms, etc.

**Example:**
```bash
make scrape-bayut
```

## Usage

### Running a Scraper

```bash
# Run a specific scraper
go run cmd/scraper/main.go -scraper bayut

# Run all scrapers
go run cmd/scraper/main.go -scraper all

# Or use Makefile
make scrape-bayut
make scrape-all
```

### Environment Variables

Scrapers require database connection settings:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=cost_of_living
DB_SSLMODE=disable
```

## Metrics

Scrapers export Prometheus metrics:

- `scraper_runs_total{scraper, status}` - Total scraper runs
- `scraper_items_scraped_total{scraper}` - Items successfully scraped
- `scraper_errors_total{scraper, error_type}` - Errors encountered
- `scraper_duration_seconds{scraper}` - Scraper execution time

View metrics at: http://localhost:9090 (when Prometheus is running)

## Best Practices

### Rate Limiting

Always respect the target website:
- Use reasonable rate limits (1-2 req/sec)
- Check and follow robots.txt
- Add appropriate User-Agent

### Error Handling

- Log all errors with context
- Increment error metrics
- Don't fail the entire scrape if one item fails
- Handle network timeouts gracefully

### Data Quality

- Validate extracted data before saving
- Set confidence scores appropriately
- Include source URLs for verification
- Add tags for categorization

## Adding a New Scraper

1. Create a new directory: `internal/scrapers/newsource/`

2. Implement the `Scraper` interface:

```go
package newsource

import (
    "context"
    "github.com/adonese/cost-of-living/internal/models"
    "github.com/adonese/cost-of-living/internal/scrapers"
)

type NewSourceScraper struct {
    config scrapers.Config
    // Add your fields
}

func NewNewSourceScraper(config scrapers.Config) *NewSourceScraper {
    return &NewSourceScraper{config: config}
}

func (s *NewSourceScraper) Name() string {
    return "newsource"
}

func (s *NewSourceScraper) CanScrape() bool {
    // Check rate limit
    return true
}

func (s *NewSourceScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
    // Implement scraping logic
    return nil, nil
}
```

3. Add parser functions in `parser.go`

4. Write tests in `*_test.go`

5. Register in `cmd/scraper/main.go`:

```go
newsourceScraper := newsource.NewNewSourceScraper(config)
service.RegisterScraper(newsourceScraper)
```

6. Add Makefile target:

```makefile
scrape-newsource:
	go run cmd/scraper/main.go -scraper newsource
```

## Testing

```bash
# Run unit tests for parsers
go test ./internal/scrapers/bayut/... -v

# Run all scraper tests
go test ./internal/scrapers/... -v
```

## Troubleshooting

### No data scraped

- Check if website structure changed
- Verify CSS selectors are correct
- Check rate limiting isn't blocking requests
- Look at logs for specific errors

### Database connection failed

- Ensure PostgreSQL is running: `docker-compose ps`
- Check environment variables
- Verify database exists: `make migrate`

### Rate limit errors

- Increase delay between requests
- Use fewer concurrent scrapers
- Check if you're respecting robots.txt

## Future Enhancements

- [ ] Add more data sources (Dubizzle, PropertyFinder, etc.)
- [ ] Implement proxy rotation for better reliability
- [ ] Add scraper scheduling via Temporal workflows
- [ ] Implement data deduplication
- [ ] Add web scraping monitoring dashboard
- [ ] Support for different categories (food, transport, etc.)
