# DEWA Scraper

## Overview

The DEWA (Dubai Electricity and Water Authority) scraper extracts official utility rate information for Dubai. It parses electricity and water consumption slabs, as well as fuel surcharges from the DEWA website.

## Data Source

- **Official Website**: https://www.dewa.gov.ae/en/consumer/billing/slab-tariff
- **Confidence Level**: 0.98 (Official government source)
- **Update Frequency**: Weekly (rates rarely change, but checked regularly)

## Data Extracted

### Electricity Rates (4 slabs)
- **Slab 1**: 0-2,000 kWh @ 23.0 fils/kWh
- **Slab 2**: 2,001-4,000 kWh @ 28.0 fils/kWh
- **Slab 3**: 4,001-6,000 kWh @ 32.0 fils/kWh
- **Slab 4**: Above 6,000 kWh @ 38.0 fils/kWh

### Water Rates (3 slabs)
- **Slab 1**: 0-5,000 IG @ 3.57 fils/IG
- **Slab 2**: 5,001-10,000 IG @ 5.24 fils/IG
- **Slab 3**: Above 10,000 IG @ 10.52 fils/IG

### Fuel Surcharge
- Variable rate (currently 6.5 fils/kWh for electricity)

## Data Structure

Each rate slab is stored as a `CostDataPoint` with:

```go
{
    Category:    "Utilities",
    SubCategory: "Electricity" | "Water" | "Fuel Surcharge",
    ItemName:    "DEWA Electricity Slab 0-2000 kWh",
    Price:       0.23,  // Converted from fils to AED
    Location: {
        Emirate: "Dubai",
        City:    "Dubai",
    },
    Source:      "dewa_official",
    SourceURL:   "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff",
    Confidence:  0.98,
    Unit:        "AED",
    Attributes: {
        "consumption_range_min": 0,
        "consumption_range_max": 2000,
        "rate_type":             "slab",
        "unit":                  "fils_per_kwh" | "fils_per_ig",
    },
}
```

## Usage

```go
import (
    "context"
    "github.com/adonese/cost-of-living/internal/scrapers"
    "github.com/adonese/cost-of-living/internal/scrapers/dewa"
)

// Create scraper
config := scrapers.Config{
    UserAgent:  "cost-of-living-bot/1.0",
    RateLimit:  1,  // 1 request per second
    Timeout:    30, // 30 seconds
    MaxRetries: 3,
}

scraper := dewa.NewDEWAScraper(config)

// Scrape data
ctx := context.Background()
dataPoints, err := scraper.Scrape(ctx)
if err != nil {
    // Handle error
}

// Process data points (typically 7-8 items)
for _, dp := range dataPoints {
    // Save to database, etc.
}
```

## Implementation Details

### Parser Components

- **parseElectricitySlabs**: Extracts electricity rate slabs from HTML table
- **parseWaterSlabs**: Extracts water rate slabs from HTML table
- **parseFuelSurcharge**: Extracts fuel surcharge from note text
- **parseConsumptionRange**: Parses consumption ranges like "0 - 2,000" or "Above 6,000"
- **extractRate**: Extracts rate values from text like "23.0 fils"
- **slabToDataPoint**: Converts parsed slab to CostDataPoint

### Rate Conversion

- All rates are converted from fils to AED (100 fils = 1 AED)
- Example: 23.0 fils/kWh becomes 0.23 AED/kWh

### Error Handling

- Returns error if no electricity or water slabs found
- Logs warnings for missing data but continues processing
- Supports context cancellation for graceful shutdown
- Includes retry logic via Temporal workflows

## Testing

Run tests with:

```bash
# Unit tests
go test ./internal/scrapers/dewa/...

# With coverage
go test ./internal/scrapers/dewa/... -cover

# Integration tests with fixtures
go test ./internal/scrapers/dewa/... -v -run Integration
```

Test coverage: **81.1%**

## Fixtures

Test fixtures are located at:
- `test/fixtures/dewa/rates_table.html` - Sample DEWA rates page

## Workflow Integration

The DEWA scraper integrates with Temporal workflows:

```go
// In your worker registration
scrapers := map[string]scrapers.Scraper{
    "dewa": dewa.NewDEWAScraper(config),
}

// Execute via workflow
input := workflow.ScraperWorkflowInput{
    ScraperName: "dewa",
    MaxRetries:  3,
}
```

## Notes

- DEWA rates are residential rates only
- Commercial and industrial rates are different and not included
- Additional charges (housing fees, sewerage, municipality) are parsed but not included in cost data points
- Rates are specific to Dubai emirate only
- The scraper respects rate limits to avoid overloading DEWA servers

## Future Enhancements

- [ ] Add support for commercial rates
- [ ] Parse additional charges as separate data points
- [ ] Add historical rate tracking
- [ ] Support multiple languages (currently English only)
- [ ] Add rate comparison features
