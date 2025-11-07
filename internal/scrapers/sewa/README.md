# SEWA Scraper

Scraper for Sharjah Electricity, Water and Gas Authority (SEWA) utility rates.

## Overview

The SEWA scraper extracts official utility tariff information for Sharjah emirate, including:

- Electricity rates (multiple consumption tiers)
- Water rates (per 1000 gallons)
- Sewerage charges
- Different rates for UAE Nationals vs Expatriates

## Data Source

- **Source**: SEWA Official Website
- **URL**: https://www.sewa.gov.ae/en/content/tariff
- **Type**: Official government rates
- **Update Frequency**: Weekly (rates change infrequently)
- **Confidence**: 0.98 (official source)

## Data Points Extracted

The scraper extracts approximately **10 data points** per scrape:

### Electricity Rates (7 data points)

**Emirati Customers (4 tiers):**
- Tier 1: 1-3,000 kWh @ 14 fils/kWh
- Tier 2: 3,001-5,000 kWh @ 18 fils/kWh
- Tier 3: 5,001-10,000 kWh @ 27.5 fils/kWh
- Tier 4: Above 10,000 kWh @ 32 fils/kWh

**Expatriate Customers (3 tiers):**
- Tier 1: 1-3,000 kWh @ 27.5 fils/kWh
- Tier 2: 3,001-5,000 kWh @ 32 fils/kWh
- Tier 3: Above 5,000 kWh @ 38 fils/kWh

### Water Rates (2 data points)

- Emirati: AED 8.00 per 1000 gallons
- Expatriate: AED 15.00 per 1000 gallons

### Sewerage Charges (1 data point)

- 50% of water consumption charge

## Data Structure

Each data point follows this structure:

```json
{
  "category": "Utilities",
  "sub_category": "Electricity|Water|Sewerage",
  "item_name": "SEWA Electricity (1-3000 kWh) - Emirati",
  "price": 0.14,
  "location": {
    "emirate": "Sharjah",
    "city": "Sharjah"
  },
  "source": "sewa_official",
  "source_url": "https://www.sewa.gov.ae/en/content/tariff",
  "confidence": 0.98,
  "unit": "AED per kWh",
  "attributes": {
    "consumption_range_min": 1,
    "consumption_range_max": 3000,
    "customer_type": "emirati",
    "rate_type": "tier",
    "unit": "fils_per_kwh",
    "rate_fils": 14.0
  }
}
```

## Customer Type Handling

SEWA differentiates rates based on customer nationality:

- **Emirati**: UAE National residents (lower rates)
- **Expatriate**: Foreign residents (higher rates)

This distinction is critical for accurate cost modeling for different user segments.

## Usage

### Basic Usage

```go
import "github.com/adonese/cost-of-living/internal/scrapers/sewa"

config := scrapers.Config{
    UserAgent:  "CostOfLiving/1.0",
    RateLimit:  1,  // 1 request per second
    Timeout:    30, // 30 seconds
    MaxRetries: 3,
}

scraper := sewa.NewSEWAScraper(config)

ctx := context.Background()
dataPoints, err := scraper.Scrape(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Scraped %d utility rates\n", len(dataPoints))
```

### With Workflow

```go
// Run via Temporal workflow
input := workflow.SEWAScraperWorkflowInput{
    MaxRetries: 3,
    Timeout:    5 * time.Minute,
}

result, err := workflow.SEWAScraperWorkflow(ctx, input)
```

## Rate Limiting

- Default: 1 request per second
- SEWA is a government website, so we use conservative rate limiting
- Recommended: Run weekly (rates rarely change)

## Error Handling

The scraper handles the following error scenarios:

1. **Network Errors**: Retried with exponential backoff
2. **HTTP Errors**: Status codes != 200 are treated as failures
3. **Parsing Errors**: Invalid HTML structure triggers an error
4. **Empty Data**: No tariffs found triggers an error
5. **Context Cancellation**: Gracefully stops scraping

## Testing

### Run Unit Tests

```bash
go test ./internal/scrapers/sewa/...
```

### Run with Coverage

```bash
go test -cover ./internal/scrapers/sewa/...
```

### Run Integration Tests

```bash
go test ./test/integration -run TestSEWA
```

## Test Fixtures

Test fixtures are located at:
- `test/fixtures/sewa/tariff_page.html`

These fixtures are created by Agent 1 and replicate the actual SEWA tariff page structure.

## Data Validation

The scraper validates:

1. **Minimum Data Points**: At least 8 data points (allowing for minor variations)
2. **Required Fields**: All data points must have category, price, location, etc.
3. **Price Ranges**: Electricity rates should be between 0.01 and 1.0 AED/kWh
4. **Customer Types**: Must be either "emirati" or "expatriate"

## Anti-Bot Measures

SEWA is a government website with minimal anti-bot protection:

- User-Agent header is sufficient
- No JavaScript rendering required
- No CAPTCHA challenges observed
- Rate limiting is conservative to respect the service

## Monitoring & Alerts

The workflow sends alerts for:

1. **Scraping Failures**: When scraper fails after all retries
2. **Low Data Volume**: When fewer than 8 data points are extracted
3. **Rate Changes**: When significant price changes are detected (future enhancement)

## Maintenance

### HTML Structure Changes

If SEWA updates their website structure:

1. Update selectors in `parser.go`
2. Update test fixture in `test/fixtures/sewa/tariff_page.html`
3. Verify all tests pass
4. Deploy updated scraper

### Rate Changes

When SEWA changes their rates:

1. No code changes needed
2. Scraper automatically extracts new rates
3. Data validation may need adjustment if tier structure changes

## Integration with Workflow

The SEWA scraper is integrated with Temporal workflows:

- **Workflow**: `SEWAScraperWorkflow`
- **Schedule**: Weekly (can be adjusted)
- **Timeout**: 5 minutes
- **Retries**: 3 attempts with exponential backoff

## Dependencies

- `github.com/PuerkitoBio/goquery`: HTML parsing
- `golang.org/x/time/rate`: Rate limiting
- `go.temporal.io/sdk`: Workflow orchestration

## Performance

- **Average Scrape Time**: < 5 seconds
- **Data Volume**: ~10 data points
- **Memory Usage**: < 10 MB
- **Network Bandwidth**: < 100 KB

## Future Enhancements

1. **Historical Rate Tracking**: Track rate changes over time
2. **Rate Change Alerts**: Notify when rates change
3. **Municipality Fees**: Extract additional fee information
4. **Service Fees**: Parse connection and service fees
5. **Multi-language Support**: Extract Arabic tariff information

## Handoff to Agent 10

For workflow integration, use this registration code:

```go
// Register SEWA scraper
registry.Register("sewa", func(config scrapers.Config) scrapers.Scraper {
    return sewa.NewSEWAScraper(config)
})

// Register SEWA workflow
worker.RegisterWorkflow(workflow.SEWAScraperWorkflow)
worker.RegisterWorkflow(workflow.ScheduledSEWAWorkflow)
```

## Support

For issues or questions:
- Check test fixtures for expected HTML structure
- Review parser.go for selector logic
- See integration tests for usage examples

## License

Part of the Cost of Living UAE project.
