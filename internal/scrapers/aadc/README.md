# AADC Scraper

Scraper for Abu Dhabi Distribution Company (AADC) utility rates - electricity and water tariffs for Abu Dhabi.

## Overview

The AADC scraper extracts residential electricity and water rates from the official AADC tariff page. AADC is the primary utility provider for Abu Dhabi and uses a tiered rate structure with different pricing for UAE nationals and expatriates.

## Data Sources

- **Source URL**: https://www.aadc.ae/en/pages/maintarrif.aspx
- **Update Frequency**: Weekly (rates change infrequently)
- **Data Quality**: High (0.98 confidence) - official government source

## Rate Structure

### Electricity Rates

#### UAE Nationals
- **Up to 30,000 kWh/month**: 5.8 fils/kWh
- **Above 30,000 kWh/month**: 6.7 fils/kWh

#### Expatriates (8 tiers)
- **Up to 400 kWh/month**: 6.7 fils/kWh
- **401-700 kWh/month**: 7.6 fils/kWh
- **701-1,000 kWh/month**: 9.5 fils/kWh
- **1,001-2,000 kWh/month**: 11.5 fils/kWh
- **2,001-3,000 kWh/month**: 17.2 fils/kWh
- **3,001-4,000 kWh/month**: 20.6 fils/kWh
- **4,001-10,000 kWh/month**: 26.8 fils/kWh
- **Above 10,000 kWh/month**: 28.7 fils/kWh

### Water Rates

- **UAE Nationals**: AED 2.09 per 1,000 Imperial Gallons
- **Expatriates**: AED 8.55 per 1,000 Imperial Gallons

## Data Output

Each rate is stored as a separate `CostDataPoint` with the following structure:

### Electricity Example
```json
{
  "category": "Utilities",
  "sub_category": "Electricity",
  "item_name": "AADC Electricity Tier Up to 400 kWh - Expatriate",
  "price": 0.067,
  "unit": "AED per kWh",
  "location": {
    "emirate": "Abu Dhabi",
    "city": "Abu Dhabi"
  },
  "source": "aadc_official",
  "confidence": 0.98,
  "attributes": {
    "customer_type": "expatriate",
    "rate_type": "tiered",
    "fils_rate": 6.7,
    "tier_min_kwh": 0,
    "tier_max_kwh": 400,
    "unit": "fils_per_kwh"
  },
  "tags": ["electricity", "utility", "aadc", "expatriate"]
}
```

### Water Example
```json
{
  "category": "Utilities",
  "sub_category": "Water",
  "item_name": "AADC Water - National",
  "price": 2.09,
  "unit": "AED per 1000 IG",
  "location": {
    "emirate": "Abu Dhabi",
    "city": "Abu Dhabi"
  },
  "source": "aadc_official",
  "confidence": 0.98,
  "attributes": {
    "customer_type": "national",
    "rate_type": "flat",
    "unit": "aed_per_1000_ig",
    "volume_unit": "imperial_gallons"
  },
  "tags": ["water", "utility", "aadc", "national"]
}
```

## Expected Output

Each scrape should yield approximately **12 data points**:
- 2 electricity tiers for nationals
- 8 electricity tiers for expatriates
- 2 water rates (1 national, 1 expatriate)

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/adonese/cost-of-living/internal/scrapers"
    "github.com/adonese/cost-of-living/internal/scrapers/aadc"
)

func main() {
    config := scrapers.Config{
        UserAgent:  "cost-of-living-scraper/1.0",
        RateLimit:  1,
        Timeout:    30,
        MaxRetries: 3,
    }

    scraper := aadc.NewAADCScraper(config)

    ctx := context.Background()
    dataPoints, err := scraper.Scrape(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Scraped %d rates\n", len(dataPoints))
}
```

### With Custom URL (for testing)

```go
scraper := aadc.NewAADCScraperWithURL(config, "http://localhost:8080/test")
```

## Implementation Details

### Parser Logic

The parser (`parser.go`) handles:

1. **Electricity Rate Parsing**
   - Finds `<section class="electricity-residential">` elements
   - Identifies customer type from `<h3>` headings (Nationals vs Expatriates)
   - Extracts tiers from table rows
   - Parses consumption ranges (e.g., "Up to 400", "401 - 700", "Above 10,000")
   - Converts fils rates to AED (1 AED = 100 fils)

2. **Water Rate Parsing**
   - Finds `<section class="water-residential">` elements
   - Identifies customer type from headings
   - Extracts flat rate per 1,000 Imperial Gallons
   - Stores rates in AED

3. **Tier Name Formatting**
   - "Up to X" tiers: `Tier Up to X kWh`
   - Range tiers: `Tier X-Y kWh`
   - Unlimited tiers: `Tier Above X kWh`

### Rate Conversion

All electricity rates are stored in **AED per kWh**, but the original fils rate is preserved in attributes:
- Display price: `0.067` AED/kWh
- Attribute `fils_rate`: `6.7`

### Customer Type Differentiation

The scraper properly differentiates between:
- **UAE Nationals**: Lower rates, fewer tiers, subsidized
- **Expatriates**: Higher rates, more granular tiers

This is critical for accurate cost calculations as rates can differ by 4x for water.

## Testing

### Run Unit Tests

```bash
go test ./internal/scrapers/aadc/...
```

### Run Integration Tests

```bash
go test ./test/integration -run TestAADC
```

### Check Coverage

```bash
go test -cover ./internal/scrapers/aadc/...
```

Expected coverage: **>80%**

## Fixtures

Test fixtures are available at:
- `test/fixtures/aadc/rates.html` - Sample AADC tariff page

## Workflow Integration

The AADC scraper is integrated with Temporal workflows:

```go
// internal/workflow/aadc_workflow.go
input := AADCScraperWorkflowInput{
    MaxRetries: 3,
    Timeout:    5 * time.Minute,
}
result, err := AADCScraperWorkflow(ctx, input)
```

### Schedule

The workflow runs on a **weekly schedule** via `ScheduledAADCWorkflow()`:
- Utility rates change infrequently (typically quarterly or annually)
- Weekly checks ensure we catch changes promptly
- Failed scrapes trigger alerts

### Validation

The workflow validates data quality:
- Expects minimum 10 data points per scrape
- Sends alerts if data volume is low
- Implements retry logic with exponential backoff

## Known Considerations

### Monthly vs Daily Consumption

AADC uses **monthly consumption** tiers (kWh/month), unlike some other utilities that use daily consumption. This is clearly indicated in:
- Tier naming: "Tier 401-700 kWh" (monthly)
- Attribute storage: `tier_min_kwh`, `tier_max_kwh` (monthly values)

### Imperial Gallons

Water rates are per **1,000 Imperial Gallons (IG)**, not liters:
- 1 Imperial Gallon ≈ 4.546 liters
- 1,000 IG ≈ 4,546 liters
- Stored in attributes as `volume_unit: "imperial_gallons"`

### Additional Fees

AADC charges additional fees not captured by this scraper:
- Sewerage charge (60% of water consumption)
- Distribution network charges
- Service connection charges

Future versions may include these.

## Error Handling

The scraper handles:
- HTTP errors (timeouts, 500s)
- Invalid HTML structure
- Missing rate tables
- Malformed consumption ranges
- Rate limit compliance

## Maintenance

### When AADC Changes Their Website

If the AADC website structure changes:

1. Update test fixture: `test/fixtures/aadc/rates.html`
2. Adjust selectors in `parser.go`:
   - Section classes: `.electricity-residential`, `.water-residential`
   - Table selectors
   - Heading identification
3. Update parser tests in `parser_test.go`
4. Verify integration tests pass

### When Rates Change

Rate changes are automatically detected and stored as new data points with updated `recorded_at` timestamps. No code changes needed.

## Performance

- **Average scrape time**: ~2-3 seconds
- **Page size**: ~50-100 KB
- **Rate limiting**: 1 request/second (configurable)
- **Memory usage**: Minimal (<10 MB)

## Future Enhancements

1. **Additional Charges**: Capture sewerage and distribution fees
2. **Commercial Rates**: Add industrial/commercial tariffs
3. **Historical Tracking**: Store rate change history
4. **Rate Calculator**: Utility to calculate bills based on consumption
5. **Multi-language Support**: Parse Arabic version for validation

## Contributing

When modifying this scraper:
1. Update tests to maintain >80% coverage
2. Verify integration test passes with fixture
3. Update this README with any structural changes
4. Test against live AADC website before deploying

## Support

For issues or questions:
- Check logs for detailed error messages
- Verify AADC website is accessible
- Review test fixtures for expected HTML structure
- Consult workflow execution history in Temporal UI

## License

Part of the UAE Cost of Living project.
