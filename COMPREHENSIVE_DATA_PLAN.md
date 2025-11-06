# Comprehensive Data Collection Plan - UAE Cost of Living
## Planning Document for Expanded Data Pointers

**Date:** November 6, 2025
**Branch:** `claude/plan-data-pointers-011CUs7kjBZWyydWgkFRR2nM`
**Status:** Planning Phase

---

## Executive Summary

This document outlines a comprehensive plan to expand the UAE Cost of Living Calculator from its current focus (housing rentals in Dubai via Bayut/Dubizzle) to a complete cost-of-living platform covering:

1. **Education** (school fees)
2. **Hidden rental costs** (deposits, commissions, utilities setup)
3. **Shared accommodations** (bed spaces, partitions)
4. **Transportation** (RTA, Careem, carpools)
5. **Multi-emirate coverage** (Sharjah, Ajman, Abu Dhabi)
6. **Utilities** (DEWA, SEWA, AADC)
7. **Job-seeker optimization** (affordable areas near business hubs)
8. **Groceries & food**
9. **Government data integration**

---

## Current State Analysis

### What We Have ✅
- Go-based API with Echo framework
- PostgreSQL + TimescaleDB for time-series data
- Temporal workflows for scheduled scraping
- 2 operational scrapers:
  - **Bayut**: Dubai apartment rentals
  - **Dubizzle**: Dubai apartment rentals
- Flexible `CostDataPoint` model supporting:
  - Categories, subcategories, locations (JSONB)
  - Attributes (JSONB for extensibility)
  - Tags, confidence scores, time-series data
  - Price ranges (min, max, median)
- Prometheus metrics + structured logging
- REST API with CRUD operations

### Architecture Strengths
- **Extensible data model**: JSONB fields allow flexible attributes per category
- **Time-series ready**: TimescaleDB hypertable for trend analysis
- **Workflow orchestration**: Temporal handles scheduling, retries, compensation
- **Observable**: Metrics and logging infrastructure in place

### Gaps to Address
- Only housing data (no other cost categories)
- Single emirate coverage (Dubai only)
- No hidden cost calculations
- No transportation data
- No education costs
- No utility rate comparisons
- No job-seeker optimization features

---

## Research Findings: Data Sources

### 1. Education / School Fees 🎓

**Primary Sources:**
- **Edarabia.com** (`https://www.edarabia.com/dubai-school-fees/`)
  - Comprehensive database: school fees by grade
  - KHDA ratings included
  - Filterable by curriculum, grade
  - PDF export available
  - **Scrapeability**: HIGH (structured HTML tables)

- **SchoolsCompared.com** (`https://schoolscompared.com/`)
  - Detailed school profiles
  - Fees, ratings, academic results
  - Membership features (may require auth)
  - **Scrapeability**: MEDIUM (may need browser automation)

- **DubaiSchools.com** (`https://dubaischools.com/`)
  - School directory with fees
  - Blog content with guides
  - **Scrapeability**: MEDIUM

**Data Range:** AED 12,723 to AED 110,000+ per annum
**Regulation:** KHDA Education Cost Index (2.35% for 2025-26)

**Data Structure:**
```go
Category: "Education"
SubCategory: "School Fees"
ItemName: "GEMS Wellington Primary School - Grade 1"
Price: 45000.0
Attributes: {
  "curriculum": "British",
  "grade": "1",
  "khda_rating": "Outstanding",
  "school_name": "GEMS Wellington Primary",
  "age_range": "3-11"
}
Tags: ["education", "school", "british-curriculum", "outstanding"]
```

---

### 2. Hidden Rental Costs 💰

**Cost Categories:**

| Cost Item | Amount | Type | Refundable |
|-----------|--------|------|------------|
| Security Deposit | 5% (unfurnished) / 10% (furnished) of annual rent | One-time | Yes |
| Agency Commission | 5% of annual rent | One-time | No |
| DEWA Deposit | AED 2,000 (apt) / AED 4,000 (villa) | One-time | Yes |
| DEWA Connection | AED 100-300 | One-time | No |
| Ejari Registration | AED 220 | Annual | No |
| Municipality Fee | 5% of annual rent | Monthly (via DEWA) | No |
| Chiller Fees | Varies by building | Monthly | No |

**Sources:**
- Bayut MyBayut blog (`https://www.bayut.com/mybayut/hidden-costs-renting-dubai/`)
- Dubizzle blog (`https://www.dubizzle.com/blog/property/hidden-costs-renting-dubai/`)
- DEWA official tariff calculator (`https://www.dewa.gov.ae/en/consumer/billing/tariff-calculator`)

**Implementation Strategy:**
- Create a **Hidden Costs Calculator** feature
- Input: Annual rent amount, property type (furnished/unfurnished, apt/villa)
- Output: Breakdown of all one-time and recurring costs
- Store as `CostDataPoint` with:
  ```go
  Category: "Housing"
  SubCategory: "Hidden Costs"
  ItemName: "Complete Move-in Costs - 1BR Unfurnished Apartment"
  Price: calculated_total
  Attributes: {
    "annual_rent": 80000,
    "security_deposit": 4000,
    "agency_commission": 4000,
    "dewa_deposit": 2000,
    "dewa_connection": 100,
    "ejari": 220,
    "monthly_municipality_fee": 333.33,
    "breakdown": {...}
  }
  ```

---

### 3. Shared Accommodations 🛏️

**Primary Sources:**

1. **Dubizzle** (`https://dubai.dubizzle.com/property-for-rent/rooms-for-rent-flatmates/`)
   - 6,323+ shared room listings (UAE-wide)
   - Bed spaces, private rooms, partitions
   - **Scrapeability**: MEDIUM (anti-bot protection, already implemented)

2. **RentItOnline.ae** (`https://rentitonline.ae/room-rental/dubai`)
   - Specializes in bed spaces
   - Targets professionals, females, students, families
   - **Scrapeability**: HIGH

3. **Homebook.ae** (`https://homebook.ae/bed-space/`)
   - Starting AED 600/month
   - No agents, no hidden fees
   - Gender-filtered listings
   - **Scrapeability**: HIGH

4. **Ewaar** (`https://dubai.ewaar.com/flatmates-bed-space-for-rent-in-dubai/`)
   - Bed spaces, partitions, rooms
   - Multiple areas (Deira, International City, Karama)
   - **Scrapeability**: MEDIUM

5. **Bayut Shared** (`https://www.bayut.com/s/shared-accommodations-for-monthly-rent-in-dubai/`)
   - Monthly basis short-term
   - Integrated with main platform
   - **Scrapeability**: MEDIUM (already have Bayut scraper)

6. **RoomDaddy** (`https://roomdaddy.com/flat-share/uae/dubai`)
   - Monthly rentals
   - **Scrapeability**: MEDIUM

**Price Range:** AED 400-2,500/month

**Data Structure:**
```go
Category: "Housing"
SubCategory: "Shared Accommodation"
ItemName: "Bed Space in International City - Male"
Price: 650.0
Unit: "AED"
Attributes: {
  "accommodation_type": "bed_space",  // bed_space, partition, shared_room, private_room
  "gender": "male",
  "location_details": "International City - China Cluster",
  "furnishing": "furnished",
  "amenities": ["wifi", "washing_machine", "kitchen_access"],
  "occupancy": "2_per_room"
}
Tags: ["shared", "bed-space", "budget", "international-city"]
```

---

### 4. Transportation 🚇🚗

#### 4.1 Public Transport (RTA)

**Official Sources:**
- **Dubai Metro Fare Calculator** (`https://www.dubaimetrorails.com/fare-calculator`)
- **RTA Official** (via data scraping or API if available)

**Fare Structure (2025):**
- **Metro/Bus Zones**: 7 zones total
- **Silver Card (Standard)**: AED 3 - AED 7.5 per journey
- **Gold Card**: 2x Silver price
- **Blue Card (Discounted)**: 50% off for students/seniors
- **Day Pass**: AED 20-22 unlimited
- **Free Transfer**: Bus within 30 min of metro

**Implementation:**
- Static data (changes infrequently)
- Manual updates with government announcements
- Store as reference data:
  ```go
  Category: "Transportation"
  SubCategory: "Public Transport"
  ItemName: "Dubai Metro - 1 Zone Silver Card"
  Price: 3.0
  Attributes: {
    "zones": 1,
    "card_type": "silver",
    "class": "standard"
  }
  ```

#### 4.2 Ride-Sharing (Careem, Uber)

**Careem (Nov 2025 Update):**
- Base fare: AED 13
- Per km: AED 2.26
- Waiting time: AED 0.5/min
- Salik (toll): AED 5/gate
- Peak surcharges: AED 7.50 (weekend evenings)

**Uber Pool:**
- Starting: AED 12/trip

**Implementation Strategy:**
- **Option 1**: Manual rate updates (ride-sharing rates change frequently)
- **Option 2**: Scrape Careem/Uber blog announcements
- **Option 3**: Calculate estimates based on known rates + common routes

**Data Structure:**
```go
Category: "Transportation"
SubCategory: "Ride Sharing"
ItemName: "Careem Base Fare"
Price: 13.0
Attributes: {
  "service": "careem",
  "rate_type": "base",
  "additional_charges": {
    "per_km": 2.26,
    "per_minute_waiting": 0.5,
    "salik": 5.0
  },
  "effective_date": "2025-11-05"
}
```

#### 4.3 Carpooling / Carlift Services

**Sources:**
- CarpoolsUAE.com (`https://carpoolsuae.com/`)
- MrBus.ae (`https://mrbus.ae/car-lift-prices/`)
- DubaiCarlift.com (`https://dubaicarlift.com/`)

**Price Ranges:**
- Daily: AED 10-25 per trip
- Weekly: AED 700
- Monthly: AED 350-1,500 (varies by route, shared vs private)

**Popular Routes:**
- Dubai to Abu Dhabi
- Dubai to Sharjah
- Within Dubai (to business districts)

**Data Structure:**
```go
Category: "Transportation"
SubCategory: "Carpooling"
ItemName: "Dubai to Abu Dhabi - Monthly Carlift"
Price: 800.0
Attributes: {
  "route": "dubai_to_abudhabi",
  "frequency": "daily_commute",
  "type": "shared_seat",
  "provider": "carlift_uae"
}
```

---

### 5. Multi-Emirate Coverage 🏙️

#### 5.1 Sharjah

**Rental Sources:**
- **Sharjah Municipality Rental Map** (`https://portal.shjmun.gov.ae/en/eservices/`)
  - Official government platform
  - Real-time rental data
  - **Scrapeability**: LOW (government portal, may require API)

- **Dubizzle Sharjah** (extend existing scraper)
- **Bayut Sharjah** (extend existing scraper)
- **PropertyFinder Sharjah**

**Rent Ranges (2025):**
- 1BR Apartment: AED 22,000-30,000/year
- 3BR Villa: AED 70,000-100,000/year

**Utilities (SEWA):**
- Electricity: 20-30 fils/kWh
- Water: AED 3-4/gallon (tiered)
- Sewerage: 1.5 fils/gallon (new in 2025)

#### 5.2 Ajman

**Sources:**
- Dubizzle Ajman
- Bayut Ajman
- PropertyFinder Ajman

**Rent Ranges:**
- 1BR: AED 18,000-28,000/year (cheapest in UAE!)
- 3BR Villa: AED 55,000-80,000/year

**Utilities:**
- Electricity: Similar to Sharjah (SEWA-like structure)
- Water: Tiered pricing

#### 5.3 Abu Dhabi

**Sources:**
- PropertyFinder Abu Dhabi
- Dubizzle Abu Dhabi
- Bayut Abu Dhabi

**Utilities (AADC/ADDC):**
- Electricity: 26.8-30.5 fils/kWh (expatriates, band-based)
- Water: AED 5.95-10.55 per 1,000 liters
- Green band (low usage) vs Red band (high usage)

**Data Structure:**
```go
Location: {
  Emirate: "Sharjah",
  City: "Sharjah",
  Area: "Al Nahda"
}
Category: "Housing"
SubCategory: "Rent"
ItemName: "1BR Apartment - Al Nahda Sharjah"
Price: 25000.0
```

---

### 6. Utilities 🔌💧

**Implementation Strategy:**
Create a **Utility Cost Calculator** with:
- Input: Emirate, property type, estimated monthly consumption
- Output: Estimated monthly utility costs

**Data Sources:**
- **DEWA** (`https://www.dewa.gov.ae/en/consumer/billing/tariff-calculator`)
- **SEWA** (`https://www.sewa.gov.ae/en/energy-calculator`)
- **AADC** (`https://www.aadc.ae/en/pages/maintarrif.aspx`)

**DEWA (Dubai):**
```go
Category: "Utilities"
SubCategory: "Electricity"
ItemName: "DEWA Electricity Fuel Surcharge"
Price: 0.05  // 5 fils per kWh
Attributes: {
  "emirate": "Dubai",
  "utility_provider": "DEWA",
  "rate_type": "fuel_surcharge",
  "unit": "fils_per_kwh",
  "effective_date": "2025-01-01"
}
```

**SEWA (Sharjah):**
```go
Category: "Utilities"
SubCategory: "Water"
ItemName: "SEWA Water - Tier 1 (0-6000 gallons)"
Price: 3.0  // AED per gallon
Attributes: {
  "emirate": "Sharjah",
  "utility_provider": "SEWA",
  "consumption_range": "0-6000",
  "unit": "aed_per_gallon"
}
```

**Additional Charges:**
```go
Category: "Utilities"
SubCategory: "Additional Fees"
ItemName: "Dubai Municipality Housing Fee"
Price: 5.0  // 5% of annual rent
Attributes: {
  "emirate": "Dubai",
  "calculation": "percentage_of_rent",
  "billing": "monthly_via_dewa"
}
```

---

### 7. Job Seeker Optimized Areas 💼

**Concept:** Identify affordable areas near major business hubs

**Business Hubs:**
- Business Bay
- Dubai Media City (DMC)
- Dubai Internet City (DIC)
- Dubai Marina
- Downtown Dubai
- DIFC (Dubai International Financial Centre)
- Knowledge Village
- Academic City

**Affordable Nearby Areas:**

| Area | Near Hub(s) | Rent Range (1BR) | Commute Time |
|------|-------------|------------------|--------------|
| Jumeirah Lake Towers (JLT) | Business Bay, Marina, DMC | AED 55K-75K | 10-15 min |
| Al Barsha | DMC, DIC | AED 45K-59K | 10 min |
| Barsha Heights (Tecom) | DIC, DMC, JLT | AED 50K-70K | 5-10 min |
| Jumeirah Village Circle (JVC) | Business Bay | AED 60K-90K | 20 min |
| Discovery Gardens | DMC, DIC | AED 45K-65K | 10 min (metro) |
| Business Bay | Downtown, DIFC | AED 70K-100K | 5 min |
| Deira | Various | AED 35K-50K | 30 min (metro) |

**Data Structure:**
```go
Category: "Housing"
SubCategory: "Job Seeker Optimized"
ItemName: "1BR Apartment - JLT (near Business Bay)"
Price: 65000.0
Attributes: {
  "area": "Jumeirah Lake Towers",
  "near_hubs": ["Business Bay", "Dubai Marina", "Media City"],
  "commute_time": {
    "business_bay": "10-15 min",
    "media_city": "10 min"
  },
  "transport_options": ["metro", "bus", "taxi"],
  "metro_nearby": true,
  "job_seeker_score": 9.0  // 0-10 scale
}
Tags: ["job-seeker", "affordable", "central", "metro-access"]
```

**Implementation:**
1. Tag existing housing data with nearby business hubs
2. Calculate "job-seeker scores" based on:
   - Proximity to business districts
   - Rent affordability (percentile-based)
   - Public transport access
   - Amenities (groceries, gyms, etc.)
3. Create API endpoint: `/api/v1/job-seeker-areas?budget=5000&workplace=business_bay`

---

### 8. Groceries & Food 🛒

**Primary Sources:**

1. **Carrefour UAE** (`https://www.carrefouruae.com/`)
   - 31 hypermarkets, 73 supermarkets
   - Online shopping with API (check if public API available)
   - Own-brand products at lower prices
   - **Scrapeability**: MEDIUM (dynamic site, may need browser automation)

2. **Lulu Hypermarket** (`https://gcc.luluhypermarket.com/en-ae/grocery/`)
   - 103 stores in UAE
   - Products from 85 countries
   - Strong in imported Asian products
   - Online shopping available
   - **Scrapeability**: MEDIUM

3. **ShoppingInformer.com** (`https://www.shoppinginformer.com/dubai/`)
   - Aggregates promotions from multiple stores
   - Lulu, Carrefour, others
   - **Scrapeability**: HIGH (may already aggregate data)

**Implementation Strategy:**
- Focus on **staple items** (basket of goods):
  - Milk, bread, eggs, rice, chicken, vegetables, fruits
- Track prices over time (time-series data)
- Identify best-value stores per category

**Data Structure:**
```go
Category: "Food & Groceries"
SubCategory: "Dairy"
ItemName: "Fresh Milk 1L - Carrefour"
Price: 6.5
Attributes: {
  "store": "carrefour",
  "brand": "carrefour_own",
  "volume": "1L",
  "unit_price_per_liter": 6.5,
  "location": "multiple",  // available across all stores
  "promotion": false
}
```

---

### 9. Government Data Integration 📊

**Primary Sources:**

1. **Bayanat** (`https://bayanat.ae/`)
   - Official UAE open data portal
   - 3,018 datasets from 50 entities
   - Categories: economy, education, health, transport, etc.
   - **API Availability**: Likely yes (open data portals typically have APIs)

2. **Dubai Statistics Center** (`https://www.dsc.gov.ae/`)
   - Dubai-specific statistics
   - Economic indicators, population data
   - **API Availability**: Check

3. **Federal Competitiveness and Statistics Centre** (`https://opendata.fcsc.gov.ae/`)
   - National-level statistics
   - **API Availability**: Likely

**Use Cases:**
- **Inflation data**: Adjust historical prices
- **Population demographics**: Contextualize demand
- **Economic indicators**: Salary trends, unemployment
- **Education statistics**: School capacity, demand
- **Transport statistics**: Usage patterns

**Implementation:**
- Check for APIs
- If no API, download CSVs/datasets periodically
- Store as reference data
- Use for enrichment and analysis (not primary cost data)

---

## Data Model Enhancements

### New Categories & Subcategories

```go
// Existing
- Housing
  - Rent
  - [NEW] Shared Accommodation
  - [NEW] Hidden Costs

// New Categories
- Education
  - School Fees
  - University Fees
  - Training/Courses

- Transportation
  - Public Transport (RTA)
  - Ride Sharing (Careem, Uber)
  - Carpooling
  - Taxi
  - [FUTURE] Car Ownership (fuel, insurance, registration)

- Utilities
  - Electricity
  - Water
  - Sewerage
  - Gas
  - Internet/Telecom
  - Additional Fees (municipality, etc.)

- Food & Groceries
  - Dairy
  - Meat & Poultry
  - Fruits & Vegetables
  - Grains & Bread
  - Household Items

- Services
  - Domestic Help
  - Cleaning Services
  - Laundry
  - Healthcare (out-of-pocket)

- Entertainment & Leisure
  - Gym Memberships
  - Cinema
  - Dining Out
  - Activities
```

### Attribute Enhancements

**Location Object:**
```go
type Location struct {
    Emirate     string    `json:"emirate"`      // Dubai, Sharjah, Ajman, etc.
    City        string    `json:"city"`
    Area        string    `json:"area"`
    Coordinates *GeoPoint `json:"coordinates"`

    // New fields
    NearbyHubs  []string  `json:"nearby_hubs,omitempty"`     // ["Business Bay", "DIFC"]
    MetroAccess bool      `json:"metro_access,omitempty"`
    MetroStation string   `json:"metro_station,omitempty"`
}
```

**New Tag System:**
```go
Tags: [
  "job-seeker-friendly",
  "budget-friendly",
  "family-friendly",
  "metro-accessible",
  "shared-accommodation",
  "hidden-cost",
  "government-source",
  "time-sensitive",  // for rates that change frequently
]
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
**Priority: HIGH**

#### 1.1 Hidden Costs Calculator
- Create calculation logic for all hidden rental costs
- Input: Annual rent, property type
- Output: Complete breakdown
- Store calculated data as `CostDataPoint`
- **No scraping needed** (calculation-based)

#### 1.2 Extend Location Coverage
- Modify existing Bayut/Dubizzle scrapers:
  - Add Sharjah URLs
  - Add Ajman URLs
  - Add Abu Dhabi URLs
- Test with different emirates
- Validate location parsing

#### 1.3 Static Data: RTA & Utilities
- Create seed data files for:
  - RTA fare tables (zones 1-7)
  - DEWA rate slabs
  - SEWA rate slabs
  - AADC/ADDC rate bands
- Create migration to populate reference data
- Build utility cost calculator endpoint

**Deliverables:**
- [ ] Hidden cost calculator API endpoint
- [ ] Multi-emirate scraper support
- [ ] Utility rate reference data
- [ ] RTA fare reference data
- [ ] Updated data model with new categories

---

### Phase 2: Shared Accommodations (Week 3-4)
**Priority: HIGH**

#### 2.1 New Scrapers
- **RentItOnline.ae**
  - Likely simple HTML structure
  - Extract: price, type, location, gender, amenities

- **Homebook.ae**
  - Focus on bed spaces
  - Extract: price, area, facilities

- **Extend Dubizzle scraper**
  - Add shared accommodation category
  - Handle bed space / partition listings

#### 2.2 Data Normalization
- Standardize accommodation types:
  - `bed_space`, `partition`, `shared_room`, `private_room`
- Extract amenities consistently
- Handle gender filtering

**Deliverables:**
- [ ] 3 new scrapers operational
- [ ] 500+ shared accommodation listings
- [ ] Search/filter API for shared accommodations

---

### Phase 3: Education Data (Week 5-6)
**Priority: MEDIUM-HIGH**

#### 3.1 School Fees Scraper
- **Edarabia.com**
  - School listings with fees
  - KHDA ratings
  - Curriculum info

- **SchoolsCompared.com** (optional, if time permits)
  - More detailed data
  - May require browser automation

#### 3.2 Data Structure
- Store per grade level
- Include curriculum, ratings, location
- Enable filtering by budget, curriculum, area

**Deliverables:**
- [ ] School fees scraper
- [ ] 100+ schools with fee data
- [ ] School search API endpoint

---

### Phase 4: Transportation (Week 7)
**Priority: MEDIUM**

#### 4.1 Ride-Sharing Rates
- Manual data entry for Careem/Uber rates
- Create update process for rate changes
- Consider scraping Careem/Uber announcement blogs

#### 4.2 Carpooling Services
- Scrape pricing from carlift websites
- Store common routes (Dubai-Abu Dhabi, etc.)
- Monthly vs daily pricing

**Deliverables:**
- [ ] Transportation cost calculator
- [ ] Careem/Uber rate data
- [ ] Carpooling options database

---

### Phase 5: Groceries (Week 8-9)
**Priority: MEDIUM**

#### 5.1 Carrefour Scraper
- Online shopping site
- Focus on staple items (30-50 products)
- Track price changes over time

#### 5.2 Lulu Scraper
- Similar to Carrefour
- Same basket of goods
- Compare prices

#### 5.3 Price Comparison Feature
- Show cheapest store per item
- Basket comparison
- Historical price trends

**Deliverables:**
- [ ] Carrefour scraper
- [ ] Lulu scraper
- [ ] Price comparison API
- [ ] 50+ tracked grocery items

---

### Phase 6: Government Data & Enrichment (Week 10)
**Priority: LOW-MEDIUM**

#### 6.1 Bayanat Integration
- Investigate API availability
- Identify relevant datasets
- Periodic import process

#### 6.2 Job Seeker Optimization
- Calculate scores for existing housing data
- Tag properties near business hubs
- Build search/recommendation API

**Deliverables:**
- [ ] Government data integration
- [ ] Job seeker search endpoint
- [ ] Area recommendation engine

---

### Phase 7: Advanced Features (Week 11-12)
**Priority: LOW (Nice-to-have)

#### 7.1 Cost-of-Living Calculators
- Monthly budget estimator
- Emirate comparison tool
- Lifestyle-based calculations (budget/moderate/premium)

#### 7.2 Search Optimization
- Advanced filters
- Multi-criteria search
- Saved searches / alerts

#### 7.3 Data Quality
- Deduplication improvements
- Outlier detection
- Confidence scoring refinement

**Deliverables:**
- [ ] Budget calculator tool
- [ ] Enhanced search features
- [ ] Data quality dashboard

---

## Technical Considerations

### Scraping Strategy

#### Complexity Tiers
1. **Static/Simple** (use Go + goquery):
   - Edarabia, RentItOnline, Homebook

2. **Medium** (existing Bayut/Dubizzle pattern):
   - Bayut extended, Dubizzle extended
   - Handle anti-bot measures with retries, delays

3. **Complex** (consider browser automation):
   - Carrefour, Lulu (dynamic JS sites)
   - SchoolsCompared (if needed)
   - Use Playwright or Selenium in Go

#### Anti-Bot Handling
- User-Agent rotation
- Rate limiting (1-2 requests/sec)
- Exponential backoff on failures
- Proxy rotation (if needed)
- Consider headless browsers for stubborn sites

### Data Storage

#### Time-Series Optimization
- TimescaleDB already configured
- Create continuous aggregates for:
  - Average rent by area (monthly)
  - Price trends (quarterly)
  - School fee changes (yearly)

#### Indexing Strategy
- Add GIN indexes for new attribute fields
- Create materialized views for common queries
- Consider full-text search for school names

### Temporal Workflows

#### Scraper Scheduling
- **High frequency** (daily):
  - Housing rentals (Bayut, Dubizzle)
  - Shared accommodations

- **Medium frequency** (weekly):
  - Groceries (prices change often)

- **Low frequency** (monthly):
  - School fees (change annually, but listings update)
  - Utilities (rates change rarely)

- **On-demand**:
  - Hidden cost calculations
  - Job seeker optimizations

#### Workflow Organization
```go
// Extend existing workflow structure
scrapers := []string{
  "bayut_dubai", "bayut_sharjah", "bayut_ajman",
  "dubizzle_dubai", "dubizzle_sharjah",
  "rentit_bedspace", "homebook_bedspace",
  "edarabia_schools",
  "carrefour_groceries", "lulu_groceries",
}
```

### API Design

#### New Endpoints

```
GET  /api/v1/cost-data-points              # Enhanced with new categories
GET  /api/v1/hidden-costs?rent=80000&type=apartment
GET  /api/v1/utility-estimate?emirate=dubai&usage=500
GET  /api/v1/schools?curriculum=british&max_fee=50000
GET  /api/v1/shared-accommodations?area=international_city&max=800
GET  /api/v1/job-seeker-areas?budget=5000&workplace=business_bay
GET  /api/v1/transportation/routes?from=jlt&to=business_bay
GET  /api/v1/groceries/basket?store=carrefour&items=milk,bread,eggs
GET  /api/v1/compare-emirates?category=housing&type=1BR

POST /api/v1/budget-calculator
  Body: {
    "rent_budget": 5000,
    "family_size": 3,
    "has_car": false,
    "school_age_children": 1,
    "lifestyle": "moderate"
  }
  Response: {
    "housing": {...},
    "education": {...},
    "transportation": {...},
    "groceries": {...},
    "utilities": {...},
    "total_monthly": 12500
  }
```

### Performance Considerations

- **Caching**: Redis for frequently accessed data (RTA fares, utility rates)
- **Pagination**: All list endpoints with default limit=50, max=100
- **Rate Limiting**: Per-endpoint rate limits
- **CDN**: Static assets, cached responses

---

## Success Metrics

### Coverage Metrics
- [ ] 3+ emirates covered (Dubai, Sharjah, Ajman minimum)
- [ ] 5+ data categories (Housing, Education, Transport, Utilities, Food)
- [ ] 10+ data sources scraped
- [ ] 10,000+ cost data points stored

### Quality Metrics
- [ ] 90%+ scraper success rate
- [ ] Average confidence score > 0.75
- [ ] <5% duplicate data
- [ ] Data freshness: 90% updated within 7 days

### Usage Metrics
- [ ] API response time < 200ms (p95)
- [ ] 99% API uptime
- [ ] Comprehensive documentation

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Anti-bot blocking | HIGH | Browser automation, proxies, rate limiting |
| Website structure changes | MEDIUM | Regular monitoring, alerts, test suite |
| Rate limiting by sources | MEDIUM | Temporal retry policies, backoff strategies |
| Data quality issues | MEDIUM | Validation rules, outlier detection, manual review |
| Legal/ToS violations | HIGH | Review ToS, use public data, rate limit responsibly |
| Government data API limits | LOW | Cache heavily, periodic batch imports |
| Scope creep | MEDIUM | Phased approach, clear priorities |

---

## Next Steps

### Immediate Actions (This Session)
1. ✅ Research complete
2. ✅ Plan documented
3. [ ] Discuss priorities with user
4. [ ] Confirm Phase 1 scope
5. [ ] Begin implementation (if approved)

### Questions for User
1. **Priority confirmation**: Is Phase 1 (Hidden Costs + Multi-Emirate + Utilities) the right starting point?
2. **Scraping ethics**: Any concerns about scraping commercial sites? Should we focus more on government/public data?
3. **Browser automation**: Willing to add Playwright/Selenium for complex sites (Carrefour, Lulu)?
4. **Timeline**: Is the 12-week roadmap realistic, or should we condense/extend?
5. **Budget calculators**: Should these be in Phase 1 (high value) or keep in Phase 7?

---

## Appendix: Data Source Summary

### Ready to Scrape (High Priority)
- ✅ Bayut (extend to Sharjah, Ajman)
- ✅ Dubizzle (extend to Sharjah, shared accommodations)
- ✅ RentItOnline.ae
- ✅ Homebook.ae
- ✅ Edarabia.com

### Moderate Effort
- 🔶 Ewaar.com
- 🔶 RoomDaddy.com
- 🔶 SchoolsCompared.com

### Complex (May Need Browser Automation)
- 🔴 Carrefour UAE
- 🔴 Lulu Hypermarket
- 🔴 Careem blog (for rate updates)

### Static/Manual Entry
- 📝 RTA fare tables
- 📝 DEWA rate slabs
- 📝 SEWA rate slabs
- 📝 AADC/ADDC rate bands
- 📝 Careem/Uber current rates

### Government/Open Data
- 🏛️ Bayanat.ae
- 🏛️ Dubai Statistics Center
- 🏛️ FCSC Open Data
- 🏛️ Sharjah Municipality Rental Map

---

**End of Planning Document**

**Document Version:** 1.0
**Last Updated:** 2025-11-06
**Author:** Claude (AI Assistant)
