# Cost of Living API - Quick Reference

## Base URL
```
http://localhost:8080
```

## Quick Start

### 1. Start the API
```bash
# Start database
docker-compose up -d

# Run migrations
go run cmd/migrate/main.go

# Start API server
go run cmd/api/main.go
```

### 2. Test Endpoints

#### Health Check
```bash
curl http://localhost:8080/health
```

## API Endpoints Cheat Sheet

### CREATE
```bash
curl -X POST http://localhost:8080/api/v1/cost-data-points \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Housing",
    "item_name": "1BR Apartment",
    "price": 85000,
    "location": {"emirate": "Dubai"},
    "source": "manual"
  }'
```

### READ (Get by ID)
```bash
# Note: Replace {id} and {recorded_at} with actual values
curl "http://localhost:8080/api/v1/cost-data-points/{id}?recorded_at={recorded_at}"
```

### READ (List with Filters)
```bash
# All records
curl "http://localhost:8080/api/v1/cost-data-points"

# With filters
curl "http://localhost:8080/api/v1/cost-data-points?category=Housing&emirate=Dubai&limit=10"

# With date range
curl "http://localhost:8080/api/v1/cost-data-points?start_date=2025-01-01T00:00:00Z&end_date=2025-12-31T23:59:59Z"
```

### UPDATE
```bash
curl -X PUT "http://localhost:8080/api/v1/cost-data-points/{id}?recorded_at={recorded_at}" \
  -H "Content-Type: application/json" \
  -d '{"price": 90000}'
```

### DELETE
```bash
curl -X DELETE "http://localhost:8080/api/v1/cost-data-points/{id}?recorded_at={recorded_at}"
```

## Query Parameters

### List Endpoint
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| category | string | Filter by category | `category=Housing` |
| emirate | string | Filter by emirate | `emirate=Dubai` |
| start_date | RFC3339 | Records from date | `start_date=2025-01-01T00:00:00Z` |
| end_date | RFC3339 | Records to date | `end_date=2025-12-31T23:59:59Z` |
| limit | int | Max records (max: 100) | `limit=20` |
| offset | int | Skip records | `offset=10` |

## Request Body Examples

### Minimal Create Request
```json
{
  "category": "Housing",
  "item_name": "1BR Apartment",
  "price": 85000,
  "location": {"emirate": "Dubai"},
  "source": "manual"
}
```

### Full Create Request
```json
{
  "category": "Housing",
  "sub_category": "Rental",
  "item_name": "1BR Apartment Marina",
  "price": 85000,
  "min_price": 80000,
  "max_price": 90000,
  "median_price": 85000,
  "sample_size": 10,
  "location": {
    "emirate": "Dubai",
    "city": "Dubai",
    "area": "Marina",
    "coordinates": {
      "lat": 25.0772,
      "lon": 55.1344
    }
  },
  "recorded_at": "2025-11-06T12:00:00Z",
  "valid_from": "2025-11-01T00:00:00Z",
  "valid_to": "2025-11-30T23:59:59Z",
  "source": "survey",
  "source_url": "https://example.com/survey",
  "confidence": 0.95,
  "unit": "AED",
  "tags": ["rental", "furnished"],
  "attributes": {
    "bedrooms": 1,
    "bathrooms": 1,
    "furnished": true
  }
}
```

### Update Request (Partial)
```json
{
  "price": 90000,
  "confidence": 0.98
}
```

## Response Examples

### Success Response (Create/Get/Update)
```json
{
  "id": "352a9181-750b-4f19-b609-dcdd61d6f541",
  "category": "Housing",
  "item_name": "1BR Apartment Marina",
  "price": 85000,
  "sample_size": 1,
  "location": {
    "emirate": "Dubai",
    "city": "Dubai",
    "area": "Marina"
  },
  "recorded_at": "2025-11-06T16:04:47.341764Z",
  "valid_from": "2025-11-06T16:04:47.341765Z",
  "source": "manual",
  "confidence": 1,
  "unit": "AED",
  "created_at": "2025-11-06T16:04:47.342262Z",
  "updated_at": "2025-11-06T16:04:47.342262Z"
}
```

### List Response
```json
{
  "data": [
    {
      "id": "...",
      "category": "Housing",
      ...
    }
  ],
  "total_count": 10,
  "limit": 10,
  "offset": 0
}
```

### Error Response
```json
{
  "error": "Bad Request",
  "message": "Validation error: ...",
  "code": 400
}
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Successful GET, PUT |
| 201 | Created - Successful POST |
| 204 | No Content - Successful DELETE |
| 400 | Bad Request - Validation errors |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error - Server/DB error |

## Categories

Suggested categories for the UAE:
- Housing
- Food
- Transportation
- Healthcare
- Education
- Entertainment
- Utilities
- Personal Care
- Clothing
- Communication

## Emirates

Valid emirates:
- Abu Dhabi
- Dubai
- Sharjah
- Ajman
- Umm Al Quwain
- Ras Al Khaimah
- Fujairah

## Testing Script

Run comprehensive API tests:
```bash
./test_api.sh
```

## Development

### Run Tests
```bash
# All tests
go test ./... -v

# Handler tests only
go test ./internal/handlers/... -v

# With coverage
go test ./... -cover
```

### Build
```bash
go build -o api ./cmd/api
```

### Run
```bash
./api
```

## Tips

1. **Timestamps**: Always use RFC3339 format for dates (e.g., `2025-11-06T16:04:47Z`)
2. **Composite Key**: Remember that records have composite primary key (id, recorded_at)
3. **URL Encoding**: When passing timestamps in URL, ensure proper encoding of special characters
4. **Pagination**: Always use pagination for large datasets (limit max: 100)
5. **Validation**: Check required fields before submitting (category, item_name, price, location.emirate, source)

## Common Patterns

### Create and Get
```bash
# Create
response=$(curl -s -X POST http://localhost:8080/api/v1/cost-data-points \
  -H "Content-Type: application/json" \
  -d '{"category": "Housing", "item_name": "1BR Apartment", "price": 85000, "location": {"emirate": "Dubai"}, "source": "manual"}')

# Extract ID and recorded_at
id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
recorded_at=$(echo "$response" | grep -o '"recorded_at":"[^"]*' | cut -d'"' -f4)

# Get the created record
curl "http://localhost:8080/api/v1/cost-data-points/${id}?recorded_at=${recorded_at}"
```

### Filter and Paginate
```bash
# Get first page
curl "http://localhost:8080/api/v1/cost-data-points?category=Housing&limit=10&offset=0"

# Get second page
curl "http://localhost:8080/api/v1/cost-data-points?category=Housing&limit=10&offset=10"
```

### Update Multiple Fields
```bash
curl -X PUT "http://localhost:8080/api/v1/cost-data-points/${id}?recorded_at=${recorded_at}" \
  -H "Content-Type: application/json" \
  -d '{
    "price": 90000,
    "confidence": 0.95,
    "tags": ["updated", "verified"]
  }'
```

## Troubleshooting

### Database Connection Failed
```bash
# Check if database is running
docker-compose ps

# Start database if stopped
docker-compose up -d

# Check logs
docker-compose logs postgres
```

### Port Already in Use
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Or change PORT in .env file
PORT=8081
```

### Validation Errors
- Ensure all required fields are present
- Check that price > 0
- Verify emirate spelling
- Confirm timestamp format (RFC3339)

## More Information

- Full documentation: `ITERATION_1.4_SUMMARY.md`
- Testing guide: `TESTING_GUIDE.md`
- Data models: `data_models.md`
