# CI/CD Pipeline Guide

## Overview

This document provides comprehensive information about the CI/CD pipeline for the UAE Cost of Living project.

## Table of Contents

- [GitHub Actions Workflows](#github-actions-workflows)
- [Local Testing](#local-testing)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)
- [Pipeline Optimization](#pipeline-optimization)

---

## GitHub Actions Workflows

### 1. Main Test Pipeline (`.github/workflows/test.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

**Jobs:**
- **test**: Runs full test suite with coverage
  - Sets up Go 1.24
  - Starts PostgreSQL with TimescaleDB
  - Runs database migrations
  - Executes unit tests with race detector
  - Executes integration tests
  - Generates coverage report
  - Validates coverage threshold (70%)
  - Uploads coverage artifacts

- **lint**: Code quality checks
  - Runs golangci-lint
  - Checks code formatting

- **build**: Builds all binaries
  - API server
  - Worker
  - Scraper
  - Trigger utilities

**Duration:** ~4-5 minutes

**Required Secrets:** None (uses service containers)

---

### 2. Scraper Validation (`.github/workflows/scraper-validation.yml`)

**Triggers:**
- Schedule: Every 6 hours (cron: `0 */6 * * *`)
- Manual: `workflow_dispatch` with scraper selection

**Jobs:**
- **validate-bayut**: Tests Bayut scraper
- **validate-dubizzle**: Tests Dubizzle scraper
- **data-quality-check**: Validates data integrity
- **notify**: Sends alerts on failures

**Features:**
- Automated issue creation on persistent failures
- Data quality validation
- Scraper health monitoring
- Validation reports

**Duration:** ~5-10 minutes per scraper

**Required Secrets:**
- `SLACK_WEBHOOK_URL` (optional, for notifications)

---

### 3. Coverage Analysis (`.github/workflows/coverage.yml`)

**Triggers:**
- Push to `main`
- Pull requests to `main`

**Jobs:**
- **coverage**: Generates detailed coverage reports
  - Overall coverage percentage
  - Per-package coverage breakdown
  - HTML coverage report
  - Uncovered code analysis
  - Coverage badge generation
  - PR comments with coverage stats

- **benchmark**: Performance benchmarking (PR only)
  - Compares against base branch
  - Detects performance regressions
  - Comments results on PR

**Duration:** ~5-7 minutes

**Required Secrets:**
- `CODECOV_TOKEN` (optional, for Codecov integration)

---

### 4. Deployment Pipeline (`.github/workflows/deploy.yml`)

**Status:** DRAFT (manual deployment only)

**Triggers:**
- Manual: `workflow_dispatch` with environment selection

**Jobs:**
- **pre-deploy-checks**: Runs tests and security scans
- **build-images**: Builds and pushes Docker images
- **deploy-staging**: Deploys to staging environment
- **deploy-production**: Deploys to production environment
- **rollback**: Handles deployment failures

**Features:**
- Blue/green deployment support (TODO)
- Smoke tests after deployment
- Automatic rollback on failure
- Incident issue creation

**Required Secrets:**
- `GITHUB_TOKEN` (for image registry)
- Deployment credentials (TBD)

---

### 5. Dependency Management (`.github/dependabot.yml`)

**Configuration:**
- Go modules: Weekly updates (Mondays, 06:00 UTC)
- GitHub Actions: Weekly updates
- Docker images: Weekly updates

**Features:**
- Auto-grouping of minor/patch updates
- Auto-assign to maintainers
- Conventional commit messages
- Labeled PRs for easy filtering

---

## Local Testing

### Quick Start

```bash
# Run all tests
make test

# Run CI test suite locally
make test-ci

# Run with coverage
make test-coverage

# Run integration tests only
make test-integration

# Run benchmarks
make test-bench

# Validate scrapers
make validate-scrapers

# Run linters
make lint

# Security scan
make security-scan

# Run full CI validation
make ci-validate
```

### Using Docker Compose

```bash
# Start test database
make test-env-up

# Run tests against test database
DATABASE_URL=postgresql://test_user:test_password@localhost:5433/cost_of_living_test?sslmode=disable make test

# Clean up
make test-env-down
```

### Advanced Testing

```bash
# Run specific test
go test -v -run TestBayutParser ./internal/scrapers/bayut/...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks with memory profiling
go test -bench=. -benchmem -memprofile=mem.out ./...
```

---

## Environment Variables

### Required for CI

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Set by CI |
| `GO_ENV` | Environment (test, dev, prod) | `test` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `CODECOV_TOKEN` | Codecov upload token | - |
| `SLACK_WEBHOOK_URL` | Slack webhook for notifications | - |
| `COVERAGE_THRESHOLD` | Minimum coverage percentage | `70` |
| `TEST_TIMEOUT` | Test timeout duration | `10m` |
| `RACE_DETECTOR` | Enable race detector | `true` |

### Setting Environment Variables

**Local:**
```bash
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname?sslmode=disable"
```

**GitHub Actions:**
1. Go to repository Settings
2. Secrets and variables > Actions
3. New repository secret
4. Add name and value

---

## CI Scripts

### setup.sh

Sets up CI environment:
- Validates environment variables
- Waits for database readiness
- Installs Go dependencies
- Runs database migrations
- Installs additional tools

**Usage:**
```bash
./scripts/ci/setup.sh
```

### run-tests.sh

Executes comprehensive test suite:
- Unit tests
- Scraper tests
- Workflow tests
- Integration tests
- Code quality checks

**Usage:**
```bash
./scripts/ci/run-tests.sh
```

**Environment Variables:**
- `TEST_TIMEOUT`: Test timeout (default: 10m)
- `RACE_DETECTOR`: Enable race detection (default: true)
- `VERBOSE`: Verbose output (default: true)

### coverage.sh

Generates coverage reports:
- Overall coverage calculation
- Per-package coverage breakdown
- HTML report generation
- Uncovered code identification
- Threshold validation

**Usage:**
```bash
./scripts/ci/coverage.sh
```

**Environment Variables:**
- `COVERAGE_FILE`: Output file (default: coverage.out)
- `COVERAGE_THRESHOLD`: Minimum coverage (default: 70)
- `COVERAGE_DIR`: Output directory (default: coverage)

### validate.sh

Validates data quality:
- Database connection checks
- Schema validation
- Data integrity checks
- Freshness validation
- Scraper-specific validation

**Usage:**
```bash
./scripts/ci/validate.sh
```

**Environment Variables:**
- `DATABASE_URL`: Database connection string
- `MIN_DATA_POINTS`: Minimum records (default: 10)
- `MAX_AGE_HOURS`: Maximum data age (default: 48)

---

## Troubleshooting

### Common Issues

#### 1. Database Connection Failures

**Symptom:** Tests fail with "connection refused"

**Solution:**
```bash
# Check database is running
docker-compose ps

# Start database if needed
make db-up

# Check connection
pg_isready -d $DATABASE_URL
```

#### 2. Coverage Below Threshold

**Symptom:** CI fails with "coverage below 70%"

**Solution:**
- Add tests for uncovered code
- Check `coverage/uncovered.md` for specific functions
- Run `make test-coverage` locally to see detailed report

#### 3. Race Detector Failures

**Symptom:** Tests fail with "DATA RACE" errors

**Solution:**
- Fix concurrent access issues
- Add proper mutex locks
- Use channels for communication

#### 4. Migration Failures

**Symptom:** "migration failed" during setup

**Solution:**
```bash
# Check migration status
make migrate-version

# Reset database
make db-down
make db-up
make migrate
```

#### 5. Lint Failures

**Symptom:** golangci-lint reports issues

**Solution:**
```bash
# Run locally
make lint

# Auto-fix formatting
gofmt -w .

# Check specific linter
golangci-lint run --disable-all --enable=errcheck
```

### Getting Help

1. Check CI logs in GitHub Actions
2. Run tests locally to reproduce
3. Review error messages carefully
4. Check recent changes that might have broken tests
5. Consult team members

---

## Pipeline Optimization

### Current Performance

- Main test pipeline: ~4-5 minutes
- Coverage analysis: ~5-7 minutes
- Scraper validation: ~5-10 minutes per scraper

### Optimization Strategies

#### 1. Caching

**Go Modules:**
```yaml
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

**Build Cache:**
```yaml
- uses: actions/cache@v4
  with:
    path: ~/.cache/go-build
    key: ${{ runner.os }}-go-build-${{ hashFiles('**/*.go') }}
```

#### 2. Parallel Execution

- Test, lint, and build jobs run in parallel
- Multiple scraper validations run concurrently
- Independent test packages can be parallelized

#### 3. Test Selection

```bash
# Run only changed packages
go test $(go list ./... | grep -E 'changed_package')

# Skip slow tests locally
go test -short ./...
```

#### 4. Resource Limits

- Use tmpfs for database in tests
- Limit service container resources
- Use test-specific timeouts

---

## Best Practices

### Writing Tests

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always clean up test data
3. **Mocking**: Use mocks for external dependencies
4. **Fixtures**: Store test data in `test/fixtures/`
5. **Table-driven**: Use table-driven tests for multiple cases

### CI Workflow Development

1. **Test locally**: Always test workflow changes locally first
2. **Small commits**: Make incremental changes
3. **Fast feedback**: Keep pipelines fast (<10 minutes)
4. **Clear messages**: Use descriptive step names
5. **Monitoring**: Watch for flaky tests

### Coverage Goals

- **Overall**: 70% minimum, 80% target
- **Critical paths**: 90%+ for core business logic
- **New code**: All new features must have tests
- **Regression**: Add tests for all bug fixes

---

## Security Scanning

### Tools Integrated

1. **gosec**: Go security scanner
2. **staticcheck**: Static analysis
3. **Dependabot**: Dependency vulnerability scanning
4. **Container scanning**: Docker image vulnerabilities (TODO)

### Running Security Scans

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
make security-scan

# Generate report
gosec -fmt=json -out=report.json ./...
```

---

## Monitoring and Alerts

### Current Setup

- Automated issue creation on scraper failures
- PR comments with coverage and benchmark results
- GitHub Actions status badges (TODO)

### Future Enhancements

- Slack notifications for pipeline failures
- Email alerts for critical issues
- Metrics dashboard for pipeline performance
- Test flakiness detection

---

## Contributing

### Adding New Tests

1. Create test file: `*_test.go`
2. Write tests with proper assertions
3. Add fixtures if needed
4. Run locally: `make test`
5. Verify coverage: `make test-coverage`
6. Commit and push

### Modifying Workflows

1. Edit workflow file in `.github/workflows/`
2. Test with `act` (GitHub Actions locally) or push to branch
3. Create PR and verify workflow runs
4. Document changes in this guide

### Adding New Scrapers

1. Implement scraper with tests
2. Add validation in `scraper-validation.yml`
3. Update `validate.sh` script
4. Add to documentation

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Package](https://pkg.go.dev/testing)
- [golangci-lint](https://golangci-lint.run/)
- [Codecov](https://about.codecov.io/)
- [Dependabot](https://docs.github.com/en/code-security/dependabot)

---

**Last Updated:** 2025-11-07
**Maintainer:** DevOps Team
