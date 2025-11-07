#!/bin/bash
# CI Coverage Analysis Script
# Generates and validates test coverage reports

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
COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"
COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-70}"
COVERAGE_HTML="${COVERAGE_HTML:-coverage.html}"
COVERAGE_DIR="${COVERAGE_DIR:-coverage}"

# Create coverage directory
mkdir -p "${COVERAGE_DIR}"

# Generate coverage report
generate_coverage() {
    echo_section "Generating Coverage Report"

    echo_info "Running tests with coverage..."

    if go test -v -race -coverprofile="${COVERAGE_DIR}/${COVERAGE_FILE}" -covermode=atomic ./...; then
        echo_info "Coverage data generated successfully"
    else
        echo_error "Failed to generate coverage data"
        return 1
    fi
}

# Generate HTML coverage report
generate_html_report() {
    echo_section "Generating HTML Coverage Report"

    if [ ! -f "${COVERAGE_DIR}/${COVERAGE_FILE}" ]; then
        echo_error "Coverage file not found: ${COVERAGE_DIR}/${COVERAGE_FILE}"
        return 1
    fi

    echo_info "Generating HTML report..."

    if go tool cover -html="${COVERAGE_DIR}/${COVERAGE_FILE}" -o "${COVERAGE_DIR}/${COVERAGE_HTML}"; then
        echo_info "HTML report generated: ${COVERAGE_DIR}/${COVERAGE_HTML}"
    else
        echo_error "Failed to generate HTML report"
        return 1
    fi
}

# Calculate total coverage
calculate_total_coverage() {
    if [ ! -f "${COVERAGE_DIR}/${COVERAGE_FILE}" ]; then
        echo_error "Coverage file not found: ${COVERAGE_DIR}/${COVERAGE_FILE}"
        return 1
    fi

    local total=$(go tool cover -func="${COVERAGE_DIR}/${COVERAGE_FILE}" | grep total | awk '{print $3}' | sed 's/%//')
    echo "$total"
}

# Generate per-package coverage report
generate_package_report() {
    echo_section "Per-Package Coverage Report"

    if [ ! -f "${COVERAGE_DIR}/${COVERAGE_FILE}" ]; then
        echo_error "Coverage file not found: ${COVERAGE_DIR}/${COVERAGE_FILE}"
        return 1
    fi

    echo_info "Calculating per-package coverage..."

    # Create detailed report
    go tool cover -func="${COVERAGE_DIR}/${COVERAGE_FILE}" > "${COVERAGE_DIR}/coverage-detailed.txt"

    # Generate package summary
    echo "# Coverage Report" > "${COVERAGE_DIR}/coverage-summary.md"
    echo "" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "" >> "${COVERAGE_DIR}/coverage-summary.md"

    local total=$(calculate_total_coverage)
    echo "## Overall Coverage" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "**Total Coverage:** ${total}%" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "" >> "${COVERAGE_DIR}/coverage-summary.md"

    echo "## Per-Package Coverage" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "| Package | Coverage |" >> "${COVERAGE_DIR}/coverage-summary.md"
    echo "|---------|----------|" >> "${COVERAGE_DIR}/coverage-summary.md"

    # Calculate per-package coverage
    awk '
    {
        if ($1 != "total:") {
            # Extract package name from file path
            split($1, parts, "/");
            pkg = "";
            for (i = 1; i < length(parts); i++) {
                if (i > 1) pkg = pkg "/";
                pkg = pkg parts[i];
            }

            # Extract coverage percentage
            gsub(/%/, "", $3);
            coverage[pkg] += $3;
            count[pkg]++;
        }
    }
    END {
        for (p in coverage) {
            if (count[p] > 0) {
                avg = coverage[p] / count[p];
                printf "| %s | %.1f%% |\n", p, avg;
            }
        }
    }
    ' "${COVERAGE_DIR}/coverage-detailed.txt" | sort >> "${COVERAGE_DIR}/coverage-summary.md"

    echo ""
    cat "${COVERAGE_DIR}/coverage-summary.md"
}

# Check coverage threshold
check_threshold() {
    echo_section "Coverage Threshold Check"

    local total=$(calculate_total_coverage)

    echo_info "Total coverage: ${total}%"
    echo_info "Required threshold: ${COVERAGE_THRESHOLD}%"

    # Use bc for floating point comparison
    if command -v bc &> /dev/null; then
        if (( $(echo "$total < $COVERAGE_THRESHOLD" | bc -l) )); then
            echo_error "Coverage ${total}% is below threshold ${COVERAGE_THRESHOLD}%"
            return 1
        fi
    else
        # Fallback to integer comparison
        local total_int=${total%.*}
        if [ "$total_int" -lt "$COVERAGE_THRESHOLD" ]; then
            echo_error "Coverage ${total}% is below threshold ${COVERAGE_THRESHOLD}%"
            return 1
        fi
    fi

    echo_info "Coverage ${total}% meets threshold ${COVERAGE_THRESHOLD}% ✓"
}

# Identify uncovered code
find_uncovered() {
    echo_section "Uncovered Code Analysis"

    if [ ! -f "${COVERAGE_DIR}/${COVERAGE_FILE}" ]; then
        echo_error "Coverage file not found: ${COVERAGE_DIR}/${COVERAGE_FILE}"
        return 1
    fi

    echo_info "Finding uncovered code..."

    # Find functions with 0% coverage
    echo "## Uncovered Functions" > "${COVERAGE_DIR}/uncovered.md"
    echo "" >> "${COVERAGE_DIR}/uncovered.md"

    go tool cover -func="${COVERAGE_DIR}/${COVERAGE_FILE}" | awk '
    $3 == "0.0%" {
        print "- `" $2 "` in " $1
    }
    ' >> "${COVERAGE_DIR}/uncovered.md"

    local uncovered_count=$(go tool cover -func="${COVERAGE_DIR}/${COVERAGE_FILE}" | awk '$3 == "0.0%"' | wc -l)

    if [ "$uncovered_count" -gt 0 ]; then
        echo_warn "Found ${uncovered_count} uncovered functions"
        cat "${COVERAGE_DIR}/uncovered.md"
    else
        echo_info "All functions have some test coverage ✓"
    fi
}

# Generate badge URL (for README)
generate_badge() {
    local total=$(calculate_total_coverage)
    local color="red"

    # Determine badge color based on coverage
    if (( $(echo "$total >= 80" | bc -l 2>/dev/null || echo "0") )); then
        color="brightgreen"
    elif (( $(echo "$total >= 70" | bc -l 2>/dev/null || echo "0") )); then
        color="green"
    elif (( $(echo "$total >= 60" | bc -l 2>/dev/null || echo "0") )); then
        color="yellow"
    elif (( $(echo "$total >= 50" | bc -l 2>/dev/null || echo "0") )); then
        color="orange"
    fi

    local badge_url="https://img.shields.io/badge/coverage-${total}%25-${color}"
    echo_info "Badge URL: ${badge_url}"
    echo "${badge_url}" > "${COVERAGE_DIR}/badge-url.txt"
}

# Cleanup function
cleanup() {
    echo_info "Coverage analysis complete"
}

trap cleanup EXIT

# Main execution
main() {
    echo_section "Starting Coverage Analysis"

    local exit_code=0

    generate_coverage || exit_code=1
    generate_html_report || exit_code=1
    generate_package_report || exit_code=1
    find_uncovered || true  # Don't fail on this
    check_threshold || exit_code=1
    generate_badge || true  # Don't fail on this

    if [ $exit_code -eq 0 ]; then
        echo_info "Coverage analysis passed! ✓"
    else
        echo_error "Coverage analysis failed"
    fi

    exit $exit_code
}

main "$@"
