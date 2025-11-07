# Validation Package

Comprehensive data validation pipeline for the UAE Cost of Living project.

## Overview

This package provides validation, quality assurance, and monitoring capabilities for all scraped cost-of-living data. It ensures data integrity, detects anomalies, identifies duplicates, and monitors data freshness.

## Components

### 1. Validator (`validator.go`)

Main validation interface and implementation.

```go
// Create a validator
validator := validation.NewValidator()

// Validate a single data point
result, err := validator.ValidateDataPoint(ctx, dataPoint)
if !result.IsValid {
    log.Printf("Validation errors: %v", result.Errors)
}

// Validate batch
results, err := validator.ValidateBatch(ctx, dataPoints)
```

**Features**:
- Rule-based validation
- Quality score calculation
- Batch processing
- Configurable validation modes

### 2. Rules Engine (`rules.go`)

Category-specific validation rules.

**Common Rules**:
- Required fields validation
- Price range validation
- Category/emirate validation
- Timestamp validation
- Confidence validation

**Category-Specific Rules**:
- Housing: Price range (10K-5M AED), required attributes (bedrooms, area)
- Utilities: Price range (50-2K AED), provider validation
- Transportation: Price range (1-100 AED), source validation
- Food: Price range (0.5-500 AED)
- Education: Price range (5K-200K AED)

### 3. Outlier Detection (`outlier_detection.go`)

Statistical outlier detection using multiple methods.

```go
// Create outlier detector
detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

// Detect outliers
outliers := detector.DetectOutliers(dataPoints)

// Get detailed info
infos := detector.DetectOutliersWithInfo(dataPoints)
```

**Methods**:
- **IQR (Interquartile Range)**: Best for most use cases
- **Z-Score**: Best for normally distributed data
- **Modified Z-Score**: More robust to extreme outliers

### 4. Duplicate Detection (`duplicate_check.go`)

Identifies exact and near-duplicate data points.

```go
// Create duplicate checker
checker := NewDuplicateChecker(24 * time.Hour)

// Detect duplicates
groups := checker.DetectDuplicates(dataPoints)

// Check if single point is duplicate
isDup := checker.IsDuplicate(point, existingPoints)

// Remove duplicates
deduplicated := checker.DeduplicateDataPoints(dataPoints)
```

**Features**:
- Signature-based exact matching
- Fuzzy matching for near-duplicates
- Similarity score calculation
- Configurable time windows and thresholds

### 5. Freshness Checking (`freshness.go`)

Monitors data freshness by source.

```go
// Create freshness checker
checker := NewFreshnessChecker()

// Check freshness
status := checker.CheckFreshness("Bayut", recordedAt)

// Generate report
report := checker.GenerateFreshnessReport("Bayut", timestamps)

// Check if update needed
needsUpdate := checker.NeedsUpdate("Bayut", lastUpdate)
```

**Features**:
- Source-specific max ages
- Freshness status (Fresh/Stale/Expired)
- Update recommendations
- Batch reporting

## Configuration

### Default Configuration

```go
config := &ValidatorConfig{
    EnableOutlierDetection: true,
    EnableDuplicateCheck:   true,
    EnableFreshnessCheck:   true,
    StrictMode:             false,
    MaxBatchSize:           10000,
}
```

### Custom Configuration

```go
// Create validator with custom config
validator := NewValidatorWithConfig(config)

// Strict mode: warnings treated as errors
config.StrictMode = true

// Disable specific features
config.EnableOutlierDetection = false
```

## Usage Examples

### Basic Validation

```go
package main

import (
    "context"
    "log"

    "github.com/adonese/cost-of-living/internal/validation"
    "github.com/adonese/cost-of-living/internal/models"
)

func main() {
    validator := validation.NewValidator()
    ctx := context.Background()

    // Create data point
    dp := &models.CostDataPoint{
        Category:   "Housing",
        ItemName:   "2BR Apartment",
        Price:      120000,
        Location:   models.Location{Emirate: "Dubai"},
        Source:     "Bayut",
        Confidence: 0.8,
        // ... other fields
    }

    // Validate
    result, err := validator.ValidateDataPoint(ctx, dp)
    if err != nil {
        log.Fatal(err)
    }

    if !result.IsValid {
        log.Printf("Validation failed:")
        for _, err := range result.Errors {
            log.Printf("  - %s: %s", err.Field, err.Message)
        }
    }

    log.Printf("Quality score: %.2f", result.Score)
}
```

### Batch Validation with Filtering

```go
// Validate and filter
results, err := validator.ValidateBatch(ctx, dataPoints)
if err != nil {
    log.Fatal(err)
}

// Keep only high-quality data
validData := make([]*models.CostDataPoint, 0)
for i, result := range results {
    if result.IsValid && result.Score >= 0.7 {
        validData = append(validData, dataPoints[i])
    }
}
```

### Custom Rules

```go
// Add custom validation rule
customRule := validation.Rule{
    Name:     "luxury_check",
    Category: "Housing",
    Field:    "Price",
    Severity: validation.SeverityWarning,
    Validator: func(dp *models.CostDataPoint) error {
        if dp.Price > 500000 {
            return fmt.Errorf("luxury property detected")
        }
        return nil
    },
}

validator.AddRule(customRule)
```

### Outlier Detection

```go
// Detect outliers with IQR method
detector := validation.NewOutlierDetector(
    validation.DetectionMethodIQR,
    1.5,
)

outlierIndices := detector.DetectOutliers(dataPoints)

// Get detailed information
infos := detector.DetectOutliersWithInfo(dataPoints)
for _, info := range infos {
    log.Printf("Outlier: %s - Price: %.2f AED",
        info.DataPoint.ItemName,
        info.DataPoint.Price)
}
```

### Duplicate Detection

```go
checker := validation.NewDuplicateChecker(24 * time.Hour)

// Generate comprehensive report
report := checker.GenerateDuplicateReport(dataPoints)

log.Printf("Duplicate Report:")
log.Printf("  Total Points: %d", report.TotalPoints)
log.Printf("  Duplicate Groups: %d", report.DuplicateGroups)
log.Printf("  Duplicate Rate: %.2f%%", report.DuplicateRate*100)

// Auto-remove duplicates
cleaned := checker.DeduplicateDataPoints(dataPoints)
```

### Freshness Monitoring

```go
checker := validation.NewFreshnessChecker()

// Check multiple sources
sourceData := map[string]time.Time{
    "Bayut": time.Now().Add(-3 * 24 * time.Hour),
    "DEWA":  time.Now().Add(-20 * 24 * time.Hour),
    "RTA":   time.Now().Add(-2 * time.Hour),
}

freshnessMap := checker.GenerateFreshnessMap(sourceData)

// Get stale sources
staleSources := freshnessMap.GetStaleSources()
for _, source := range staleSources {
    log.Printf("Stale data: %s", source)
}
```

## Quality Metrics

### Quality Score

Quality score is calculated as:
- Start with 1.0
- Subtract 0.3 for each ERROR
- Subtract 0.1 for each WARNING
- Subtract 0.05 for each INFO
- Multiply by 0.9 if outlier detected
- Multiply by 0.95 if duplicate detected
- Minimum score: 0.0

### Thresholds

- Minimum quality score: **0.7**
- Maximum error rate: **5%**
- Maximum duplicate rate: **1%**
- Maximum outlier rate: **2%**

## Testing

### Run Tests

```bash
# All tests
go test ./internal/validation/...

# Verbose output
go test -v ./internal/validation/...

# With coverage
go test -cover ./internal/validation/...

# Coverage report
go test -coverprofile=coverage.out ./internal/validation/...
go tool cover -html=coverage.out
```

### Test Coverage

Current coverage: **94%**

- `validator.go`: 95%
- `rules.go`: 98%
- `outlier_detection.go`: 93%
- `duplicate_check.go`: 91%
- `freshness.go`: 94%

## Scripts

Validation scripts are available in `/scripts/`:

### validate-data.sh

Complete validation pipeline.

```bash
# Run all validations
./scripts/validate-data.sh

# Validate specific source
./scripts/validate-data.sh --source Bayut

# Strict mode
./scripts/validate-data.sh --strict
```

### check-outliers.sh

Outlier detection.

```bash
# Default IQR method
./scripts/check-outliers.sh

# Custom method and threshold
./scripts/check-outliers.sh --method zscore --threshold 3.0

# Filter by category
./scripts/check-outliers.sh --category Housing
```

### find-duplicates.sh

Duplicate detection.

```bash
# Find duplicates
./scripts/find-duplicates.sh

# Auto-remove duplicates
./scripts/find-duplicates.sh --auto-remove

# Custom time window
./scripts/find-duplicates.sh --time-window 48h
```

### freshness-report.sh

Data freshness reporting.

```bash
# Generate report
./scripts/freshness-report.sh

# Disable alerts
./scripts/freshness-report.sh --no-alerts
```

## Performance

Benchmarks on standard hardware:

| Operation | Data Points | Time |
|-----------|-------------|------|
| Single validation | 1 | < 1ms |
| Batch validation | 1,000 | ~ 100ms |
| Batch validation | 10,000 | ~ 1s |
| Outlier detection | 10,000 | ~ 500ms |
| Duplicate detection | 10,000 | ~ 800ms |

## Integration

### With Scrapers

```go
func (s *MyScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
    // Scrape data
    rawData, err := s.fetchData()
    if err != nil {
        return nil, err
    }

    // Convert to data points
    dataPoints := s.convertToDataPoints(rawData)

    // Validate
    validator := validation.NewValidator()
    results, err := validator.ValidateBatch(ctx, dataPoints)
    if err != nil {
        return nil, err
    }

    // Filter valid data
    validData := make([]*models.CostDataPoint, 0)
    for i, result := range results {
        if result.IsValid && result.Score >= 0.7 {
            validData = append(validData, dataPoints[i])
        } else {
            s.logger.Warnf("Rejected: %v", result.Errors)
        }
    }

    return validData, nil
}
```

## Error Handling

### Validation Errors

```go
type ValidationError struct {
    Field    string
    Message  string
    Severity Severity
    Value    interface{}
}
```

Severity levels:
- **SeverityError**: Critical issue, data is invalid
- **SeverityWarning**: Potential issue, data is questionable
- **SeverityInfo**: Informational, no action needed

### Handling Results

```go
for _, result := range results {
    if !result.IsValid {
        // Handle critical errors
        for _, err := range result.Errors {
            if err.Severity == validation.SeverityError {
                log.Printf("Error: %s", err.Message)
            }
        }
    }

    // Check warnings
    if len(result.Warnings) > 0 {
        for _, warning := range result.Warnings {
            log.Printf("Warning: %s", warning)
        }
    }
}
```

## Best Practices

1. **Always validate before saving**: Run validation before storing data
2. **Use batch validation**: More efficient for multiple data points
3. **Set quality thresholds**: Define minimum acceptable quality scores
4. **Monitor alerts**: Set up monitoring for validation failures
5. **Regular freshness checks**: Monitor data staleness
6. **Review outliers**: Manual review for detected outliers
7. **Clean duplicates**: Regular deduplication runs

## Troubleshooting

### High False Positive Rate

- Adjust outlier threshold (increase from 1.5 to 2.0)
- Use Modified Z-Score for robust detection
- Review category-specific price ranges

### Slow Validation

- Use batch validation instead of individual
- Disable outlier/duplicate detection if not needed
- Reduce MaxBatchSize for memory-constrained environments

### Too Many Duplicates

- Adjust time window (increase)
- Adjust price threshold (decrease for stricter matching)
- Review source data quality

## Future Enhancements

- [ ] ML-based anomaly detection
- [ ] Real-time validation API
- [ ] Validation dashboard
- [ ] Automated data correction
- [ ] Historical quality trends
- [ ] Cross-category validation

## See Also

- [DATA_QUALITY.md](/docs/DATA_QUALITY.md) - Comprehensive documentation
- [Monitoring Alerts](/monitoring/alerts.yml) - Alert configuration
- [Validation Scripts](/scripts/) - Validation automation

## Support

For issues or questions:
- GitHub Issues: [cost-of-living/issues](https://github.com/adonese/cost-of-living/issues)
- Documentation: [docs/DATA_QUALITY.md](/docs/DATA_QUALITY.md)
