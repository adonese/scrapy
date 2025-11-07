#!/bin/bash

# find-duplicates.sh - Duplicate detection script
# Identifies duplicate and near-duplicate data points

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}  Duplicate Detection Pipeline${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Configuration
TIME_WINDOW="24h"      # Time window for duplicate detection
PRICE_THRESHOLD=0.05   # 5% price difference threshold
AUTO_REMOVE=false
OUTPUT_FILE="duplicates-report.json"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --time-window)
            TIME_WINDOW="$2"
            shift 2
            ;;
        --price-threshold)
            PRICE_THRESHOLD="$2"
            shift 2
            ;;
        --auto-remove)
            AUTO_REMOVE=true
            shift
            ;;
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --time-window DURATION    Time window for duplicate detection (e.g., 24h, 48h)"
            echo "  --price-threshold VALUE   Price similarity threshold (0.0-1.0, default: 0.05)"
            echo "  --auto-remove            Automatically remove duplicates (keeps first occurrence)"
            echo "  --output FILE            Output file for duplicate report"
            echo "  --help                   Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}Configuration:${NC}"
echo "  Time Window: $TIME_WINDOW"
echo "  Price Threshold: $PRICE_THRESHOLD"
echo "  Auto Remove: $AUTO_REMOVE"
echo "  Output File: $OUTPUT_FILE"
echo ""

# Run duplicate detection
echo -e "${YELLOW}Scanning for duplicates...${NC}"

if [ -f "./bin/duplicate-finder" ]; then
    CMD="./bin/duplicate-finder --time-window $TIME_WINDOW --price-threshold $PRICE_THRESHOLD --output $OUTPUT_FILE"

    if [ "$AUTO_REMOVE" = true ]; then
        CMD="$CMD --auto-remove"
    fi

    $CMD

    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Duplicate detection failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Duplicate detection complete${NC}"
    echo ""

    # Display summary if report exists
    if [ -f "$OUTPUT_FILE" ]; then
        echo -e "${YELLOW}Duplicate Summary:${NC}"

        if command -v jq &> /dev/null; then
            TOTAL=$(jq '.total_points' "$OUTPUT_FILE")
            GROUPS=$(jq '.duplicate_groups' "$OUTPUT_FILE")
            DUPLICATES=$(jq '.total_duplicates' "$OUTPUT_FILE")
            RATE=$(jq '.duplicate_rate' "$OUTPUT_FILE")

            echo "  Total Data Points: $TOTAL"
            echo "  Duplicate Groups: $GROUPS"
            echo "  Total Duplicates: $DUPLICATES"
            echo "  Duplicate Rate: $(echo "$RATE * 100" | bc)%"
            echo ""

            if (( $(echo "$RATE > 0.01" | bc -l) )); then
                echo -e "${YELLOW}  ⚠ Duplicate rate exceeds 1% threshold${NC}"
                echo -e "  Consider running with --auto-remove to clean up duplicates"
            else
                echo -e "${GREEN}  ✓ Duplicate rate within acceptable range${NC}"
            fi

            if [ "$AUTO_REMOVE" = true ]; then
                echo ""
                echo -e "${GREEN}  ✓ Duplicates have been removed${NC}"
                REMOVED=$(jq '.removed_count // 0' "$OUTPUT_FILE")
                echo "  Removed: $REMOVED duplicate data points"
            fi
        else
            echo "  Report saved to: $OUTPUT_FILE"
            echo "  Install 'jq' for detailed summary"
        fi
    fi
else
    echo -e "${YELLOW}⚠ Duplicate finder binary not found.${NC}"
    echo "Building duplicate finder..."

    go build -o ./bin/duplicate-finder ./cmd/duplicate-finder/

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built duplicate finder${NC}"
        echo "Please run this script again."
    else
        echo -e "${RED}❌ Failed to build duplicate finder${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Duplicate Detection Complete${NC}"
echo -e "${GREEN}====================================${NC}"
