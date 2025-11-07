#!/bin/bash

# Test Coverage Reporter
# Generates comprehensive test coverage reports for the entire project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== UAE Cost of Living - Coverage Report ===${NC}"
echo ""

# Get project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Remove old coverage files
echo -e "${YELLOW}Cleaning old coverage files...${NC}"
rm -f coverage*.out coverage*.html 2>/dev/null || true

# Run all tests with coverage
echo ""
echo -e "${YELLOW}Running all tests with coverage...${NC}"
echo ""

# Run unit tests
echo -e "${BLUE}→ Unit tests...${NC}"
go test -v -race -coverprofile=coverage-unit.out -covermode=atomic ./internal/... ./pkg/... 2>&1 | grep -E "(PASS|FAIL|coverage:|---)"

# Run integration tests
echo ""
echo -e "${BLUE}→ Integration tests...${NC}"
go test -v -race -coverprofile=coverage-integration.out -covermode=atomic ./test/integration/... 2>&1 | grep -E "(PASS|FAIL|coverage:|---)" || echo "No integration tests found"

# Combine all coverage files
echo ""
echo -e "${YELLOW}Combining coverage reports...${NC}"

# Create combined coverage file
echo "mode: atomic" > coverage-all.out

# Merge all coverage files
for f in coverage-unit.out coverage-integration.out; do
    if [ -f "$f" ]; then
        tail -q -n +2 "$f" >> coverage-all.out
    fi
done

# Generate coverage reports
echo ""
echo -e "${GREEN}=== Coverage Summary ===${NC}"
echo ""

# Overall coverage
echo -e "${BLUE}Overall Coverage:${NC}"
go tool cover -func=coverage-all.out | tail -1

echo ""
echo -e "${BLUE}Package Coverage:${NC}"
go tool cover -func=coverage-all.out | grep -E "total:" | head -10

echo ""
echo -e "${BLUE}Detailed Coverage by Component:${NC}"
echo ""

# Scrapers
echo -e "${YELLOW}Scrapers:${NC}"
go tool cover -func=coverage-all.out | grep -E "internal/scrapers" | grep -v "test" || echo "  No coverage data"

echo ""
echo -e "${YELLOW}Workflows:${NC}"
go tool cover -func=coverage-all.out | grep -E "internal/workflow" | grep -v "test" || echo "  No coverage data"

echo ""
echo -e "${YELLOW}Services:${NC}"
go tool cover -func=coverage-all.out | grep -E "internal/services" | grep -v "test" || echo "  No coverage data"

echo ""
echo -e "${YELLOW}Repositories:${NC}"
go tool cover -func=coverage-all.out | grep -E "internal/repository" | grep -v "test" || echo "  No coverage data"

# Generate HTML reports
echo ""
echo -e "${YELLOW}Generating HTML coverage reports...${NC}"

# Overall coverage
go tool cover -html=coverage-all.out -o coverage-all.html
echo -e "${GREEN}✓ Overall coverage: coverage-all.html${NC}"

# Unit tests coverage
if [ -f "coverage-unit.out" ]; then
    go tool cover -html=coverage-unit.out -o coverage-unit.html
    echo -e "${GREEN}✓ Unit test coverage: coverage-unit.html${NC}"
fi

# Integration tests coverage
if [ -f "coverage-integration.out" ]; then
    go tool cover -html=coverage-integration.out -o coverage-integration.html
    echo -e "${GREEN}✓ Integration test coverage: coverage-integration.html${NC}"
fi

# Check coverage threshold
echo ""
echo -e "${YELLOW}Checking coverage threshold...${NC}"

COVERAGE=$(go tool cover -func=coverage-all.out | grep total | awk '{print $3}' | sed 's/%//')
THRESHOLD=70

echo "Coverage: ${COVERAGE}%"
echo "Threshold: ${THRESHOLD}%"

if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✓ Coverage meets threshold!${NC}"
else
    echo -e "${RED}✗ Coverage below threshold (${COVERAGE}% < ${THRESHOLD}%)${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Coverage Report Complete ===${NC}"
