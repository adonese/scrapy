#!/bin/bash

echo "Testing metrics endpoint..."
echo ""

# Make a few API requests
echo "1. Making requests to generate metrics..."
curl -s http://localhost:8080/health > /dev/null && echo "  - /health request sent"
curl -s http://localhost:8080/api/v1/cost-data-points?limit=5 > /dev/null && echo "  - /api/v1/cost-data-points request sent"
curl -s http://localhost:8080/health > /dev/null && echo "  - /health request sent (again)"

echo ""
echo "2. Checking metrics endpoint..."
echo ""

# Check metrics endpoint
METRICS=$(curl -s http://localhost:8080/metrics | grep "http_requests_total")

if [ -z "$METRICS" ]; then
    echo "ERROR: No metrics found!"
    exit 1
else
    echo "SUCCESS: Metrics found!"
    echo ""
    echo "Sample metrics:"
    echo "$METRICS" | head -n 10
    echo ""
    echo "All checks passed!"
fi
