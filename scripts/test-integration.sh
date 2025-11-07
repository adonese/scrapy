#!/bin/bash

# Integration Test Runner
# Runs all integration tests for the UAE Cost of Living project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== UAE Cost of Living - Integration Tests ===${NC}"
echo ""

# Get project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Run integration tests
echo -e "${YELLOW}Running integration tests...${NC}"
echo ""

# Run tests with verbose output and race detection
go test -v -race -timeout 5m ./test/integration/... \
    -coverprofile=coverage-integration.out \
    -covermode=atomic

TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All integration tests passed!${NC}"

    # Generate coverage report
    echo ""
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -func=coverage-integration.out | tail -1

    # Generate HTML coverage report
    go tool cover -html=coverage-integration.out -o coverage-integration.html
    echo -e "${GREEN}✓ HTML coverage report: coverage-integration.html${NC}"
else
    echo ""
    echo -e "${RED}✗ Integration tests failed${NC}"
    exit $TEST_EXIT_CODE
fi

echo ""
echo -e "${GREEN}=== Integration Tests Complete ===${NC}"
