# Operations Runbook - UAE Cost of Living

## Table of Contents

1. [Daily Operations](#daily-operations)
2. [Weekly Operations](#weekly-operations)
3. [Monthly Operations](#monthly-operations)
4. [Incident Response](#incident-response)
5. [Maintenance Tasks](#maintenance-tasks)
6. [Monitoring & Alerts](#monitoring--alerts)
7. [Troubleshooting Procedures](#troubleshooting-procedures)
8. [Emergency Contacts](#emergency-contacts)

## Daily Operations

### Morning Health Check (9:00 AM)

```bash
# 1. Check system health
./scripts/health-check.sh

# 2. Review scraper success rates from past 24h
docker logs --since 24h cost-of-living-worker | grep "scraper_success"

# 3. Check data freshness
./scripts/freshness-report.sh

# 4. Review validation reports
./scripts/validate-data.sh --summary

# 5. Check database health
psql -h localhost -U postgres -d cost_of_living -c "
SELECT
    source,
    COUNT(*) as records,
    MAX(recorded_at) as latest
FROM cost_data_points
WHERE recorded_at > NOW() - INTERVAL '24 hours'
GROUP BY source;"
```

### Expected Metrics

| Metric | Threshold | Action if Below |
|--------|-----------|----------------|
| Scraper Success Rate | > 95% | Investigate failures |
| Validation Pass Rate | > 95% | Review validation logs |
| Duplicate Rate | < 1% | Check deduplicate logic |
| Outlier Rate | < 2% | Review price ranges |
| Database Response Time | < 100ms | Check query performance |

### Data Quality Checks

```bash
# Check for anomalies
./scripts/check-outliers.sh --threshold 2.0

# Find duplicates
./scripts/find-duplicates.sh --auto-remove

# Generate quality report
./scripts/quality-report.sh --output daily-report-$(date +%Y%m%d).json
```

### Alert Review

1. Check Slack #alerts channel
2. Review GitHub issues tagged `scraper-failure`
3. Check Temporal UI for failed workflows: http://localhost:8233

## Weekly Operations

### Monday Morning (10:00 AM)

#### 1. Review Scraper Performance

```bash
# Generate weekly report
./scripts/weekly-scraper-report.sh

# Check individual scraper stats
for scraper in bayut dubizzle dewa sewa aadc rta careem; do
    echo "=== $scraper ==="
    go run cmd/scraper/main.go -scraper $scraper -dry-run
done
```

#### 2. Update Rate Limits (if needed)

Check if any scrapers are being rate-limited:

```bash
# Review logs for rate limit errors
docker logs cost-of-living-worker | grep "rate.limit\|429\|too many requests"

# Adjust rate limits in configuration
# Edit: internal/scrapers/<scraper>/<scraper>.go
```

#### 3. Check for Website Changes

```bash
# Test each scraper with current fixtures
make test-scrapers

# If tests fail, update fixtures:
# 1. Download latest HTML from target sites
# 2. Update test/fixtures/<scraper>/*.html
# 3. Update CSS selectors in internal/scrapers/<scraper>/parser.go
# 4. Re-run tests
```

#### 4. Database Maintenance

```bash
# Check database size
psql -h localhost -U postgres -d cost_of_living -c "
SELECT pg_size_pretty(pg_database_size('cost_of_living'));"

# Analyze query performance
psql -h localhost -U postgres -d cost_of_living -c "
SELECT
    query,
    calls,
    total_time,
    mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;"

# Vacuum if needed
psql -h localhost -U postgres -d cost_of_living -c "VACUUM ANALYZE cost_data_points;"
```

#### 5. Review Error Logs

```bash
# Check application logs
docker logs --since 7d cost-of-living-worker | grep ERROR

# Check Temporal workflow failures
# Open: http://localhost:8233
# Filter: ExecutionStatus=Failed, StartTime=Last 7 days
```

## Monthly Operations

### First Monday of Month

#### 1. Dependency Updates

```bash
# Check for Go module updates
go list -u -m all

# Update dependencies
go get -u ./...
go mod tidy

# Run full test suite
make test-all

# If tests pass, commit updates
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

#### 2. Security Audit

```bash
# Run security scan
make security-scan

# Check for vulnerable dependencies
go list -json -m all | nancy sleuth

# Review and update .env secrets
# Rotate database passwords if needed
```

#### 3. Performance Review

```bash
# Generate monthly performance report
./scripts/monthly-performance-report.sh

# Review and optimize slow queries
psql -h localhost -U postgres -d cost_of_living -c "
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC
LIMIT 20;"

# Check index usage
psql -h localhost -U postgres -d cost_of_living -c "
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY schemaname, tablename;"
```

#### 4. Backup Verification

```bash
# Verify backups exist
ls -lh /backups/cost_of_living/

# Test backup restoration (on test environment)
pg_restore -h test-db -U postgres -d cost_of_living_test /backups/latest.dump

# Verify restored data
psql -h test-db -U postgres -d cost_of_living_test -c "SELECT COUNT(*) FROM cost_data_points;"
```

#### 5. Documentation Updates

- Review and update README.md
- Update SCRAPERS.md with any changes
- Update API documentation
- Review and close stale GitHub issues

## Incident Response

### Scraper Failures

#### Symptoms
- Scraper returns 0 data points
- HTTP errors (403, 429, 500)
- Validation failures
- Timeout errors

#### Response Procedure

1. **Identify the failing scraper**
   ```bash
   docker logs cost-of-living-worker | grep ERROR | tail -50
   ```

2. **Check scraper status**
   ```bash
   go run cmd/scraper/main.go -scraper <name> -dry-run
   ```

3. **Diagnose the issue**
   - **Rate Limiting**: Increase delays, use proxy
   - **Website Changes**: Update CSS selectors, fixtures
   - **Anti-Bot**: Implement browser automation
   - **Network Issues**: Check connectivity, DNS

4. **Apply fix**
   ```bash
   # Update scraper code
   vi internal/scrapers/<scraper>/<scraper>.go

   # Test fix
   go test ./internal/scrapers/<scraper>/... -v

   # Deploy
   make build
   docker-compose restart worker
   ```

5. **Verify resolution**
   ```bash
   # Monitor logs
   docker logs -f cost-of-living-worker

   # Check metrics
   ./scripts/scraper-health-check.sh
   ```

6. **Document incident**
   - Create postmortem in docs/incidents/
   - Update runbook with lessons learned

### Data Quality Issues

#### Symptoms
- High validation failure rate
- Unexpected outliers
- Duplicate data points
- Stale data warnings

#### Response Procedure

1. **Identify affected data**
   ```bash
   ./scripts/validate-data.sh --source <scraper> --verbose
   ```

2. **Analyze the issue**
   ```bash
   # Check outliers
   ./scripts/check-outliers.sh --category <category>

   # Check duplicates
   ./scripts/find-duplicates.sh --source <scraper>

   # Check freshness
   ./scripts/freshness-report.sh
   ```

3. **Apply correction**
   ```sql
   -- Remove invalid data
   DELETE FROM cost_data_points
   WHERE source = '<scraper>'
   AND confidence < 0.5
   AND recorded_at > '2025-11-07';

   -- Re-run scraper
   ```

4. **Adjust validation rules** (if needed)
   ```go
   // Edit: internal/validation/rules.go
   // Adjust thresholds, add exceptions
   ```

5. **Monitor for recurrence**

### Database Performance Degradation

#### Symptoms
- Slow API responses
- Query timeouts
- High CPU/memory usage
- Connection pool exhaustion

#### Response Procedure

1. **Identify slow queries**
   ```sql
   SELECT
       pid,
       now() - query_start as duration,
       state,
       query
   FROM pg_stat_activity
   WHERE state != 'idle'
   ORDER BY duration DESC;
   ```

2. **Analyze query plans**
   ```sql
   EXPLAIN ANALYZE
   SELECT * FROM cost_data_points
   WHERE source = 'bayut'
   AND recorded_at > NOW() - INTERVAL '7 days';
   ```

3. **Optimize queries**
   - Add missing indexes
   - Update statistics
   - Adjust connection pool settings

4. **Apply fixes**
   ```sql
   -- Example: Add index
   CREATE INDEX CONCURRENTLY idx_cost_data_points_source_recorded
   ON cost_data_points(source, recorded_at DESC);

   -- Update statistics
   ANALYZE cost_data_points;
   ```

5. **Monitor improvement**
   ```bash
   ./scripts/db-performance-check.sh
   ```

### Workflow Failures

#### Symptoms
- Temporal workflows stuck in running state
- Activities timing out
- Retry exhaustion

#### Response Procedure

1. **Check Temporal UI**
   - Open: http://localhost:8233
   - Find failed workflow
   - Review execution history

2. **Diagnose failure**
   ```bash
   # Check worker logs
   docker logs cost-of-living-worker | grep -A 20 "workflow failed"
   ```

3. **Retry or reset**
   ```bash
   # From Temporal UI:
   # - Click on failed workflow
   # - Click "Reset" or "Terminate"
   # - Start new workflow execution
   ```

4. **Fix underlying issue**
   - Update workflow code
   - Adjust timeout settings
   - Fix activity implementation

5. **Redeploy**
   ```bash
   make build
   docker-compose restart worker
   ```

## Maintenance Tasks

### Database Cleanup

```bash
# Remove data older than 2 years
psql -h localhost -U postgres -d cost_of_living -c "
DELETE FROM cost_data_points
WHERE recorded_at < NOW() - INTERVAL '2 years';"

# Vacuum to reclaim space
psql -h localhost -U postgres -d cost_of_living -c "VACUUM FULL cost_data_points;"
```

### Log Rotation

```bash
# Rotate application logs (if not using Docker logging)
find /var/log/cost-of-living -name "*.log" -mtime +30 -delete

# Rotate Docker logs
docker-compose logs --tail=1000 worker > logs/worker-$(date +%Y%m%d).log
docker-compose logs --tail=1000 db > logs/db-$(date +%Y%m%d).log
```

### Backup Procedures

```bash
# Daily backup
pg_dump -h localhost -U postgres -Fc cost_of_living > /backups/cost_of_living-$(date +%Y%m%d).dump

# Weekly backup to S3 (if configured)
aws s3 cp /backups/cost_of_living-$(date +%Y%m%d).dump s3://backups/cost-of-living/

# Backup retention: Keep last 7 daily, 4 weekly, 12 monthly
find /backups -name "cost_of_living-*.dump" -mtime +7 -delete
```

## Monitoring & Alerts

### Metrics to Monitor

1. **Scraper Metrics**
   - Success rate per scraper
   - Items scraped per run
   - Execution time
   - Error rate by type

2. **Data Quality Metrics**
   - Validation pass rate
   - Duplicate rate
   - Outlier rate
   - Data freshness by source

3. **System Metrics**
   - API response time
   - Database query performance
   - Worker CPU/memory usage
   - Database connections

4. **Workflow Metrics**
   - Workflow success rate
   - Activity timeout rate
   - Retry count
   - Queue depth

### Alert Thresholds

| Alert | Threshold | Severity | Action |
|-------|-----------|----------|--------|
| Scraper failure | >5% failure rate | WARNING | Investigate within 4 hours |
| Scraper complete failure | 0 data points | ERROR | Immediate investigation |
| Data staleness | >7 days old | WARNING | Trigger scraper run |
| Data expiration | >30 days old | ERROR | Investigate scraper |
| Validation failure | >10% failure rate | WARNING | Review validation rules |
| Outlier spike | >5% outlier rate | WARNING | Check price ranges |
| Database slow query | >1s avg query time | WARNING | Optimize queries |
| Database down | Connection refused | CRITICAL | Immediate response |
| Workflow stuck | >30 min execution | WARNING | Check Temporal worker |
| Disk space | >85% usage | WARNING | Clean up old data |

### Alert Channels

- **Slack**: #alerts channel for all warnings/errors
- **Email**: ops-team@example.com for critical alerts
- **PagerDuty**: For critical database/service outages
- **GitHub Issues**: Automated issue creation for scraper failures

## Troubleshooting Procedures

### Common Issues

#### 1. "No data scraped"

```bash
# Check website accessibility
curl -I https://www.bayut.com

# Test scraper with verbose logging
LOG_LEVEL=debug go run cmd/scraper/main.go -scraper bayut

# Check CSS selectors
go test ./internal/scrapers/bayut -run TestParser -v
```

#### 2. "Database connection refused"

```bash
# Check PostgreSQL is running
docker-compose ps

# Check connection settings
env | grep DB_

# Test connection
psql -h localhost -U postgres -d cost_of_living -c "SELECT 1;"

# Restart database
docker-compose restart db
```

#### 3. "Temporal worker not processing workflows"

```bash
# Check worker logs
docker logs cost-of-living-worker

# Check Temporal server
docker-compose ps temporal

# Restart worker
docker-compose restart worker

# Check task queue
# Open: http://localhost:8233/namespaces/default/task-queues/scraper-tasks
```

#### 4. "High memory usage"

```bash
# Check container stats
docker stats cost-of-living-worker

# Check for memory leaks
go tool pprof http://localhost:6060/debug/pprof/heap

# Restart worker to clear memory
docker-compose restart worker
```

## Emergency Contacts

### On-Call Rotation
- **Primary**: ops-primary@example.com
- **Secondary**: ops-secondary@example.com
- **Escalation**: engineering-lead@example.com

### Vendor Contacts
- **AWS Support**: support.aws.com
- **Database Admin**: dba-team@example.com
- **Security Team**: security@example.com

### Useful Links
- **Temporal UI**: http://localhost:8233
- **Grafana**: http://localhost:3000
- **GitHub**: https://github.com/adonese/cost-of-living
- **Documentation**: /home/adonese/src/cost-of-living/docs/

---

**Last Updated**: 2025-11-07
**Version**: 1.0
**Owner**: Operations Team
