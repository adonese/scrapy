#!/bin/bash
# CI Data Validation Script
# Validates scraper data quality and integrity

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo_section() {
    echo -e "\n${BLUE}===================================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}===================================================${NC}\n"
}

# Configuration
DATABASE_URL="${DATABASE_URL:-}"
MIN_DATA_POINTS="${MIN_DATA_POINTS:-10}"
MAX_AGE_HOURS="${MAX_AGE_HOURS:-48}"

# Validation counters
CHECKS_PASSED=0
CHECKS_FAILED=0

# Check database connection
check_database() {
    echo_section "Database Connection Check"

    if [ -z "${DATABASE_URL}" ]; then
        echo_error "DATABASE_URL is not set"
        ((CHECKS_FAILED++))
        return 1
    fi

    if pg_isready -d "${DATABASE_URL}" > /dev/null 2>&1; then
        echo_info "Database connection OK ✓"
        ((CHECKS_PASSED++))
    else
        echo_error "Database connection failed ✗"
        ((CHECKS_FAILED++))
        return 1
    fi
}

# Validate data schema
validate_schema() {
    echo_section "Schema Validation"

    echo_info "Checking required tables..."

    local required_tables=("cost_data_points")
    local missing_tables=()

    for table in "${required_tables[@]}"; do
        if psql "${DATABASE_URL}" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '${table}');" | grep -q 't'; then
            echo_info "Table '${table}' exists ✓"
        else
            echo_error "Table '${table}' missing ✗"
            missing_tables+=("${table}")
        fi
    done

    if [ ${#missing_tables[@]} -eq 0 ]; then
        echo_info "Schema validation passed ✓"
        ((CHECKS_PASSED++))
    else
        echo_error "Schema validation failed. Missing tables: ${missing_tables[*]}"
        ((CHECKS_FAILED++))
        return 1
    fi
}

# Validate data integrity
validate_data_integrity() {
    echo_section "Data Integrity Validation"

    # Check for NULL values in required fields
    echo_info "Checking for NULL values in required fields..."

    local null_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE category IS NULL
           OR subcategory IS NULL
           OR value IS NULL
           OR collected_at IS NULL;
    " | xargs)

    if [ "$null_count" -eq 0 ]; then
        echo_info "No NULL values in required fields ✓"
        ((CHECKS_PASSED++))
    else
        echo_error "Found ${null_count} records with NULL values in required fields ✗"
        ((CHECKS_FAILED++))
    fi

    # Check for negative prices
    echo_info "Checking for invalid price values..."

    local negative_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE category = 'housing'
          AND value < 0;
    " | xargs)

    if [ "$negative_count" -eq 0 ]; then
        echo_info "No negative price values ✓"
        ((CHECKS_PASSED++))
    else
        echo_error "Found ${negative_count} records with negative prices ✗"
        ((CHECKS_FAILED++))
    fi

    # Check for unrealistic prices (e.g., > 1,000,000 AED for housing)
    echo_info "Checking for unrealistic price values..."

    local unrealistic_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE category = 'housing'
          AND value > 1000000;
    " | xargs)

    if [ "$unrealistic_count" -eq 0 ]; then
        echo_info "No unrealistic price values ✓"
        ((CHECKS_PASSED++))
    else
        echo_warn "Found ${unrealistic_count} records with prices > 1M AED (may need review)"
        # This is a warning, not a failure
    fi
}

# Validate data freshness
validate_data_freshness() {
    echo_section "Data Freshness Validation"

    echo_info "Checking data age..."

    local old_data_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE collected_at < NOW() - INTERVAL '${MAX_AGE_HOURS} hours';
    " | xargs)

    local total_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*) FROM cost_data_points;
    " | xargs)

    echo_info "Total records: ${total_count}"
    echo_info "Records older than ${MAX_AGE_HOURS} hours: ${old_data_count}"

    if [ "$total_count" -ge "$MIN_DATA_POINTS" ]; then
        echo_info "Minimum data points threshold met ✓"
        ((CHECKS_PASSED++))
    else
        echo_error "Insufficient data points (${total_count} < ${MIN_DATA_POINTS}) ✗"
        ((CHECKS_FAILED++))
    fi

    # Check for recent data
    local recent_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE collected_at > NOW() - INTERVAL '24 hours';
    " | xargs)

    if [ "$recent_count" -gt 0 ]; then
        echo_info "Found ${recent_count} records from last 24 hours ✓"
        ((CHECKS_PASSED++))
    else
        echo_warn "No recent data (last 24 hours)"
        # This might be OK during initial setup
    fi
}

# Validate scraper-specific data
validate_scraper_data() {
    echo_section "Scraper Data Validation"

    # Check Bayut data
    echo_info "Validating Bayut scraper data..."

    local bayut_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE source = 'bayut';
    " | xargs)

    echo_info "Bayut records: ${bayut_count}"

    if [ "$bayut_count" -gt 0 ]; then
        # Validate Bayut-specific fields
        local bayut_invalid=$(psql "${DATABASE_URL}" -t -c "
            SELECT COUNT(*)
            FROM cost_data_points
            WHERE source = 'bayut'
              AND (metadata->>'location' IS NULL
                   OR metadata->>'property_type' IS NULL);
        " | xargs)

        if [ "$bayut_invalid" -eq 0 ]; then
            echo_info "Bayut data validation passed ✓"
            ((CHECKS_PASSED++))
        else
            echo_error "Found ${bayut_invalid} Bayut records with missing metadata ✗"
            ((CHECKS_FAILED++))
        fi
    fi

    # Check Dubizzle data
    echo_info "Validating Dubizzle scraper data..."

    local dubizzle_count=$(psql "${DATABASE_URL}" -t -c "
        SELECT COUNT(*)
        FROM cost_data_points
        WHERE source = 'dubizzle';
    " | xargs)

    echo_info "Dubizzle records: ${dubizzle_count}"

    if [ "$dubizzle_count" -gt 0 ]; then
        # Validate Dubizzle-specific fields
        local dubizzle_invalid=$(psql "${DATABASE_URL}" -t -c "
            SELECT COUNT(*)
            FROM cost_data_points
            WHERE source = 'dubizzle'
              AND (metadata->>'location' IS NULL
                   OR metadata->>'listing_type' IS NULL);
        " | xargs)

        if [ "$dubizzle_invalid" -eq 0 ]; then
            echo_info "Dubizzle data validation passed ✓"
            ((CHECKS_PASSED++))
        else
            echo_error "Found ${dubizzle_invalid} Dubizzle records with missing metadata ✗"
            ((CHECKS_FAILED++))
        fi
    fi
}

# Validate data distribution
validate_data_distribution() {
    echo_section "Data Distribution Validation"

    echo_info "Checking data distribution across categories..."

    psql "${DATABASE_URL}" -c "
        SELECT
            category,
            COUNT(*) as count,
            ROUND(AVG(value), 2) as avg_value,
            ROUND(MIN(value), 2) as min_value,
            ROUND(MAX(value), 2) as max_value
        FROM cost_data_points
        GROUP BY category
        ORDER BY category;
    "

    echo_info "Checking data distribution across sources..."

    psql "${DATABASE_URL}" -c "
        SELECT
            source,
            COUNT(*) as count,
            MIN(collected_at) as oldest,
            MAX(collected_at) as newest
        FROM cost_data_points
        GROUP BY source
        ORDER BY source;
    "

    ((CHECKS_PASSED++))
}

# Generate validation report
generate_report() {
    echo_section "Validation Report"

    local total=$((CHECKS_PASSED + CHECKS_FAILED))
    echo_info "Total checks: $total"
    echo_info "Passed: ${GREEN}$CHECKS_PASSED${NC}"

    if [ $CHECKS_FAILED -gt 0 ]; then
        echo_error "Failed: $CHECKS_FAILED"
        return 1
    else
        echo_info "Failed: 0"
    fi

    # Save report to file
    cat > validation-report.txt <<EOF
Data Validation Report
Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

Summary:
- Total checks: $total
- Passed: $CHECKS_PASSED
- Failed: $CHECKS_FAILED

Status: $([ $CHECKS_FAILED -eq 0 ] && echo "PASSED" || echo "FAILED")
EOF

    echo_info "Report saved to validation-report.txt"
}

# Cleanup function
cleanup() {
    echo_info "Validation complete"
}

trap cleanup EXIT

# Main execution
main() {
    echo_section "Starting Data Validation"

    local exit_code=0

    check_database || exit_code=1
    validate_schema || exit_code=1
    validate_data_integrity || exit_code=1
    validate_data_freshness || exit_code=1
    validate_scraper_data || exit_code=1
    validate_data_distribution || true  # Don't fail on this

    generate_report || exit_code=1

    if [ $exit_code -eq 0 ]; then
        echo_info "All validations passed! ✓"
    else
        echo_error "Some validations failed"
    fi

    exit $exit_code
}

main "$@"
