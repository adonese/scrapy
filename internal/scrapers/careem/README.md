# Careem Ride-Sharing Rates Scraper

## Overview

The Careem scraper collects ride-sharing rates for Careem services across UAE emirates. Unlike traditional scrapers, this implementation uses a **multi-source aggregation strategy** since Careem doesn't provide a public API for rate information.

## Architecture

### Multi-Source Strategy

Due to the lack of an official API, the scraper implements a fallback strategy across multiple sources:

```
Priority 1: Official API (if available)
    ↓ (not available)
Priority 2: Careem Help Center
    ↓ (parsing complex)
Priority 3: News Articles & Press Releases
    ↓ (unreliable)
Priority 4: Static Fixture (fallback)
```

### Components

- **careem.go** - Main scraper with source aggregation
- **sources.go** - RateSource interface and implementations
- **parser.go** - Rate parsing and data point conversion
- **careem_test.go** - Unit tests
- **parser_test.go** - Parser tests

## Rate Components

The scraper extracts the following rate components:

1. **Base Fare** - Starting fare for any trip
2. **Per Kilometer Rate** - Charge per kilometer traveled
3. **Per Minute Wait** - Charge per minute of waiting
4. **Minimum Fare** - Minimum charge for any trip
5. **Peak Surcharge** - Multiplier during peak hours (7-9 AM, 5-8 PM weekdays)
6. **Airport Surcharge** - Additional charge for airport pickups (DXB, DWC)
7. **Salik Toll** - Per toll gate crossing charge

### Service Types

- **careem_go** - Economy service (default)
- **careem_go_plus** - Premium service
- **careem_comfort** - Comfortable ride option

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/adonese/cost-of-living/internal/scrapers"
    "github.com/adonese/cost-of-living/internal/scrapers/careem"
)

config := scrapers.Config{
    UserAgent:  "Mozilla/5.0 (compatible; Bot/1.0)",
    RateLimit:  2,
    Timeout:    30,
    MaxRetries: 3,
}

scraper := careem.NewCareemScraper(config)
dataPoints, err := scraper.Scrape(context.Background())
```

### With Custom Sources

```go
// Use only static source for testing
staticSource := careem.NewStaticSource("test/fixtures/careem/rates.json")
scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{
    staticSource,
})
```

### Fare Estimation

```go
// After scraping
fare, err := scraper.EstimateFare(
    10.0,  // distance in km
    5.0,   // wait time in minutes
    true,  // is peak hour
    false, // is airport trip
    2,     // number of Salik gates
)
// Returns estimated fare in AED
```

### Rate Summary

```go
summary, err := scraper.GetRatesSummary()
fmt.Println(summary)
// Outputs formatted rate information
```

## Data Structure

### CareemRates

```go
type CareemRates struct {
    ServiceType              string
    Emirate                  string
    BaseFare                 float64
    PerKm                    float64
    PerMinuteWait            float64
    MinimumFare              float64
    PeakSurchargeMultiplier  float64
    AirportSurcharge         float64
    SalikToll                float64
    EffectiveDate            string
    Source                   string
    LastUpdated              time.Time
    Confidence               float32
    Rates                    []ServiceRate
}
```

### Output Format

Each rate component is converted to a `CostDataPoint`:

```go
{
    Category:    "Transportation",
    SubCategory: "Ride Sharing",
    ItemName:    "Base Fare",
    Price:       8.0,
    Location:    {Emirate: "Dubai", City: "Dubai"},
    Source:      "careem_rates",
    Confidence:  0.75,
    Unit:        "AED",
    Tags:        ["careem", "ride_sharing", "base_fare"],
    Attributes: {
        "rate_type":      "base_fare",
        "service":        "careem_go",
        "effective_date": "2025-01-01",
        "data_source":    "static_file"
    }
}
```

## Rate Sources

### 1. Static Source (Default Fallback)

Reads from JSON fixture file with known rates.

```go
source := careem.NewStaticSource("path/to/rates.json")
```

**Confidence:** 0.7 (lower due to potential staleness)

### 2. Help Center Source

Attempts to scrape Careem help center pages.

```go
source := careem.NewHelpCenterSource(userAgent)
```

**Confidence:** 0.85 (if parsing succeeds)

### 3. News Source

Searches news articles for rate information.

```go
source := careem.NewNewsSource(userAgent)
```

**Confidence:** 0.75 (variable quality)

### 4. API Source

Would use official API if available.

```go
source := careem.NewAPISource(apiKey)
```

**Confidence:** 0.95 (currently not available)

## Validation

Rates are validated for:

- **Positivity:** All rates must be > 0
- **Consistency:** Minimum fare >= Base fare
- **Range checks:** Rates within reasonable bounds for UAE
  - Base fare < 100 AED
  - Per km < 10 AED
  - Minimum fare < 200 AED
- **Peak surcharge:** Must be >= 1.0 if set

## Rate Change Detection

The scraper monitors for significant rate changes (>10% by default):

```go
changes := careem.DetectRateChange(oldRates, newRates, 10.0)
// Returns array of change descriptions
```

Logs warnings when significant changes are detected.

## Testing

### Run Unit Tests

```bash
go test ./internal/scrapers/careem/...
```

### Run Integration Tests

```bash
go test ./test/integration -run TestCareem
```

### Check Coverage

```bash
go test -cover ./internal/scrapers/careem/...
```

## Workflow Integration

### Monthly Execution

Careem rates change infrequently, so the workflow runs monthly:

```go
workflow.ExecuteChildWorkflow(ctx, CareemScraperWorkflow, input)
```

### Validation Workflow

Separate workflow validates data freshness and quality:

```go
workflow.ExecuteChildWorkflow(ctx, ValidateCareemRatesWorkflow)
```

## Configuration

Recommended settings:

```go
Config{
    UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
    RateLimit:  2,    // 2 requests per second
    Timeout:    30,   // 30 seconds
    MaxRetries: 5,    // More retries due to multiple sources
}
```

## Unique Challenges

1. **No Official API** - Requires creative sourcing
2. **Rate Staleness** - Rates may not update frequently
3. **Lower Confidence** - Due to unofficial sources (0.7-0.85 vs 0.9+ for official sources)
4. **Multiple Sources** - Need fallback and aggregation logic
5. **Validation Complexity** - Must verify rates are reasonable

## Monitoring Requirements

### Data Freshness
- Alert if data > 60 days old
- Automatic refresh if > 90 days old

### Confidence Scores
- Monitor average confidence
- Alert if confidence drops below 0.6

### Rate Changes
- Log all changes > 10%
- Alert on changes > 20%

### Source Availability
- Track which sources are working
- Alert if all sources fail

## Handoff Notes for Agent 10

### Integration Requirements

1. **Scraper Registration:**
   ```go
   registry.Register("careem", func(cfg scrapers.Config) scrapers.Scraper {
       return careem.NewCareemScraper(cfg)
   })
   ```

2. **Workflow Registration:**
   ```go
   worker.RegisterWorkflow(workflow.CareemScraperWorkflow)
   worker.RegisterWorkflow(workflow.ScheduledCareemWorkflow)
   worker.RegisterWorkflow(workflow.ValidateCareemRatesWorkflow)
   ```

3. **Schedule Configuration:**
   ```go
   Schedule: "0 0 1 * *"  // First day of each month
   ```

### Special Considerations

- **Longer Timeout:** Use 10-minute timeout (tries multiple sources)
- **More Retries:** Configure 5 retries instead of standard 3
- **Lower Confidence:** Expect 0.7-0.85 confidence scores
- **Validation:** Run validation workflow weekly
- **Monitoring:** Set up alerts for data staleness

## Future Improvements

1. **Web Scraping Enhancement**
   - Implement robust Help Center scraping
   - Add OCR for rate screenshots
   - Crowdsource rate verification

2. **API Integration**
   - Monitor for official API availability
   - Implement when available

3. **Machine Learning**
   - Predict rate changes
   - Validate rates against patterns

4. **User Contributions**
   - Allow verified users to submit rates
   - Cross-validate user submissions

## Example Output

```
Careem Rates Summary
=====================
Service: careem_go
Emirate: Dubai
Source: static_file (Confidence: 70.0%)
Effective Date: 2025-01-01

Base Rates:
- Base Fare: 8.00 AED
- Per Kilometer: 1.97 AED
- Per Minute Wait: 0.50 AED
- Minimum Fare: 12.00 AED

Surcharges:
- Peak Hour Multiplier: 1.50x
- Airport Surcharge: 20.00 AED
- Salik Toll (per gate): 5.00 AED

Last Updated: 2025-01-07 12:30:45
```

## Contact & Support

For issues or questions about the Careem scraper:
- Check test files for usage examples
- Review integration tests for edge cases
- See orchestration document for project status
