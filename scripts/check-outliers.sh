#!/bin/bash

# check-outliers.sh - Outlier detection script
# Detects statistical outliers in cost-of-living data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}  Outlier Detection Pipeline${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Configuration
METHOD="iqr"           # Detection method: iqr, zscore, modified-zscore
THRESHOLD=1.5          # Outlier threshold
CATEGORY=""
SOURCE=""
OUTPUT_FILE="outliers-report.json"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --method)
            METHOD="$2"
            shift 2
            ;;
        --threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        --category)
            CATEGORY="$2"
            shift 2
            ;;
        --source)
            SOURCE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --method METHOD         Detection method (iqr, zscore, modified-zscore)"
            echo "  --threshold THRESHOLD   Outlier threshold (default: 1.5 for IQR, 3.0 for z-score)"
            echo "  --category CATEGORY     Filter by category"
            echo "  --source SOURCE         Filter by source"
            echo "  --output FILE           Output file for outlier report"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}Configuration:${NC}"
echo "  Detection Method: $METHOD"
echo "  Threshold: $THRESHOLD"
[ -n "$CATEGORY" ] && echo "  Category Filter: $CATEGORY"
[ -n "$SOURCE" ] && echo "  Source Filter: $SOURCE"
echo "  Output File: $OUTPUT_FILE"
echo ""

# Run outlier detection
echo -e "${YELLOW}Running outlier detection...${NC}"

if [ -f "./bin/outlier-detector" ]; then
    CMD="./bin/outlier-detector --method $METHOD --threshold $THRESHOLD --output $OUTPUT_FILE"

    [ -n "$CATEGORY" ] && CMD="$CMD --category $CATEGORY"
    [ -n "$SOURCE" ] && CMD="$CMD --source $SOURCE"

    $CMD

    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Outlier detection failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Outlier detection complete${NC}"
    echo ""

    # Display summary if report exists
    if [ -f "$OUTPUT_FILE" ]; then
        echo -e "${YELLOW}Outlier Summary:${NC}"
        # Parse JSON and display summary (requires jq)
        if command -v jq &> /dev/null; then
            TOTAL=$(jq '.total_points' "$OUTPUT_FILE")
            OUTLIERS=$(jq '.outlier_count' "$OUTPUT_FILE")
            RATE=$(jq '.outlier_rate' "$OUTPUT_FILE")

            echo "  Total Data Points: $TOTAL"
            echo "  Outliers Detected: $OUTLIERS"
            echo "  Outlier Rate: $(echo "$RATE * 100" | bc)%"

            if (( $(echo "$RATE > 0.02" | bc -l) )); then
                echo -e "${YELLOW}  ⚠ Outlier rate exceeds 2% threshold${NC}"
            else
                echo -e "${GREEN}  ✓ Outlier rate within acceptable range${NC}"
            fi
        else
            echo "  Report saved to: $OUTPUT_FILE"
            echo "  Install 'jq' for detailed summary"
        fi
    fi
else
    echo -e "${YELLOW}⚠ Outlier detector binary not found.${NC}"
    echo "Building outlier detector..."

    # Try to build the detector
    go build -o ./bin/outlier-detector ./cmd/outlier-detector/

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built outlier detector${NC}"
        echo "Please run this script again."
    else
        echo -e "${RED}❌ Failed to build outlier detector${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Outlier Detection Complete${NC}"
echo -e "${GREEN}====================================${NC}"
