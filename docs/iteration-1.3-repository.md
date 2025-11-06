# Iteration 1.3: Repository Pattern and CRUD Operations

## Overview

This iteration implements the repository pattern with full CRUD operations for the `cost_data_points` table. The implementation follows clean architecture principles with clear separation between domain models, repository interfaces, and database implementations.

## Components Implemented

### 1. Domain Model (`/internal/models/cost_data_point.go`)

The `CostDataPoint` model represents the core data structure with:
- Basic cost information (price, min, max, median)
- Location data (emirate, city, area, coordinates)
- Temporal data (recorded_at, valid_from, valid_to)
- Metadata (source, confidence, unit, tags)
- Flexible attributes (JSONB)

**Key Structures:**
```go
type CostDataPoint struct {
    ID          string
    Category    string
    SubCategory string
    ItemName    string
    Price       float64
    // ... other fields
    Location    Location
    Tags        []string
    Attributes  map[string]interface{}
}

type Location struct {
    Emirate     string
    City        string
    Area        string
    Coordinates *GeoPoint
}
```

### 2. Repository Interface (`/internal/repository/cost_data_point.go`)

Clean interface defining all CRUD operations:
- `Create(ctx, *CostDataPoint) error`
- `GetByID(ctx, id, recordedAt) (*CostDataPoint, error)`
- `List(ctx, filter) ([]*CostDataPoint, error)`
- `Update(ctx, *CostDataPoint) error`
- `Delete(ctx, id, recordedAt) error`

**ListFilter Options:**
- Category filtering (exact match)
- Emirate filtering (JSON field query)
- Date range filtering (start/end date)
- Pagination (limit/offset)

### 3. PostgreSQL Implementation (`/internal/repository/postgres/cost_data_point.go`)

Full implementation with:
- Proper JSONB handling for location and attributes
- PostgreSQL array handling for tags (using lib/pq)
- Nullable field support
- Auto-generated UUIDs and timestamps
- Context-aware operations
- Comprehensive error handling

**Key Features:**
- Handles composite primary key (id, recorded_at)
- Marshals/unmarshals JSONB fields transparently
- Converts between Go types and SQL types
- Parameterized queries to prevent SQL injection
- Proper NULL handling with sql.NullString, sql.NullFloat64, etc.

### 4. Integration Tests (`/internal/repository/postgres/cost_data_point_test.go`)

Comprehensive test suite covering:
- **Create**: successful creation, custom ID, minimal fields
- **GetByID**: existing records, non-existent records
- **List**: all records, category filter, emirate filter, date range, pagination, combined filters
- **Update**: successful update, non-existent records
- **Delete**: successful deletion, non-existent records

All tests use real database (integration tests).

## Project Structure

```
/home/adonese/src/cost-of-living/
├── internal/
│   ├── models/
│   │   └── cost_data_point.go          # Domain model
│   └── repository/
│       ├── cost_data_point.go          # Repository interface
│       └── postgres/
│           ├── cost_data_point.go      # PostgreSQL implementation
│           └── cost_data_point_test.go # Integration tests
├── examples/
│   └── repository_usage.go             # Usage examples
├── scripts/
│   └── test-repository.sh              # Test runner script
├── cmd/api/main.go                     # Wired up repository
└── Makefile                            # Updated with test commands
```

## Usage Examples

### Basic CRUD Operations

```go
// Initialize repository
repo := postgres.NewCostDataPointRepository(db.GetConn())
ctx := context.Background()

// Create
cdp := &models.CostDataPoint{
    Category: "Housing",
    ItemName: "1BR Apartment in Marina",
    Price:    85000,
    Location: models.Location{
        Emirate: "Dubai",
        City:    "Dubai",
        Area:    "Marina",
    },
    Source: "example",
}
err := repo.Create(ctx, cdp)

// Get by ID
retrieved, err := repo.GetByID(ctx, cdp.ID, cdp.RecordedAt)

// List with filters
filter := repository.ListFilter{
    Category: "Housing",
    Emirate:  "Dubai",
    Limit:    10,
}
results, err := repo.List(ctx, filter)

// Update
cdp.Price = 90000
err = repo.Update(ctx, cdp)

// Delete
err = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
```

### Advanced Filtering

```go
// Date range query
startDate := time.Now().AddDate(0, 0, -30)
endDate := time.Now()

filter := repository.ListFilter{
    Category:  "Food",
    Emirate:   "Dubai",
    StartDate: &startDate,
    EndDate:   &endDate,
    Limit:     100,
}
results, err := repo.List(ctx, filter)
```

## Testing

### Running Tests

```bash
# Setup database and run migrations
make setup

# Run all tests
make test

# Run only repository tests
make test-repo

# Run repository tests directly
go test -v ./internal/repository/postgres/... -count=1
```

### Test Requirements

Tests require:
1. Running PostgreSQL database with TimescaleDB
2. Migrations applied to create tables
3. Test data is automatically cleaned up

### Test Coverage

All CRUD operations are covered with:
- Happy path scenarios
- Error scenarios
- Edge cases (minimal fields, empty results, etc.)
- JSONB and array handling
- Composite primary key handling

## Implementation Decisions

### 1. Composite Primary Key Handling

The table uses `(id, recorded_at)` as primary key for TimescaleDB optimization. All operations that target specific records require both values:

```go
GetByID(ctx, id string, recordedAt time.Time)
Delete(ctx, id string, recordedAt time.Time)
```

### 2. JSONB Encoding/Decoding

Location and attributes are stored as JSONB:
- **Encoding**: Marshal Go struct/map to JSON before INSERT/UPDATE
- **Decoding**: Unmarshal JSON bytes to Go struct/map after SELECT

This provides flexibility while maintaining type safety in Go code.

### 3. Array Handling

Tags are stored as PostgreSQL TEXT[] array using `github.com/lib/pq`:
- Use `pq.Array()` for scanning and inserting
- Transparently converts between Go []string and PostgreSQL TEXT[]

### 4. Nullable Fields

Optional fields use sql.Null* types:
- `sql.NullString` for optional text fields
- `sql.NullFloat64` for optional numeric fields
- `sql.NullTime` for optional timestamp fields

Helper functions convert between Go types and Null types.

### 5. Auto-generation

The repository auto-generates:
- **UUIDs** if not provided (using uuid_generate_v4())
- **Timestamps** (recorded_at, valid_from) if not set
- **Default values** (sample_size=1, confidence=1.0, unit="AED")

### 6. Error Handling

- Returns specific errors for common cases (e.g., "cost data point not found")
- Wraps database errors with context using fmt.Errorf
- Distinguishes between sql.ErrNoRows and other errors

## Database Queries

### Create Query
```sql
INSERT INTO cost_data_points (
    id, category, sub_category, item_name, price, min_price, max_price,
    median_price, sample_size, location, recorded_at, valid_from, valid_to,
    source, source_url, confidence, unit, tags, attributes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING created_at, updated_at
```

### List Query (with filters)
```sql
SELECT * FROM cost_data_points
WHERE category = $1
  AND location->>'emirate' = $2
  AND recorded_at >= $3
  AND recorded_at <= $4
ORDER BY recorded_at DESC
LIMIT $5 OFFSET $6
```

### Update Query
```sql
UPDATE cost_data_points SET
    category = $1, sub_category = $2, item_name = $3,
    price = $4, min_price = $5, max_price = $6,
    -- ... other fields
WHERE id = $18 AND recorded_at = $19
```

## Performance Considerations

1. **Indexes**: Existing indexes on category, emirate, recorded_at ensure fast queries
2. **GIN Indexes**: JSONB and array fields have GIN indexes for efficient filtering
3. **Parameterized Queries**: All queries use parameterized statements (prevent SQL injection and enable query caching)
4. **Connection Pooling**: Reuses existing database connection pool from pkg/database
5. **TimescaleDB**: Table is a hypertable partitioned by recorded_at for time-series optimization

## What's NOT Included (By Design)

- ❌ ORM (using pure database/sql)
- ❌ Complex query builders
- ❌ Full-text search
- ❌ Advanced filtering (by tags, attributes, coordinates)
- ❌ Caching layer
- ❌ Transaction management
- ❌ Audit logging
- ❌ Soft deletes

These can be added in future iterations as needed.

## Next Steps (Iteration 1.4)

The repository is ready for integration with REST API endpoints:

1. Create handlers in `/internal/handlers/cost_data_point.go`
2. Add routes to main.go
3. Implement request/response DTOs
4. Add validation
5. Add API documentation

Example endpoints to implement:
- `POST /api/v1/cost-data-points` - Create
- `GET /api/v1/cost-data-points/:id` - Get by ID
- `GET /api/v1/cost-data-points` - List with filters
- `PUT /api/v1/cost-data-points/:id` - Update
- `DELETE /api/v1/cost-data-points/:id` - Delete

## Verification

The implementation is verified and working:

```bash
✅ All tests passing (100%)
✅ Code compiles successfully
✅ Database integration working
✅ JSONB handling correct
✅ Array handling correct
✅ Filtering working
✅ Pagination working
```

Test output:
```
PASS: TestCreate (3 subtests)
PASS: TestGetByID (2 subtests)
PASS: TestList (6 subtests)
PASS: TestUpdate (2 subtests)
PASS: TestDelete (2 subtests)
ok  	github.com/adonese/cost-of-living/internal/repository/postgres	0.213s
```
