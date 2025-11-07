# Workflow Implementation Plan - UAE Cost of Living
## Temporal-First, Zero Hardcoding, Continuous Data Collection

**Date:** November 7, 2025
**Architecture:** Workflow-centric, agent-based data collection
**Core Principle:** Every data source = a Temporal workflow

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Temporal Scheduler                        │
│  (Orchestrates all workflows based on frequency)            │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┬─────────────────┐
        ▼              ▼              ▼                 ▼
   [Daily        [Weekly         [Monthly         [On-Demand
   Workflows]    Workflows]      Workflows]       Workflows]
        │              │              │                 │
        ├─ bayut       ├─ dewa_rates ├─ edarabia      ├─ calculator
        ├─ dubizzle    ├─ sewa_rates ├─ bayanat       └─ enrichment
        ├─ rentit      ├─ carrefour  └─ dsc_stats
        ├─ homebook    └─ lulu
        └─ ewaar
                       │
                       ▼
              ┌────────────────────┐
              │  CostDataPoint DB  │
              │  (TimescaleDB)     │
              └────────────────────┘
```

---

## 📦 Implementation Batches (Dependency-Based)

### Batch 1: Foundation Extension
**Goal:** Multi-emirate coverage with existing infrastructure
**Dependencies:** None (extends current code)

#### 1.1 Multi-Emirate Housing Workflows
```go
// Extend existing scrapers to new emirates

Workflows to create:
├─ bayut_sharjah_workflow.go
├─ bayut_ajman_workflow.go
├─ bayut_abudhabi_workflow.go
├─ dubizzle_sharjah_workflow.go
└─ dubizzle_shared_workflow.go

Implementation:
- Reuse internal/scrapers/bayut/bayut.go structure
- Add URL configurations for each emirate
- Same parser logic, different base URLs
- Register workflows in worker
```

**Estimated Effort:** 3-4 days
**Files to Modify:**
- `internal/scrapers/bayut/bayut.go` (add emirate parameter)
- `internal/scrapers/dubizzle/dubizzle.go` (add emirate + shared category)
- `internal/workflow/scraper_workflow.go` (register new workflows)
- `cmd/worker/main.go` (register workflow definitions)

**Data Output:**
```go
Category: "Housing"
SubCategory: "Rent" or "Shared Accommodation"
Location: {Emirate: "Sharjah"/"Ajman"/"Abu Dhabi"}
Source: "bayut" or "dubizzle"
```

---

### Batch 2: Official Rates Scrapers (Simple HTML)
**Goal:** Scrape government utility/transport rate tables
**Dependencies:** None (new scrapers)

#### 2.1 Utility Rates Workflows

**DEWA Rates Scraper**
```go
// internal/scrapers/dewa/dewa_rates.go

Target URL: https://www.dewa.gov.ae/en/consumer/billing/slab-tariff
Method: HTTP GET + goquery (HTML parsing)
Frequency: Weekly (rates change infrequently, but verify weekly)

Data to Extract:
- Electricity slabs (consumption ranges + fils/kWh)
- Water slabs (consumption ranges + AED/m³)
- Fuel surcharge
- Housing fee (5% of rent)

Output Structure:
{
  Category: "Utilities",
  SubCategory: "Electricity",
  ItemName: "DEWA Electricity Slab 1 (0-2000 kWh)",
  Price: 0.23,  // 23 fils per kWh
  Location: {Emirate: "Dubai"},
  Source: "dewa_official",
  SourceURL: "https://...",
  Confidence: 0.98,
  Attributes: {
    "consumption_range_min": 0,
    "consumption_range_max": 2000,
    "rate_type": "slab",
    "unit": "fils_per_kwh"
  }
}
```

**SEWA Rates Scraper**
```go
// internal/scrapers/sewa/sewa_rates.go

Target URL: https://www.sewa.gov.ae/en/energy-calculator
Method: HTTP GET + goquery
Frequency: Weekly

Data to Extract:
- Electricity tiers (20-30 fils/kWh by consumption)
- Water tiers (AED 3-4 per gallon by consumption)
- Sewerage fees (1.5 fils/gallon as of 2025)

Output Structure:
{
  Category: "Utilities",
  SubCategory: "Electricity",
  ItemName: "SEWA Electricity Tier 1 (0-2000 kWh)",
  Price: 0.20,
  Location: {Emirate: "Sharjah"},
  Source: "sewa_official",
  Confidence: 0.98
}
```

**AADC Rates Scraper**
```go
// internal/scrapers/aadc/aadc_rates.go

Target URL: https://www.aadc.ae/en/pages/maintarrif.aspx
Method: HTTP GET + goquery
Frequency: Weekly

Data to Extract:
- Green band electricity rates (26.8 fils/kWh)
- Red band electricity rates (30.5 fils/kWh)
- Water rates (tiered by daily consumption)
- Band thresholds (20 kWh/day for electricity)

Output Structure:
{
  Category: "Utilities",
  SubCategory: "Electricity",
  ItemName: "AADC Electricity Green Band - Expatriate",
  Price: 0.268,
  Location: {Emirate: "Abu Dhabi"},
  Source: "aadc_official",
  Attributes: {
    "band": "green",
    "daily_limit": 20,
    "customer_type": "expatriate"
  }
}
```

#### 2.2 Transportation Rates Workflows

**RTA Fares Scraper**
```go
// internal/scrapers/rta/rta_fares.go

Target URLs:
- https://www.dubaimetrorails.com/fare-calculator
- https://www.rta.ae (check for official API or data)

Method: HTTP GET + HTML parsing OR API if available
Frequency: Weekly (fares change rarely, but announcements happen)

Data to Extract:
- Metro fares by zone (7 zones)
- Card types: Silver, Gold, Blue (student/senior)
- Bus fares
- Day passes
- Tram fares
- Water bus fares

Output Structure:
{
  Category: "Transportation",
  SubCategory: "Public Transport",
  ItemName: "Dubai Metro 1-Zone Fare - Silver Card",
  Price: 3.0,
  Location: {Emirate: "Dubai", City: "Dubai"},
  Source: "rta_official",
  Attributes: {
    "zones_crossed": 1,
    "card_type": "silver",
    "transport_mode": "metro"
  }
}
```

**Careem Rates Scraper**
```go
// internal/scrapers/careem/careem_rates.go

Target URLs:
- https://www.thenationalnews.com (search for Careem rate announcements)
- Careem blog/press releases
- Alternative: Careem help pages

Method: Web scraping news sites + Careem official announcements
Frequency: Monthly (check for rate changes)
Challenge: No structured API, need to monitor announcements

Data to Extract:
- Base fare
- Per km rate
- Per minute waiting rate
- Peak hour surcharges
- Salik toll charges

Output Structure:
{
  Category: "Transportation",
  SubCategory: "Ride Sharing",
  ItemName: "Careem Base Fare",
  Price: 13.0,
  Location: {Emirate: "Dubai"},
  Source: "careem_official_announcement",
  RecordedAt: time.Now(),
  ValidFrom: "2025-11-05",  // when rate became effective
  Attributes: {
    "rate_type": "base_fare",
    "per_km": 2.26,
    "per_minute_wait": 0.5,
    "salik_toll": 5.0
  }
}
```

**Estimated Effort:** 5-6 days
**Files to Create:**
- `internal/scrapers/dewa/dewa_rates.go`
- `internal/scrapers/sewa/sewa_rates.go`
- `internal/scrapers/aadc/aadc_rates.go`
- `internal/scrapers/rta/rta_fares.go`
- `internal/scrapers/careem/careem_rates.go`
- `internal/workflow/utility_rates_workflow.go`
- `internal/workflow/transport_rates_workflow.go`

---

### Batch 3: Shared Accommodations (Simple Sites)
**Goal:** Budget housing data (bed spaces, partitions)
**Dependencies:** None (new scrapers, simple HTML)

#### 3.1 RentItOnline Scraper
```go
// internal/scrapers/rentit/rentit.go

Target URL: https://rentitonline.ae/room-rental/dubai
Method: HTTP GET + goquery
Frequency: Daily
Complexity: LOW (appears to be simple HTML listings)

Data to Extract:
- Price (monthly)
- Accommodation type (bed space, partition, private room)
- Location (area, neighborhood)
- Gender (male/female/mixed)
- Amenities (wifi, AC, kitchen, etc.)
- Available from date

Output Structure:
{
  Category: "Housing",
  SubCategory: "Shared Accommodation",
  ItemName: "Bed Space - Deira - Male",
  Price: 650.0,
  Location: {Emirate: "Dubai", Area: "Deira"},
  Source: "rentit",
  Attributes: {
    "accommodation_type": "bed_space",
    "gender": "male",
    "amenities": ["wifi", "kitchen_shared", "washing_machine"],
    "furnishing": "furnished"
  },
  Tags: ["shared", "budget", "bed-space"]
}
```

#### 3.2 Homebook Scraper
```go
// internal/scrapers/homebook/homebook.go

Target URL: https://homebook.ae/bed-space/
Method: HTTP GET + goquery
Frequency: Daily
Complexity: LOW

Data to Extract:
- Price (starting from AED 600/month)
- Location
- Room type
- Gender preference
- Amenities
- "No agent, no hidden fees" tag

Similar output structure to RentItOnline
```

#### 3.3 Ewaar Scraper
```go
// internal/scrapers/ewaar/ewaar.go

Target URL: https://dubai.ewaar.com/flatmates-bed-space-for-rent-in-dubai/
Method: HTTP GET + goquery
Frequency: Daily
Complexity: MEDIUM (check site structure)

Data to Extract:
- Bed space listings
- Partitions
- Shared rooms
- Locations (Deira, International City, Karama)
```

**Estimated Effort:** 6-7 days
**Files to Create:**
- `internal/scrapers/rentit/rentit.go`
- `internal/scrapers/rentit/parser.go`
- `internal/scrapers/homebook/homebook.go`
- `internal/scrapers/homebook/parser.go`
- `internal/scrapers/ewaar/ewaar.go`
- `internal/scrapers/ewaar/parser.go`
- `internal/workflow/shared_accommodation_workflow.go`

---

### Batch 4: Education Data (Medium Complexity)
**Goal:** School fees across UAE
**Dependencies:** None (new scrapers)

#### 4.1 Edarabia Schools Scraper
```go
// internal/scrapers/edarabia/edarabia_schools.go

Target URL: https://www.edarabia.com/dubai-school-fees/
Method: HTTP GET + goquery (appears to be HTML tables)
Frequency: Monthly (fees change yearly, but listings update monthly)
Complexity: MEDIUM (tables to parse, multiple pages)

Data to Extract:
- School name
- Fees by grade (FS1, Grade 1-12)
- Curriculum (British, American, IB, etc.)
- KHDA rating (Outstanding, Very Good, Good, etc.)
- Location (emirate, area)
- Contact info

Output Structure:
{
  Category: "Education",
  SubCategory: "School Fees",
  ItemName: "GEMS Wellington Primary - Grade 1",
  Price: 45000.0,  // annual fee
  Location: {Emirate: "Dubai", Area: "Al Satwa"},
  Source: "edarabia",
  Attributes: {
    "school_name": "GEMS Wellington Primary School",
    "grade": "1",
    "curriculum": "British",
    "khda_rating": "Outstanding",
    "age_range": "3-11",
    "academic_year": "2025-2026"
  },
  Tags: ["education", "school", "british-curriculum", "outstanding"]
}
```

#### 4.2 SchoolsCompared Scraper (Optional)
```go
// internal/scrapers/schoolscompared/schoolscompared.go

Target URL: https://schoolscompared.com/
Method: Browser automation (Playwright) - may require login/membership
Frequency: Monthly
Complexity: HIGH (potential paywall, complex JS site)

Data to Extract:
- More detailed school profiles
- Teacher ratings
- Facilities information
- Parent reviews

Note: May defer this if paywall/membership required
```

**Estimated Effort:** 5-6 days (Edarabia only), +3 days if SchoolsCompared
**Files to Create:**
- `internal/scrapers/edarabia/edarabia_schools.go`
- `internal/scrapers/edarabia/parser.go`
- `internal/workflow/education_workflow.go`

---

### Batch 5: Carpooling Services (Simple Sites)
**Goal:** Commute cost alternatives
**Dependencies:** None

#### 5.1 Carpooling Scrapers
```go
// internal/scrapers/carlift/carlift.go

Target URLs:
- https://carpoolsuae.com/
- https://mrbus.ae/car-lift-prices/
- https://dubaicarlift.com/

Method: HTTP GET + goquery
Frequency: Weekly
Complexity: LOW-MEDIUM

Data to Extract:
- Routes (Dubai-Abu Dhabi, Dubai-Sharjah, etc.)
- Pricing (daily, weekly, monthly)
- Seat type (shared, private)
- Service provider

Output Structure:
{
  Category: "Transportation",
  SubCategory: "Carpooling",
  ItemName: "Dubai to Abu Dhabi - Monthly Shared",
  Price: 800.0,
  Location: {Emirate: "Dubai"},
  Source: "carpools_uae",
  Attributes: {
    "route_from": "Dubai",
    "route_to": "Abu Dhabi",
    "frequency": "monthly",
    "seat_type": "shared",
    "provider": "carpools_uae"
  },
  Tags: ["carpooling", "commute", "budget"]
}
```

**Estimated Effort:** 3-4 days
**Files to Create:**
- `internal/scrapers/carlift/carlift.go`
- `internal/scrapers/carlift/parser.go`

---

### Batch 6: Browser Automation Infrastructure
**Goal:** Set up Playwright for complex sites
**Dependencies:** None (infrastructure)

#### 6.1 Playwright Integration
```go
// pkg/browser/playwright.go

Setup:
1. Install Playwright for Go
   - github.com/playwright-community/playwright-go

2. Create browser manager:
   - Browser pool for concurrent scraping
   - Headless mode configuration
   - User-Agent rotation
   - Cookie/session management

3. Create reusable browser context:
   - Screenshot capability (for debugging)
   - Network interception (optional)
   - Wait strategies (wait for selectors, network idle)

Implementation:
package browser

import (
    "github.com/playwright-community/playwright-go"
)

type Manager struct {
    pw      *playwright.Playwright
    browser playwright.Browser
}

func NewManager() (*Manager, error) {
    pw, err := playwright.Run()
    if err != nil {
        return nil, err
    }

    browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(true),
    })
    if err != nil {
        return nil, err
    }

    return &Manager{
        pw:      pw,
        browser: browser,
    }, nil
}

func (m *Manager) NewPage() (playwright.Page, error) {
    context, err := m.browser.NewContext(playwright.BrowserNewContextOptions{
        UserAgent: playwright.String("Mozilla/5.0 ..."),
    })
    if err != nil {
        return nil, err
    }

    return context.NewPage()
}

func (m *Manager) Close() error {
    m.browser.Close()
    return m.pw.Stop()
}
```

**Estimated Effort:** 2-3 days
**Files to Create:**
- `pkg/browser/playwright.go`
- `pkg/browser/manager.go`
- `pkg/browser/helpers.go` (wait strategies, selectors)
- `go.mod` updates (add playwright dependency)

---

### Batch 7: Groceries (Browser Automation Required)
**Goal:** Track grocery prices at major chains
**Dependencies:** Batch 6 (Playwright infrastructure)

#### 7.1 Carrefour Scraper (Complex)
```go
// internal/scrapers/carrefour/carrefour.go

Target URL: https://www.carrefouruae.com/
Method: Playwright (dynamic JS site, AJAX-loaded content)
Frequency: Weekly
Complexity: HIGH

Why Playwright needed:
- Site uses heavy JavaScript for loading products
- Infinite scroll or pagination via JS
- AJAX requests for product data
- May have anti-bot protections

Scraping Strategy:
1. Navigate to category page (e.g., /grocery/dairy)
2. Wait for products to load (wait for specific selector)
3. Scroll to trigger lazy loading (if applicable)
4. Extract product cards:
   - Name
   - Price
   - Unit (per kg, per liter, per piece)
   - Brand
   - Promotions/discounts
5. Handle pagination

Focus on Staple Items (30-50 products):
- Milk (1L, 2L)
- Bread
- Eggs (12 pack)
- Rice (1kg, 5kg)
- Chicken breast (1kg)
- Tomatoes (1kg)
- Onions (1kg)
- Potatoes (1kg)
- Cooking oil
- Sugar
- Tea/Coffee

Output Structure:
{
  Category: "Food & Groceries",
  SubCategory: "Dairy",
  ItemName: "Fresh Milk 1L - Carrefour Brand",
  Price: 6.5,
  Location: {Emirate: "UAE", City: "Multiple"},
  Source: "carrefour_online",
  Attributes: {
    "store": "carrefour",
    "brand": "carrefour_own",
    "volume": "1L",
    "unit_price_per_liter": 6.5,
    "barcode": "...",
    "in_stock": true,
    "promotion": false
  },
  Tags: ["groceries", "dairy", "staple"]
}

Implementation:
func (s *CarrefourScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
    // Initialize browser via Playwright
    page, err := s.browserManager.NewPage()
    if err != nil {
        return nil, err
    }
    defer page.Close()

    // Navigate to dairy category
    if _, err := page.Goto("https://www.carrefouruae.com/mafuae/en/c/dairy"); err != nil {
        return nil, err
    }

    // Wait for products to load
    if err := page.WaitForSelector(".product-card", playwright.PageWaitForSelectorOptions{
        Timeout: playwright.Float(30000), // 30 seconds
    }); err != nil {
        return nil, err
    }

    // Extract product data
    products, err := page.Locator(".product-card").All()
    // ... parse products

    return dataPoints, nil
}
```

#### 7.2 Lulu Hypermarket Scraper (Complex)
```go
// internal/scrapers/lulu/lulu.go

Target URL: https://gcc.luluhypermarket.com/en-ae/grocery/
Method: Playwright
Frequency: Weekly
Complexity: HIGH

Similar to Carrefour:
- Dynamic JS site
- AJAX product loading
- Focus on same staple items basket
- Track prices over time

Additional consideration:
- Lulu has strong Asian product selection
- May want to track import items separately
```

**Estimated Effort:** 8-10 days (both scrapers)
**Files to Create:**
- `internal/scrapers/carrefour/carrefour.go`
- `internal/scrapers/carrefour/parser.go`
- `internal/scrapers/carrefour/products.go` (product list definitions)
- `internal/scrapers/lulu/lulu.go`
- `internal/scrapers/lulu/parser.go`
- `internal/scrapers/lulu/products.go`
- `internal/workflow/groceries_workflow.go`

**Challenges:**
- Anti-bot detection (may need to slow down, add delays)
- Site structure changes frequently
- May need to handle captchas (manual intervention or service)

---

### Batch 8: Government Data Integration
**Goal:** Leverage official open data portals
**Dependencies:** None (API-based, likely)

#### 8.1 Bayanat API Integration
```go
// internal/scrapers/bayanat/bayanat.go

Target URL: https://bayanat.ae/
Method: API (check for public API) OR web scraping if no API
Frequency: Monthly (datasets update infrequently)
Complexity: LOW-MEDIUM

Check for API:
- Many open data portals have CKAN, Socrata, or custom APIs
- Look for /api/ endpoints or documentation
- Example: https://bayanat.ae/api/3/action/package_search

Data to Extract:
- Economic indicators (inflation, CPI)
- Population demographics
- Transport statistics
- Education statistics
- Housing market reports

Use Cases:
- Contextual enrichment (not primary cost data)
- Trend analysis
- Inflation adjustments for historical data

Output Structure:
{
  Category: "Reference Data",
  SubCategory: "Economic Indicators",
  ItemName: "UAE Consumer Price Index",
  Price: 105.2,  // index value
  Source: "bayanat_official",
  Attributes: {
    "data_type": "cpi",
    "period": "Q3 2025",
    "base_year": "2020"
  }
}
```

#### 8.2 Dubai Statistics Center
```go
// internal/scrapers/dsc/dsc.go

Target URL: https://www.dsc.gov.ae/
Method: Check for API or downloadable datasets (CSV, Excel)
Frequency: Monthly
Complexity: MEDIUM

Data to Extract:
- Dubai-specific economic data
- Rental price indices (official)
- Population statistics
- Employment data

Use for:
- Validation of scraped data (compare our averages to official indices)
- Market trend analysis
```

**Estimated Effort:** 4-5 days
**Files to Create:**
- `internal/scrapers/bayanat/bayanat.go`
- `internal/scrapers/bayanat/api_client.go` (if API available)
- `internal/scrapers/dsc/dsc.go`
- `internal/workflow/government_data_workflow.go`

---

## 🔄 Workflow Orchestration

### Scheduler Configuration
```go
// internal/workflow/scheduler.go

type ScraperSchedule struct {
    Name      string
    Frequency time.Duration
    Priority  int  // 1-10, higher = more important
}

var Schedules = []ScraperSchedule{
    // Daily workflows (housing market changes frequently)
    {Name: "bayut_dubai", Frequency: 24 * time.Hour, Priority: 10},
    {Name: "bayut_sharjah", Frequency: 24 * time.Hour, Priority: 9},
    {Name: "bayut_ajman", Frequency: 24 * time.Hour, Priority: 8},
    {Name: "dubizzle_dubai", Frequency: 24 * time.Hour, Priority: 10},
    {Name: "dubizzle_sharjah", Frequency: 24 * time.Hour, Priority: 9},
    {Name: "dubizzle_shared", Frequency: 24 * time.Hour, Priority: 8},
    {Name: "rentit_bedspace", Frequency: 24 * time.Hour, Priority: 7},
    {Name: "homebook_bedspace", Frequency: 24 * time.Hour, Priority: 7},
    {Name: "ewaar_shared", Frequency: 24 * time.Hour, Priority: 6},

    // Weekly workflows (rates change infrequently)
    {Name: "dewa_rates", Frequency: 7 * 24 * time.Hour, Priority: 9},
    {Name: "sewa_rates", Frequency: 7 * 24 * time.Hour, Priority: 9},
    {Name: "aadc_rates", Frequency: 7 * 24 * time.Hour, Priority: 8},
    {Name: "rta_fares", Frequency: 7 * 24 * time.Hour, Priority: 8},
    {Name: "carrefour_groceries", Frequency: 7 * 24 * time.Hour, Priority: 7},
    {Name: "lulu_groceries", Frequency: 7 * 24 * time.Hour, Priority: 7},
    {Name: "carlift_prices", Frequency: 7 * 24 * time.Hour, Priority: 5},

    // Monthly workflows
    {Name: "careem_rates", Frequency: 30 * 24 * time.Hour, Priority: 6},
    {Name: "edarabia_schools", Frequency: 30 * 24 * time.Hour, Priority: 7},
    {Name: "bayanat_import", Frequency: 30 * 24 * time.Hour, Priority: 4},
    {Name: "dsc_stats", Frequency: 30 * 24 * time.Hour, Priority: 4},
}

func MasterSchedulerWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting master scheduler workflow")

    for _, schedule := range Schedules {
        // Schedule each scraper as a child workflow
        childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
            WorkflowID: fmt.Sprintf("%s-%s", schedule.Name, time.Now().Format("20060102")),
            CronSchedule: cronFromDuration(schedule.Frequency),
        })

        input := ScraperWorkflowInput{
            ScraperName: schedule.Name,
            MaxRetries:  3,
        }

        workflow.ExecuteChildWorkflow(childCtx, ScraperWorkflow, input)
    }

    return nil
}
```

---

## 🧪 Testing Strategy

### Unit Tests (Per Scraper)
```go
// internal/scrapers/dewa/dewa_rates_test.go

func TestDEWARatesScraper_ParseElectricitySlab(t *testing.T) {
    html := `<table>...</table>`  // sample HTML

    scraper := NewDEWARatesScraper(config)
    dataPoints, err := scraper.parseElectricitySlabs(html)

    assert.NoError(t, err)
    assert.Greater(t, len(dataPoints), 0)
    assert.Equal(t, "Utilities", dataPoints[0].Category)
}
```

### Integration Tests (Workflow)
```go
// internal/workflow/scraper_workflow_test.go

func TestDEWARatesWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    env.ExecuteWorkflow(DEWARatesWorkflow, ScraperWorkflowInput{
        ScraperName: "dewa_rates",
        MaxRetries:  1,
    })

    assert.True(t, env.IsWorkflowCompleted())
    assert.NoError(t, env.GetWorkflowError())
}
```

### End-to-End Tests
```bash
# scripts/test-scraper.sh
#!/bin/bash

# Test DEWA scraper end-to-end
go run cmd/scraper/main.go -scraper dewa_rates

# Verify data was stored
psql -d cost_of_living -c "SELECT * FROM cost_data_points WHERE source='dewa_official' ORDER BY recorded_at DESC LIMIT 5;"
```

---

## 📊 Monitoring & Observability

### Scraper Metrics (Extend Existing)
```go
// pkg/metrics/scraper_metrics.go

var (
    // Existing
    ScraperItemsScraped  *prometheus.CounterVec  // by scraper name
    ScraperErrorsTotal   *prometheus.CounterVec  // by scraper, error type

    // New
    ScraperDuration      *prometheus.HistogramVec  // scraper execution time
    ScraperLastRun       *prometheus.GaugeVec      // timestamp of last successful run
    ScraperDataFreshness *prometheus.GaugeVec      // hours since last data update
)
```

### Workflow Monitoring
```go
// Monitor via Temporal UI + Prometheus

Metrics to track:
- Workflow success/failure rate per scraper
- Average execution time per scraper
- Retry count per scraper
- Data points collected per run
- Time since last successful scrape (staleness)
```

### Alerting Rules
```yaml
# monitoring/alerts.yml

groups:
  - name: scraper_alerts
    rules:
      - alert: ScraperFailing
        expr: scraper_errors_total > 10
        for: 1h
        annotations:
          summary: "Scraper {{ $labels.scraper }} is experiencing high errors"

      - alert: DataStale
        expr: (time() - scraper_last_run_timestamp) > 86400
        annotations:
          summary: "Scraper {{ $labels.scraper }} hasn't run in 24 hours"
```

---

## 🚀 Implementation Order Summary

### Quick Reference
```
Batch 1: Multi-Emirate Extension       [3-4 days]  ← START HERE
  ├─ bayut_sharjah
  ├─ bayut_ajman
  ├─ dubizzle_sharjah
  └─ dubizzle_shared

Batch 2: Official Rates Scrapers       [5-6 days]
  ├─ dewa_rates
  ├─ sewa_rates
  ├─ aadc_rates
  ├─ rta_fares
  └─ careem_rates

Batch 3: Shared Accommodations         [6-7 days]
  ├─ rentit_bedspace
  ├─ homebook_bedspace
  └─ ewaar_shared

Batch 4: Education                     [5-6 days]
  └─ edarabia_schools

Batch 5: Carpooling                    [3-4 days]
  └─ carlift_services

Batch 6: Playwright Infrastructure     [2-3 days]  ← Foundation for Batch 7
  └─ pkg/browser setup

Batch 7: Groceries (Complex)           [8-10 days] ← Requires Batch 6
  ├─ carrefour_groceries
  └─ lulu_groceries

Batch 8: Government Data               [4-5 days]
  ├─ bayanat_api
  └─ dsc_stats

TOTAL ESTIMATED EFFORT: 36-45 days
```

---

## 🎯 Recommended Starting Point

### Start with Batch 1 + Batch 2
**Reasoning:**
1. **Batch 1** leverages existing code (quick win, multi-emirate coverage)
2. **Batch 2** demonstrates workflow approach for "semi-static" data (utilities, transport)
3. Both can be done in parallel by different developers/agents
4. Together they provide immediate value:
   - 3 emirates covered
   - Official utility rates (enables calculator)
   - Transport costs (RTA, Careem)

**Combined Effort:** 8-10 days
**Outcome:**
- Housing data from Dubai, Sharjah, Ajman
- Complete utility rate tables (DEWA, SEWA, AADC)
- RTA fare data
- Careem rate tracking
- 6-8 new workflows operational

---

## 🔧 Technical Setup Checklist

### Before Starting Any Batch

- [ ] Review existing scraper architecture (`internal/scrapers/bayut`, `internal/scrapers/dubizzle`)
- [ ] Understand current workflow structure (`internal/workflow/scraper_workflow.go`)
- [ ] Check Temporal worker registration (`cmd/worker/main.go`)
- [ ] Verify database schema can accommodate new categories (should be fine with JSONB)
- [ ] Set up test environment with database

### For Playwright Batches (Batch 6 & 7)

- [ ] Install Playwright for Go: `go get github.com/playwright-community/playwright-go`
- [ ] Install browser binaries: `playwright install chromium`
- [ ] Test basic browser automation locally
- [ ] Plan for headless mode in production
- [ ] Consider Docker image with Playwright dependencies

---

## 📝 Notes & Considerations

### Anti-Bot Strategies
- Start with simple rate limiting (1-2 requests/sec)
- Add random delays between requests (2-5 seconds)
- Rotate User-Agents
- For complex sites (groceries), use Playwright to appear more human-like
- Monitor error rates, adjust strategy as needed

### Data Quality
- All scraped data gets confidence score:
  - Official sources (DEWA, RTA): 0.95-0.98
  - Established sites (Bayut, Dubizzle): 0.80-0.85
  - New/untested sources: 0.60-0.75
- Implement outlier detection (flag prices >3 std deviations from mean)
- Deduplication logic (same property listed multiple times)

### Legal & Ethical
- Check robots.txt for each site
- Respect rate limits
- Identify scraper via User-Agent (don't pretend to be a regular browser)
- For official government data, prioritize APIs over scraping
- If a site blocks, respect it and seek alternatives

### Failure Handling
- All workflows have retry policies (exponential backoff)
- Failed scrapes logged to metrics
- Compensation activities (mark data as stale if scrape fails)
- Alert on consecutive failures (>3 in 24 hours)

---

## ✅ Success Criteria

### Per Batch
- [ ] All workflows in batch registered and running
- [ ] Data successfully stored in database
- [ ] Tests passing (unit + integration)
- [ ] Prometheus metrics showing successful scrapes
- [ ] No errors in Temporal workflow history
- [ ] Data visible via API endpoints

### Overall (All Batches Complete)
- [ ] 18+ workflows operational
- [ ] 10,000+ cost data points in database
- [ ] 7+ categories covered (Housing, Utilities, Transport, Education, Food, Services, Shared)
- [ ] 3+ emirates covered (Dubai, Sharjah, Ajman minimum)
- [ ] Daily/weekly/monthly scrapes running on schedule
- [ ] 90%+ scraper success rate
- [ ] <5% duplicate data
- [ ] API serving fresh data (<24h old for daily scrapers)

---

## 🤔 Decision Points

### 1. Start with Batch 1 or Batch 2?
**Option A:** Batch 1 (Multi-Emirate)
- Pro: Leverages existing code, quick win
- Pro: 2x-3x housing data immediately
- Con: Doesn't demonstrate workflow approach for different data types

**Option B:** Batch 2 (Official Rates)
- Pro: Demonstrates scraping static/semi-static data
- Pro: Enables utility calculator feature
- Con: Doesn't expand housing coverage

**Option C:** Both in parallel
- Pro: Best of both worlds
- Con: More coordination needed

**Recommendation:** Start with Batch 1 (quick win), then Batch 2

---

### 2. Defer Playwright batches?
**Question:** Should we defer Batch 6 & 7 (groceries with Playwright)?

**Defer if:**
- Want to launch faster with simpler scrapers
- Playwright adds complexity/dependencies
- Groceries are lower priority than other categories

**Don't defer if:**
- Groceries are key value proposition
- Have bandwidth for complex scrapers
- Want complete cost-of-living coverage

**Recommendation:** Do Batches 1-5 first, evaluate, then decide on 6-7

---

### 3. Government data priority?
**Question:** Is Batch 8 (government data) worth the effort?

**Value:**
- Contextual enrichment (inflation data, official indices)
- Validation of scraped data
- Credibility boost (using official sources)

**Effort:**
- 4-5 days
- Depends on API availability

**Recommendation:** Medium-low priority, do after Batches 1-5

---

## 📅 Suggested Timeline (Sequential)

```
Week 1:     Batch 1 (Multi-Emirate)
Week 2:     Batch 2 Part 1 (DEWA, SEWA, AADC)
Week 3:     Batch 2 Part 2 (RTA, Careem) + Batch 3 Start
Week 4:     Batch 3 Complete (Shared Accommodations)
Week 5:     Batch 4 (Education) + Batch 5 (Carpooling)
Week 6:     Batch 6 (Playwright Setup)
Week 7-8:   Batch 7 (Groceries with Playwright)
Week 9:     Batch 8 (Government Data) + Testing/Polish

TOTAL: 9 weeks (45 days)
```

---

## 🚀 Next Steps

**If you approve this plan:**

1. I'll create implementation todos for Batch 1
2. Start with extending Bayut scraper to Sharjah
3. Then Ajman, then Dubizzle extensions
4. Test each workflow as we go
5. Move to Batch 2 once Batch 1 is complete

**Questions for you:**
1. Start with Batch 1 (Multi-Emirate Extension)?
2. Run Batch 1 + 2 in parallel or sequential?
3. Any batches to deprioritize or skip?
4. Comfortable with Playwright for groceries (Batches 6-7)?

Ready to start implementing Batch 1? 🎯
