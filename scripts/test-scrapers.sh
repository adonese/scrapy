#!/bin/bash

# Scraper Test Runner
# Runs tests specifically for scraper implementations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== UAE Cost of Living - Scraper Tests ===${NC}"
echo ""

# Get project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Running Bayut scraper tests...${NC}"
go test -v -race ./internal/scrapers/bayut/... \
    -coverprofile=coverage-bayut.out

echo ""
echo -e "${YELLOW}Running Dubizzle scraper tests...${NC}"
go test -v -race ./internal/scrapers/dubizzle/... \
    -coverprofile=coverage-dubizzle.out

echo ""
echo -e "${YELLOW}Running scraper integration tests...${NC}"
go test -v -race ./test/integration/... \
    -run "TestBayut|TestDubizzle" \
    -coverprofile=coverage-scraper-integration.out

# Combine coverage reports
echo ""
echo -e "${YELLOW}Generating combined coverage report...${NC}"

# Merge coverage files
echo "mode: atomic" > coverage-scrapers.out
tail -q -n +2 coverage-bayut.out coverage-dubizzle.out coverage-scraper-integration.out >> coverage-scrapers.out 2>/dev/null || true

# Display coverage
echo ""
echo -e "${BLUE}=== Scraper Coverage Summary ===${NC}"
go tool cover -func=coverage-scrapers.out | grep -E "(bayut|dubizzle|scraper)" | grep -v "test"

echo ""
echo -e "${YELLOW}Total scraper coverage:${NC}"
go tool cover -func=coverage-scrapers.out | tail -1

# Generate HTML report
go tool cover -html=coverage-scrapers.out -o coverage-scrapers.html
echo ""
echo -e "${GREEN}✓ HTML coverage report: coverage-scrapers.html${NC}"

echo ""
echo -e "${GREEN}=== Scraper Tests Complete ===${NC}"
