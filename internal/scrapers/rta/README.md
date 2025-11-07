# RTA Scraper

Comprehensive scraper for Dubai Roads and Transport Authority (RTA) public transport fare information.

## Overview

The RTA scraper extracts official fare data for Dubai's public transportation system, including:
- Dubai Metro (Red and Green lines)
- Dubai Bus
- Dubai Tram
- Dubai Taxi

## Features

- **Zone-based Pricing**: Supports Dubai Metro's 7-zone fare system
- **Card Type Differentiation**: Extracts fares for different Nol card types (Silver, Gold, Blue, Red)
- **Multiple Transport Modes**: Covers metro, bus, tram, and taxi fares
- **Day Pass Options**: Includes day pass pricing for public transport
- **Official Source**: Scrapes from RTA's official website with high confidence (0.95)

## Data Extracted

### Metro Fares
- Regular fare (default single journey)
- Zone-based fares (1 zone, 2 zones, all zones)
- Card type variations (Silver Standard, Gold Premium)
- Day pass options

### Bus Fares
- Zone-based single journey fares
- Day pass option

### Tram Fares
- Single journey (Silver and Gold)
- Day pass (Silver and Gold)

### Taxi Fares
- Day flag down (6 AM - 10 PM)
- Night flag down (10 PM - 6 AM)
- Airport pickup
- Per kilometer rate
- Waiting time rate
- Minimum fare

## Output Structure

Each fare is extracted as a `CostDataPoint` with:

```go
{
    Category: "Transportation",
    SubCategory: "Public Transport" | "Taxi",
    ItemName: "Dubai Metro 1 Zone - Silver Card (Standard)",
    Price: 3.0,  // AED
    Location: {Emirate: "Dubai", City: "Dubai"},
    Source: "rta_official",
    SourceURL: "https://www.rta.ae/...",
    Confidence: 0.95,
    Unit: "AED",
    Attributes: {
        "transport_mode": "metro",
        "card_type": "silver",
        "fare_type": "single_journey",
        "zones_crossed": 1
    }
}
```

## Zone System

Dubai Metro operates on a 7-zone fare system:
- Zone 1-2: Short distances within a single area
- Zone 2-5: Medium distances crossing areas
- Zone 1-7: All zones (longest journeys)

Fares increase with the number of zones crossed.

## Card Types

### Silver Card (Standard)
- Standard class metro and tram
- All bus services
- Base fare pricing

### Gold Card (Premium)
- First class metro and tram
- Typically 2x Silver card price
- Premium comfort and less crowding

### Blue Card (Concession)
- Students and seniors
- People of determination
- Discounted fares (not always shown in public fare tables)

### Red Ticket (Single Use)
- Disposable paper ticket
- Single journey or day pass
- Slightly higher per-journey cost

## Usage

```go
import (
    "context"
    "github.com/adonese/cost-of-living/internal/scrapers"
    "github.com/adonese/cost-of-living/internal/scrapers/rta"
)

// Create scraper
config := scrapers.Config{
    UserAgent:  "MyApp/1.0",
    RateLimit:  1,
    Timeout:    30,
    MaxRetries: 3,
}

scraper := rta.NewRTAScraper(config)

// Scrape with retry logic
ctx := context.Background()
dataPoints, err := scraper.ScrapeWithRetry(ctx)
if err != nil {
    log.Fatal(err)
}

// Validate data
err = rta.ValidateFareData(dataPoints)
if err != nil {
    log.Printf("Validation warning: %v", err)
}

// Process results
for _, dp := range dataPoints {
    log.Printf("%s: AED %.2f", dp.ItemName, dp.Price)
}
```

## Expected Output

Typical scrape extracts 25-30 data points:
- Metro: 9-10 fares (Regular + zones × card types + day pass)
- Bus: 4 fares (3 zone-based + 1 day pass)
- Tram: 4 fares (2 single journey + 2 day pass)
- Taxi: 6 fares (flag down, per km, waiting, minimum, airport)

## Workflow Integration

The RTA scraper is designed to run on a weekly schedule as fares change infrequently:

```go
// Execute via Temporal workflow
input := workflow.RTAWorkflowInput{
    MaxRetries: 3,
}

result, err := workflow.RTAScraperWorkflow(ctx, input)
```

Fares typically update:
- Quarterly: Minor adjustments
- Annually: Major fare structure changes
- Ad-hoc: During major transport expansions

## Testing

### Unit Tests
```bash
go test ./internal/scrapers/rta/
```

### Integration Tests
```bash
go test ./test/integration -run TestRTA
```

### Coverage
```bash
go test -cover ./internal/scrapers/rta/
```

Target: >80% code coverage

## Test Fixtures

Located at: `test/fixtures/rta/fare_calculator.html`

Contains representative HTML structure from RTA's fare calculator page with:
- Complete metro fare table
- Bus fare information
- Tram fare details
- Taxi rate structure
- Nol card information

## Implementation Notes

### Parser Strategy
1. **Section-based parsing**: Identifies fare sections by CSS selectors and headings
2. **Table-based extraction**: Parses HTML tables for structured fare data
3. **Multi-selector fallback**: Tries multiple selectors to handle layout changes
4. **Flexible price parsing**: Handles various currency formats (AED, Dhs, numbers only)

### Robustness
- Multiple CSS selector strategies for each fare type
- Graceful degradation if sections are missing
- Validation of minimum expected data points
- Retry logic with exponential backoff

### Maintenance
- Fares rarely change (typically quarterly or annually)
- HTML structure is stable (official government site)
- Parsers have fallback strategies for resilience

## Data Quality

### Confidence Levels
- **0.95**: High confidence for all RTA fares (official source)
- Consistent with other utility scrapers (DEWA, SEWA)

### Validation
Built-in validation ensures:
- All prices are positive
- Category is "Transportation"
- Location is "Dubai"
- Minimum expected fares are present (especially metro)
- Gold card is approximately 2x Silver card for metro

### Known Limitations
1. Blue card (concession) fares not always publicly listed
2. Zone calculations require station-to-zone mapping (not implemented)
3. Peak/off-peak pricing not differentiated (if applicable)
4. Special event pricing not captured

## Handoff for Agent 10 (Workflow Integration)

### Integration Points
1. **Scraper Registration**: Register "rta" scraper in scraper factory
2. **Workflow Scheduling**: Add RTA workflow to weekly schedule
3. **Zone Logic**: Zone calculation logic available in `zones.go`
4. **Validation**: Use `ValidateFareData()` for data quality checks

### Card Type Benefits
- Silver: Standard pricing, most economical
- Gold: 2x price, premium comfort, less crowding
- Blue: Concession rates for students/seniors
- Red: Single-use, convenient for tourists

### Important Functions
- `GetMetroFares()`: Returns standard metro fare structure
- `GetBusFares()`: Returns bus fare structure
- `GetTramFares()`: Returns tram fare structure
- `GetTaxiFares()`: Returns taxi fare structure
- `FormatItemName()`: Standardizes fare item names
- `ValidateFareData()`: Validates extracted data quality

## Related Documentation

- Project root: `/home/adonese/src/cost-of-living`
- Scraper interface: `internal/scrapers/scraper.go`
- Data models: `internal/models/cost_data_point.go`
- Workflow: `internal/workflow/rta_workflow.go`
- Integration tests: `test/integration/rta_integration_test.go`

## References

- RTA Official Website: https://www.rta.ae
- Nol Card Information: https://nol.ae
- Dubai Metro Map: Available on RTA website
- Fare Calculator: RTA mobile app and website

## Version History

- **v1.0** (2025-11-07): Initial implementation
  - Complete metro, bus, tram, and taxi fare extraction
  - Zone-based pricing support
  - Multiple card type differentiation
  - Comprehensive test coverage (>80%)
  - Integration with Temporal workflows

## Contact

For issues or questions about the RTA scraper, refer to the main project documentation or raise an issue in the project repository.
