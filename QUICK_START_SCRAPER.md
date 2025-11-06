# Quick Start: Running the Bayut Scraper

## Prerequisites

1. Database running:
```bash
docker-compose up -d postgres
```

2. Migrations applied:
```bash
make migrate
```

3. Environment variables set (or use `.env` file):
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=cost_of_living
export DB_SSLMODE=disable
```

## Run the Scraper

### Method 1: Using Makefile (Easiest)
```bash
make scrape-bayut
```

### Method 2: Direct Go Run
```bash
go run cmd/scraper/main.go -scraper bayut
```

### Method 3: Build and Run Binary
```bash
go build -o bin/scraper cmd/scraper/main.go
./bin/scraper -scraper bayut
```

## Expected Output

```
{"level":"INFO","msg":"Starting scraper CLI","scraper":"bayut"}
{"level":"INFO","msg":"Connected to database successfully"}
{"level":"INFO","msg":"Registered scraper","name":"bayut"}
{"level":"INFO","msg":"Running scraper","name":"bayut"}
{"level":"INFO","msg":"Starting Bayut scrape"}
{"level":"INFO","msg":"Completed Bayut scrape","count":6}
{"level":"INFO","msg":"Scraper completed","name":"bayut","scraped":6,"saved":6}
{"level":"INFO","msg":"Scraping completed successfully"}
```

## Verify Data

### Check record count:
```bash
docker exec cost-of-living-db psql -U postgres -d cost_of_living \
  -c "SELECT COUNT(*) FROM cost_data_points WHERE source='bayut';"
```

### View latest listings:
```bash
docker exec cost-of-living-db psql -U postgres -d cost_of_living \
  -c "SELECT item_name, price, location FROM cost_data_points WHERE source='bayut' LIMIT 5;"
```

### Get price statistics:
```bash
docker exec cost-of-living-db psql -U postgres -d cost_of_living \
  -c "SELECT MIN(price) as min, MAX(price) as max, AVG(price)::numeric(10,2) as avg FROM cost_data_points WHERE source='bayut';"
```

## Run Tests

```bash
# All scraper tests
go test ./internal/scrapers/... -v

# Just Bayut parser tests
go test ./internal/scrapers/bayut/... -v

# With coverage
go test ./internal/scrapers/bayut/... -cover
```

## Common Issues

### "Database connection failed"
- Ensure PostgreSQL is running: `docker-compose ps`
- Check environment variables
- Verify database exists: `make migrate`

### "No data scraped"
- Check internet connectivity
- Website might have changed structure
- Check logs for specific errors

### "Rate limit exceeded"
- Wait a few seconds and retry
- Scraper limits to 1 request per second

## File Structure

```
cmd/scraper/
  └── main.go                    # CLI entry point

internal/scrapers/
  ├── scraper.go                 # Interface definition
  ├── README.md                  # Full documentation
  └── bayut/
      ├── bayut.go              # Main scraper
      ├── parser.go             # Data parsing
      └── parser_test.go        # Tests

internal/services/
  └── scraper_service.go        # Service layer
```

## Next Steps

1. **Schedule with Temporal** (Iteration 1.8)
   - Run scraper periodically
   - Handle failures with retries
   - Monitor execution

2. **Add More Scrapers**
   - Dubizzle for housing
   - Supermarkets for food prices
   - Transport data

3. **Enhance Current Scraper**
   - Pagination (more than 10 listings)
   - More categories (sale, commercial)
   - Image extraction

## Support

See `/internal/scrapers/README.md` for detailed documentation.
See `ITERATION_1.7_SUMMARY.md` for implementation details.
