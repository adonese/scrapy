# UAE Cost of Living Data Models

## Core Design Principles
- **Time-series first**: Every data point has temporal context for trend analysis
- **Experience-driven**: Combine hard data with qualitative experiences
- **Growth-oriented**: Built-in viral mechanics and user value loops
- **Liability-conscious**: User content is structured as "personal experiences" not advice

## 1. Time-Series Cost Data

```go
// Base model for all cost items with time-series support
type CostDataPoint struct {
    ID           string    `db:"id" json:"id"`
    Category     string    `db:"category" json:"category"`
    SubCategory  string    `db:"sub_category" json:"sub_category"`
    ItemName     string    `db:"item_name" json:"item_name"`

    // Price data with statistical depth
    Price        float64   `db:"price" json:"price"`
    MinPrice     float64   `db:"min_price" json:"min_price"`
    MaxPrice     float64   `db:"max_price" json:"max_price"`
    MedianPrice  float64   `db:"median_price" json:"median_price"`
    SampleSize   int       `db:"sample_size" json:"sample_size"`

    // Temporal and geographic context
    Location     Location  `db:"location" json:"location"`
    RecordedAt   time.Time `db:"recorded_at" json:"recorded_at"`
    ValidFrom    time.Time `db:"valid_from" json:"valid_from"`
    ValidTo      time.Time `db:"valid_to" json:"valid_to"`

    // Source tracking
    Source       string    `db:"source" json:"source"`
    SourceURL    string    `db:"source_url" json:"source_url"`
    Confidence   float32   `db:"confidence" json:"confidence"` // 0-1 score

    // Metadata
    Unit         string    `db:"unit" json:"unit"`
    Tags         []string  `db:"tags" json:"tags"`
    Attributes   JSONB     `db:"attributes" json:"attributes"` // flexible field
}

type Location struct {
    Emirate      string    `json:"emirate"`
    City         string    `json:"city"`
    Area         string    `json:"area"`
    Coordinates  *GeoPoint `json:"coordinates,omitempty"`
}

// Aggregated trends for fast queries
type TrendSnapshot struct {
    ID           string    `db:"id"`
    Category     string    `db:"category"`
    Location     Location  `db:"location"`
    Period       string    `db:"period"` // daily, weekly, monthly

    // Statistical measures
    AvgPrice     float64   `db:"avg_price"`
    PriceChange  float64   `db:"price_change"` // % change
    Volatility   float64   `db:"volatility"`

    // Trend indicators
    TrendDirection string  `db:"trend_direction"` // up, down, stable
    TrendStrength  float32 `db:"trend_strength"` // 0-1

    SnapshotDate time.Time `db:"snapshot_date"`
}
```

## 2. User Experience Layer (Reddit-like but safer)

```go
// Anonymous experience sharing with verification
type Experience struct {
    ID          string    `db:"id"`
    UserHash    string    `db:"user_hash"` // anonymous but consistent

    // Content
    Title       string    `db:"title"`
    Content     string    `db:"content"`
    Category    string    `db:"category"`
    Tags        []string  `db:"tags"`

    // Verification without liability
    Type        string    `db:"type"` // "personal_experience", "observation", "tip"
    Disclaimer  bool      `db:"disclaimer"` // always true

    // Context
    Location    Location  `db:"location"`
    DateRange   string    `db:"date_range"` // "2024", "Q1 2024", vague for privacy

    // Engagement (for growth hacks)
    ViewCount   int       `db:"view_count"`
    HelpfulCount int      `db:"helpful_count"`
    SaveCount   int       `db:"save_count"`
    ShareCount  int       `db:"share_count"`

    // Moderation
    Status      string    `db:"status"` // pending, approved, flagged
    AIScore     float32   `db:"ai_score"` // sentiment/helpfulness score

    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
}

// Cost breakdowns shared by users
type BudgetShare struct {
    ID          string    `db:"id"`
    UserHash    string    `db:"user_hash"`

    // Demographics (optional, for context)
    Profile     struct {
        FamilySize   int    `json:"family_size,omitempty"`
        Lifestyle    string `json:"lifestyle,omitempty"` // budget, moderate, premium
        WorkSector   string `json:"work_sector,omitempty"`
    } `db:"profile"`

    // Monthly breakdown
    Housing      float64  `db:"housing"`
    Transport    float64  `db:"transport"`
    Food         float64  `db:"food"`
    Utilities    float64  `db:"utilities"`
    Healthcare   float64  `db:"healthcare"`
    Education    float64  `db:"education"`
    Savings      float64  `db:"savings"`
    Other        float64  `db:"other"`

    TotalMonthly float64  `db:"total_monthly"`
    Location     Location `db:"location"`

    // Growth mechanics
    Views        int      `db:"views"`
    Compares     int      `db:"compares"` // how many compared with this

    SharedAt     time.Time `db:"shared_at"`
}
```

## 3. Growth & Engagement Models

```go
// Calculators that users create and share (viral mechanism)
type CustomCalculator struct {
    ID           string   `db:"id"`
    CreatorHash  string   `db:"creator_hash"`

    Name         string   `db:"name"`
    Description  string   `db:"description"`

    // Calculator logic stored as JSON
    Formula      JSONB    `db:"formula"`
    Categories   []string `db:"categories"`

    // Viral metrics
    UseCount     int      `db:"use_count"`
    ForkCount    int      `db:"fork_count"`
    ShareURL     string   `db:"share_url"`

    CreatedAt    time.Time `db:"created_at"`
}

// Comparison requests (growth hack: "See how you compare")
type Comparison struct {
    ID          string    `db:"id"`
    UserHash    string    `db:"user_hash"`

    UserData    JSONB     `db:"user_data"` // their inputs
    Percentile  float32   `db:"percentile"`

    // Insights generated
    Insights    []string  `db:"insights"`

    // Share mechanism
    ShareCode   string    `db:"share_code"`
    SharedCount int       `db:"shared_count"`

    CreatedAt   time.Time `db:"created_at"`
}

// Email alerts for price changes (retention)
type PriceAlert struct {
    ID          string    `db:"id"`
    Email       string    `db:"email"` // hashed

    Category    string    `db:"category"`
    Location    Location  `db:"location"`
    Threshold   float64   `db:"threshold"`
    Direction   string    `db:"direction"` // above, below

    Active      bool      `db:"active"`
    LastTriggered *time.Time `db:"last_triggered"`

    CreatedAt   time.Time `db:"created_at"`
}
```

## 4. Analytics & Intelligence Models

```go
// Predictions based on historical data
type PricePrediction struct {
    ID          string    `db:"id"`
    Category    string    `db:"category"`
    Location    Location  `db:"location"`

    PredictionDate time.Time `db:"prediction_date"`
    TargetDate     time.Time `db:"target_date"`

    PredictedPrice float64  `db:"predicted_price"`
    Confidence     float32  `db:"confidence"`
    Model          string   `db:"model"` // which ML model

    // Track accuracy
    ActualPrice    *float64 `db:"actual_price"`
    Error          *float64 `db:"error"`
}

// Seasonal patterns detection
type SeasonalPattern struct {
    ID          string   `db:"id"`
    Category    string   `db:"category"`

    Pattern     string   `db:"pattern"` // "ramadan_spike", "summer_dip", etc
    Magnitude   float64  `db:"magnitude"` // % change

    StartMonth  int      `db:"start_month"`
    EndMonth    int      `db:"end_month"`

    Confidence  float32  `db:"confidence"`
    YearsData   int      `db:"years_data"` // how many years analyzed
}

// Area scoring for recommendations
type AreaScore struct {
    ID          string    `db:"id"`
    Location    Location  `db:"location"`

    // Composite scores (0-100)
    AffordabilityScore float32 `db:"affordability_score"`
    ConvenienceScore   float32 `db:"convenience_score"`
    QualityScore       float32 `db:"quality_score"`

    // Factors
    AvgRent           float64 `db:"avg_rent"`
    TransportAccess   float32 `db:"transport_access"`
    AmenitiesCount    int     `db:"amenities_count"`

    // User feedback integration
    UserRating        float32 `db:"user_rating"`
    ExperienceCount   int     `db:"experience_count"`

    LastUpdated       time.Time `db:"last_updated"`
}
```

## 5. Search & Discovery Optimization

```go
// Search index for experiences
type SearchIndex struct {
    ID          string   `db:"id"`
    ContentType string   `db:"content_type"` // experience, calculator, budget
    ContentID   string   `db:"content_id"`

    // Full-text search fields
    Title       string   `db:"title"`
    Body        string   `db:"body"`
    Tags        []string `db:"tags"`

    // Ranking factors
    Popularity  float32  `db:"popularity"`
    Freshness   float32  `db:"freshness"`
    Relevance   float32  `db:"relevance"`

    // Facets for filtering
    Location    Location `db:"location"`
    Category    string   `db:"category"`
    DateRange   string   `db:"date_range"`
}
```

## 6. Growth Hack Implementations

### Viral Loops Built Into Models:
1. **"How do you compare?"** - Users input their costs, see percentile, share results
2. **Budget Templates** - Users share anonymized budgets others can fork
3. **Price Drop Alerts** - Email notifications drive return visits
4. **Experience Karma** - Helpful content earns points, unlocks features
5. **Calculator Builder** - Users create custom calculators, share with friends

### Liability Protection:
- All user content marked as "personal experience"
- Automatic disclaimers on every page
- No financial advice, only data and experiences
- AI moderation for problematic content
- Anonymous by default (hash-based identity)

## Database Migrations

```sql
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS pg_trgm; -- for fuzzy search

-- Convert main tables to hypertables for time-series
SELECT create_hypertable('cost_data_points', 'recorded_at');
SELECT create_hypertable('trend_snapshots', 'snapshot_date');

-- Indexes for common queries
CREATE INDEX idx_cost_category_location_time ON cost_data_points(category, location, recorded_at DESC);
CREATE INDEX idx_experiences_category_status ON experiences(category, status) WHERE status = 'approved';
CREATE GIN INDEX idx_search_fulltext ON search_index USING gin(to_tsvector('english', title || ' ' || body));
```

## Next Steps:
1. Implement TimescaleDB for efficient time-series queries
2. Set up materialized views for trend calculations
3. Build AI pipeline for content moderation
4. Create sharing mechanisms with tracking
5. Implement anonymous but consistent user identification