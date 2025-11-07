#!/bin/bash

# freshness-report.sh - Data freshness reporting script
# Checks freshness of data from all sources

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}  Data Freshness Report${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Configuration
OUTPUT_FILE="freshness-report.json"
ALERT_STALE=true

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --no-alerts)
            ALERT_STALE=false
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --output FILE      Output file for freshness report"
            echo "  --no-alerts        Disable alerts for stale data"
            echo "  --help             Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}Configuration:${NC}"
echo "  Output File: $OUTPUT_FILE"
echo "  Alert on Stale Data: $ALERT_STALE"
echo ""

# Run freshness check
echo -e "${YELLOW}Checking data freshness...${NC}"

if [ -f "./bin/freshness-checker" ]; then
    CMD="./bin/freshness-checker --output $OUTPUT_FILE"

    $CMD

    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Freshness check failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Freshness check complete${NC}"
    echo ""

    # Display summary if report exists
    if [ -f "$OUTPUT_FILE" ]; then
        echo -e "${YELLOW}Freshness Summary by Source:${NC}"
        echo ""

        if command -v jq &> /dev/null; then
            # Parse and display freshness status for each source
            SOURCES=$(jq -r 'keys[]' "$OUTPUT_FILE")

            FRESH_COUNT=0
            STALE_COUNT=0
            EXPIRED_COUNT=0

            for SOURCE in $SOURCES; do
                STATUS=$(jq -r ".\"$SOURCE\"" "$OUTPUT_FILE")

                if [ "$STATUS" = "FRESH" ]; then
                    echo -e "  ${GREEN}✓${NC} $SOURCE: $STATUS"
                    ((FRESH_COUNT++))
                elif [ "$STATUS" = "STALE" ]; then
                    echo -e "  ${YELLOW}⚠${NC} $SOURCE: $STATUS"
                    ((STALE_COUNT++))
                else
                    echo -e "  ${RED}✗${NC} $SOURCE: $STATUS"
                    ((EXPIRED_COUNT++))
                fi
            done

            echo ""
            echo -e "${YELLOW}Summary:${NC}"
            echo "  Fresh Sources: $FRESH_COUNT"
            echo "  Stale Sources: $STALE_COUNT"
            echo "  Expired Sources: $EXPIRED_COUNT"
            echo ""

            # Alert if there are stale sources
            if [ $STALE_COUNT -gt 0 ] || [ $EXPIRED_COUNT -gt 0 ]; then
                if [ "$ALERT_STALE" = true ]; then
                    echo -e "${YELLOW}⚠ WARNING: Some data sources are stale or expired!${NC}"
                    echo ""
                    echo "Stale sources should be updated to maintain data quality."
                    echo "Run the appropriate scrapers to refresh the data."
                    echo ""

                    # List specific stale sources
                    if [ $STALE_COUNT -gt 0 ]; then
                        echo -e "${YELLOW}Stale Sources:${NC}"
                        for SOURCE in $SOURCES; do
                            STATUS=$(jq -r ".\"$SOURCE\"" "$OUTPUT_FILE")
                            if [ "$STATUS" = "STALE" ]; then
                                echo "  - $SOURCE"
                            fi
                        done
                        echo ""
                    fi

                    if [ $EXPIRED_COUNT -gt 0 ]; then
                        echo -e "${RED}Expired Sources:${NC}"
                        for SOURCE in $SOURCES; do
                            STATUS=$(jq -r ".\"$SOURCE\"" "$OUTPUT_FILE")
                            if [ "$STATUS" = "EXPIRED" ]; then
                                echo "  - $SOURCE"
                            fi
                        done
                        echo ""
                    fi
                fi
            else
                echo -e "${GREEN}✓ All data sources are fresh!${NC}"
            fi
        else
            echo "  Report saved to: $OUTPUT_FILE"
            echo "  Install 'jq' for detailed summary"
        fi
    fi
else
    echo -e "${YELLOW}⚠ Freshness checker binary not found.${NC}"
    echo "Building freshness checker..."

    go build -o ./bin/freshness-checker ./cmd/freshness-checker/

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built freshness checker${NC}"
        echo "Please run this script again."
    else
        echo -e "${RED}❌ Failed to build freshness checker${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  Freshness Report Complete${NC}"
echo -e "${GREEN}====================================${NC}"
