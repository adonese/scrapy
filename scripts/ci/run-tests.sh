#!/bin/bash
# CI Test Execution Script
# Runs the complete test suite with proper error handling

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
TEST_TIMEOUT="${TEST_TIMEOUT:-10m}"
RACE_DETECTOR="${RACE_DETECTOR:-true}"
VERBOSE="${VERBOSE:-true}"
COVERAGE="${COVERAGE:-true}"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Run unit tests
run_unit_tests() {
    echo_section "Running Unit Tests"

    local args="-timeout ${TEST_TIMEOUT}"

    if [ "${VERBOSE}" = "true" ]; then
        args="$args -v"
    fi

    if [ "${RACE_DETECTOR}" = "true" ]; then
        args="$args -race"
    fi

    echo_info "Test arguments: $args"
    echo_info "Testing packages: ./pkg/... ./internal/models/... ./internal/handlers/... ./internal/repository/..."

    if go test $args ./pkg/... ./internal/models/... ./internal/handlers/... ./internal/repository/...; then
        echo_info "Unit tests passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "Unit tests failed ✗"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Run scraper tests
run_scraper_tests() {
    echo_section "Running Scraper Tests"

    local args="-timeout ${TEST_TIMEOUT}"

    if [ "${VERBOSE}" = "true" ]; then
        args="$args -v"
    fi

    if [ "${RACE_DETECTOR}" = "true" ]; then
        args="$args -race"
    fi

    echo_info "Testing scrapers..."

    if go test $args ./internal/scrapers/...; then
        echo_info "Scraper tests passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "Scraper tests failed ✗"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Run workflow tests
run_workflow_tests() {
    echo_section "Running Workflow Tests"

    local args="-timeout ${TEST_TIMEOUT}"

    if [ "${VERBOSE}" = "true" ]; then
        args="$args -v"
    fi

    if [ "${RACE_DETECTOR}" = "true" ]; then
        args="$args -race"
    fi

    echo_info "Testing workflows..."

    if go test $args ./internal/workflow/...; then
        echo_info "Workflow tests passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "Workflow tests failed ✗"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    echo_section "Running Integration Tests"

    if [ -z "${DATABASE_URL:-}" ]; then
        echo_warn "DATABASE_URL not set, skipping integration tests"
        return 0
    fi

    local args="-timeout ${TEST_TIMEOUT} -tags=integration"

    if [ "${VERBOSE}" = "true" ]; then
        args="$args -v"
    fi

    if [ "${RACE_DETECTOR}" = "true" ]; then
        args="$args -race"
    fi

    echo_info "Testing integration..."

    if go test $args ./...; then
        echo_info "Integration tests passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "Integration tests failed ✗"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Run code quality checks
run_quality_checks() {
    echo_section "Running Code Quality Checks"

    # go vet
    echo_info "Running go vet..."
    if go vet ./...; then
        echo_info "go vet passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "go vet failed ✗"
        ((TESTS_FAILED++))
        return 1
    fi

    # go fmt check
    echo_info "Checking code formatting..."
    local unformatted=$(gofmt -l .)
    if [ -z "$unformatted" ]; then
        echo_info "Code formatting check passed ✓"
        ((TESTS_PASSED++))
    else
        echo_error "Code formatting check failed ✗"
        echo_error "Unformatted files:"
        echo "$unformatted"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Generate test summary
print_summary() {
    echo_section "Test Summary"

    local total=$((TESTS_PASSED + TESTS_FAILED))
    echo_info "Total test suites: $total"
    echo_info "Passed: ${GREEN}$TESTS_PASSED${NC}"

    if [ $TESTS_FAILED -gt 0 ]; then
        echo_error "Failed: $TESTS_FAILED"
        return 1
    else
        echo_info "Failed: 0"
    fi
}

# Cleanup function
cleanup() {
    echo_info "Cleaning up..."
    # Add any cleanup logic here
}

# Set up trap for cleanup
trap cleanup EXIT

# Main execution
main() {
    echo_section "Starting CI Test Suite"

    local exit_code=0

    # Run all test suites
    run_unit_tests || exit_code=1
    run_scraper_tests || exit_code=1
    run_workflow_tests || exit_code=1
    run_integration_tests || exit_code=1
    run_quality_checks || exit_code=1

    # Print summary
    print_summary || exit_code=1

    if [ $exit_code -eq 0 ]; then
        echo_info "All tests passed! 🎉"
    else
        echo_error "Some tests failed 😞"
    fi

    exit $exit_code
}

main "$@"
