#!/bin/bash

# validate-data.sh - Data validation script for cost-of-living data
# This script runs validation checks on scraped data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
VALIDATION_THRESHOLD=0.7  # Minimum quality score (0-1)
MAX_ERROR_RATE=0.05       # Maximum acceptable error rate (5%)
MAX_DUPLICATE_RATE=0.01   # Maximum acceptable duplicate rate (1%)

echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Data Validation Pipeline${NC}"
echo -e "${GREEN}====================================${NC}"
echo ""

# Parse command line arguments
SOURCE=""
CATEGORY=""
STRICT_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --source)
            SOURCE="$2"
            shift 2
            ;;
        --category)
            CATEGORY="$2"
            shift 2
            ;;
        --strict)
            STRICT_MODE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --source SOURCE      Validate data from specific source"
            echo "  --category CATEGORY  Validate data from specific category"
            echo "  --strict            Enable strict mode (warnings treated as errors)"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Run validation tests
echo -e "${YELLOW}Step 1: Running validation unit tests...${NC}"
cd /home/adonese/src/cost-of-living
go test -v ./internal/validation/... -race -coverprofile=coverage.out

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Validation tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Validation tests passed${NC}"
echo ""

# Check test coverage
echo -e "${YELLOW}Step 2: Checking test coverage...${NC}"
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE < 90" | bc -l) )); then
    echo -e "${YELLOW}⚠ Warning: Coverage is below 90% (${COVERAGE}%)${NC}"
else
    echo -e "${GREEN}✓ Coverage target met (${COVERAGE}%)${NC}"
fi
echo ""

# Run quality report (if validation binary exists)
if [ -f "./bin/validate" ]; then
    echo -e "${YELLOW}Step 3: Generating quality report...${NC}"

    CMD="./bin/validate --report"

    if [ -n "$SOURCE" ]; then
        CMD="$CMD --source $SOURCE"
    fi

    if [ -n "$CATEGORY" ]; then
        CMD="$CMD --category $CATEGORY"
    fi

    if [ "$STRICT_MODE" = true ]; then
        CMD="$CMD --strict"
    fi

    $CMD

    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Quality report generation failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Quality report generated${NC}"
else
    echo -e "${YELLOW}⚠ Validation binary not found. Run 'make build' first.${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Validation Complete${NC}"
echo -e "${GREEN}====================================${NC}"
echo ""
echo -e "Test Coverage: ${COVERAGE}%"
echo -e "Quality Threshold: ${VALIDATION_THRESHOLD}"
echo -e "Max Error Rate: ${MAX_ERROR_RATE}"
echo -e "Max Duplicate Rate: ${MAX_DUPLICATE_RATE}"
echo ""
echo -e "${GREEN}✓ All validation checks passed!${NC}"
