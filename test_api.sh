#!/bin/bash

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Cost of Living API - Iteration 1.4${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

BASE_URL="http://localhost:8080"

# Function to print test header
print_test() {
    echo -e "\n${BLUE}TEST: $1${NC}"
    echo "----------------------------------------"
}

# Function to print result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
    fi
}

# 1. Test Health Endpoint
print_test "1. Health Check"
response=$(curl -s "${BASE_URL}/health")
echo "Response: $response"
if echo "$response" | grep -q '"status":"ok"'; then
    print_result 0
else
    print_result 1
fi

# 2. Test Create Cost Data Point
print_test "2. Create Cost Data Point (Housing)"
response=$(curl -s -X POST "${BASE_URL}/api/v1/cost-data-points" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Housing",
    "item_name": "1BR Apartment Marina",
    "price": 85000,
    "location": {"emirate": "Dubai", "city": "Dubai", "area": "Marina"},
    "source": "manual"
  }')
echo "Response: $response"
if echo "$response" | grep -q '"id"'; then
    ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    RECORDED_AT=$(echo "$response" | grep -o '"recorded_at":"[^"]*' | cut -d'"' -f4)
    echo "Created ID: $ID"
    echo "Recorded At: $RECORDED_AT"
    print_result 0
else
    print_result 1
fi

# 3. Test Create with Validation Error
print_test "3. Create with Validation Error (Missing Required Fields)"
response=$(curl -s -X POST "${BASE_URL}/api/v1/cost-data-points" \
  -H "Content-Type: application/json" \
  -d '{"category": "Housing", "price": 85000}')
echo "Response: $response"
if echo "$response" | grep -q '"error":"Bad Request"'; then
    print_result 0
else
    print_result 1
fi

# 4. Test Get by ID (with recorded_at)
if [ ! -z "$ID" ] && [ ! -z "$RECORDED_AT" ]; then
    print_test "4. Get Cost Data Point by ID"
    # Convert timestamp to URL-safe format (RFC3339)
    ENCODED_TIME=$(echo "$RECORDED_AT" | sed 's/+/%2B/g')
    response=$(curl -s "${BASE_URL}/api/v1/cost-data-points/${ID}?recorded_at=${ENCODED_TIME}")
    echo "Response: $response"
    if echo "$response" | grep -q "\"id\":\"$ID\""; then
        print_result 0
    else
        print_result 1
    fi
fi

# 5. Test List All
print_test "5. List All Cost Data Points"
response=$(curl -s "${BASE_URL}/api/v1/cost-data-points")
echo "Response: $response"
if echo "$response" | grep -q '"data"'; then
    print_result 0
else
    print_result 1
fi

# 6. Test List with Filters
print_test "6. List with Category Filter"
response=$(curl -s "${BASE_URL}/api/v1/cost-data-points?category=Housing&limit=5")
echo "Response: $response"
if echo "$response" | grep -q '"category":"Housing"'; then
    print_result 0
else
    print_result 1
fi

# 7. Create another record for more testing
print_test "7. Create Another Cost Data Point (Food)"
response=$(curl -s -X POST "${BASE_URL}/api/v1/cost-data-points" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Food",
    "item_name": "Restaurant Meal",
    "price": 50,
    "location": {"emirate": "Abu Dhabi", "city": "Abu Dhabi"},
    "source": "survey"
  }')
echo "Response: $response"
if echo "$response" | grep -q '"id"'; then
    FOOD_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    FOOD_RECORDED_AT=$(echo "$response" | grep -o '"recorded_at":"[^"]*' | cut -d'"' -f4)
    echo "Created Food ID: $FOOD_ID"
    print_result 0
else
    print_result 1
fi

# 8. Test List with Emirate Filter
print_test "8. List with Emirate Filter (Abu Dhabi)"
response=$(curl -s "${BASE_URL}/api/v1/cost-data-points?emirate=Abu%20Dhabi")
echo "Response: $response"
if echo "$response" | grep -q '"emirate":"Abu Dhabi"'; then
    print_result 0
else
    print_result 1
fi

# 9. Test Update
if [ ! -z "$ID" ] && [ ! -z "$RECORDED_AT" ]; then
    print_test "9. Update Cost Data Point"
    ENCODED_TIME=$(echo "$RECORDED_AT" | sed 's/+/%2B/g')
    response=$(curl -s -X PUT "${BASE_URL}/api/v1/cost-data-points/${ID}?recorded_at=${ENCODED_TIME}" \
      -H "Content-Type: application/json" \
      -d '{"price": 90000}')
    echo "Response: $response"
    if echo "$response" | grep -q '"price":90000'; then
        print_result 0
    else
        print_result 1
    fi
fi

# 10. Test Delete
if [ ! -z "$FOOD_ID" ] && [ ! -z "$FOOD_RECORDED_AT" ]; then
    print_test "10. Delete Cost Data Point"
    ENCODED_TIME=$(echo "$FOOD_RECORDED_AT" | sed 's/+/%2B/g')
    status_code=$(curl -s -o /dev/null -w "%{http_code}" \
      -X DELETE "${BASE_URL}/api/v1/cost-data-points/${FOOD_ID}?recorded_at=${ENCODED_TIME}")
    echo "HTTP Status: $status_code"
    if [ "$status_code" = "204" ]; then
        print_result 0
    else
        print_result 1
    fi
fi

# 11. Test Pagination
print_test "11. Test Pagination (limit=2, offset=0)"
response=$(curl -s "${BASE_URL}/api/v1/cost-data-points?limit=2&offset=0")
echo "Response: $response"
if echo "$response" | grep -q '"limit":2'; then
    print_result 0
else
    print_result 1
fi

# 12. Test Not Found
print_test "12. Get Non-existent Record"
response=$(curl -s "${BASE_URL}/api/v1/cost-data-points/nonexistent-id?recorded_at=2025-01-01T00:00:00Z")
echo "Response: $response"
if echo "$response" | grep -q '"error"'; then
    print_result 0
else
    print_result 1
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}API Testing Complete${NC}"
echo -e "${BLUE}========================================${NC}"
