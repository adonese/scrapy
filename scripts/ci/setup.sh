#!/bin/bash
# CI Environment Setup Script
# Sets up the environment for continuous integration testing

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check required environment variables
check_env() {
    echo_info "Checking environment variables..."

    if [ -z "${DATABASE_URL:-}" ]; then
        echo_error "DATABASE_URL is not set"
        exit 1
    fi

    echo_info "Environment variables OK"
}

# Wait for database to be ready
wait_for_db() {
    echo_info "Waiting for database to be ready..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if pg_isready -d "${DATABASE_URL}" > /dev/null 2>&1; then
            echo_info "Database is ready"
            return 0
        fi

        echo_warn "Database not ready, attempt $attempt/$max_attempts"
        sleep 2
        ((attempt++))
    done

    echo_error "Database failed to become ready after $max_attempts attempts"
    exit 1
}

# Install Go dependencies
install_deps() {
    echo_info "Installing Go dependencies..."

    if ! go mod download; then
        echo_error "Failed to download Go dependencies"
        exit 1
    fi

    if ! go mod verify; then
        echo_error "Go module verification failed"
        exit 1
    fi

    echo_info "Dependencies installed successfully"
}

# Run database migrations
run_migrations() {
    echo_info "Running database migrations..."

    if ! go run cmd/migrate/main.go up; then
        echo_error "Database migration failed"
        exit 1
    fi

    echo_info "Migrations completed successfully"
}

# Setup test data (if needed)
setup_test_data() {
    if [ "${SETUP_TEST_DATA:-false}" = "true" ]; then
        echo_info "Setting up test data..."
        # Add test data setup logic here if needed
        echo_info "Test data setup complete"
    fi
}

# Install additional tools
install_tools() {
    echo_info "Installing additional tools..."

    # Install golangci-lint if not present
    if ! command -v golangci-lint &> /dev/null; then
        echo_info "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
    fi

    # Install gosec if not present
    if ! command -v gosec &> /dev/null; then
        echo_info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi

    echo_info "Tools installed successfully"
}

# Verify setup
verify_setup() {
    echo_info "Verifying setup..."

    # Check Go version
    echo_info "Go version: $(go version)"

    # Check database connection
    if ! pg_isready -d "${DATABASE_URL}" > /dev/null 2>&1; then
        echo_error "Database connection verification failed"
        exit 1
    fi

    echo_info "Setup verification complete"
}

# Main setup flow
main() {
    echo_info "Starting CI environment setup..."

    check_env
    wait_for_db
    install_deps
    run_migrations
    setup_test_data
    install_tools
    verify_setup

    echo_info "CI environment setup complete!"
}

main "$@"
