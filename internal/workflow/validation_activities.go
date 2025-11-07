package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/validation"
	"github.com/adonese/cost-of-living/pkg/logger"
)

// ValidateRecentDataActivity validates recently scraped data
// This activity is called after batch scraping to ensure data quality
func ValidateRecentDataActivity(ctx context.Context) (*ValidationStats, error) {
	logger.Info("Starting validation of recent data")

	// Get activity dependencies
	deps := GetActivityDependencies()
	if deps == nil || deps.Repository == nil {
		return nil, fmt.Errorf("activity dependencies not set")
	}

	// Create validator
	validator := validation.NewValidator()

	// Get recent data points (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	filter := repository.ListFilter{
		StartDate: &since,
		Limit:     10000, // Maximum batch size
		Offset:    0,
	}
	dataPoints, err := deps.Repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent data: %w", err)
	}

	if len(dataPoints) == 0 {
		logger.Info("No recent data points to validate")
		return &ValidationStats{
			TotalValidated: 0,
			ValidCount:     0,
			InvalidCount:   0,
			QualityScore:   1.0,
		}, nil
	}

	logger.Info("Validating data points", "count", len(dataPoints))

	// Validate batch
	results, err := validator.ValidateBatch(ctx, dataPoints)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate statistics
	stats := &ValidationStats{
		TotalValidated: len(results),
		ValidCount:     0,
		InvalidCount:   0,
		QualityScore:   0.0,
	}

	totalScore := 0.0
	for _, result := range results {
		if result.IsValid {
			stats.ValidCount++
		} else {
			stats.InvalidCount++
		}
		totalScore += result.Score
	}

	if len(results) > 0 {
		stats.QualityScore = totalScore / float64(len(results))
	}

	logger.Info("Validation completed",
		"total", stats.TotalValidated,
		"valid", stats.ValidCount,
		"invalid", stats.InvalidCount,
		"quality_score", stats.QualityScore)

	return stats, nil
}

// ValidateScraperDataActivity validates data from a specific scraper
func ValidateScraperDataActivity(ctx context.Context, scraperName string, since time.Time) (*ValidationStats, error) {
	logger.Info("Validating scraper data", "scraper", scraperName)

	deps := GetActivityDependencies()
	if deps == nil || deps.Repository == nil {
		return nil, fmt.Errorf("activity dependencies not set")
	}

	validator := validation.NewValidator()

	// Get data points from this scraper (use List with filters)
	// Note: The current repository interface doesn't support filtering by source directly
	// We'll get all recent data and filter in memory
	filter := repository.ListFilter{
		StartDate: &since,
		Limit:     10000,
		Offset:    0,
	}
	allPoints, err := deps.Repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}

	// Filter by scraper name (source)
	dataPoints := make([]*models.CostDataPoint, 0)
	for _, dp := range allPoints {
		if dp.Source == scraperName {
			dataPoints = append(dataPoints, dp)
		}
	}

	if len(dataPoints) == 0 {
		return &ValidationStats{
			TotalValidated: 0,
			ValidCount:     0,
			InvalidCount:   0,
			QualityScore:   1.0,
		}, nil
	}

	// Validate batch
	results, err := validator.ValidateBatch(ctx, dataPoints)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate statistics
	stats := &ValidationStats{
		TotalValidated: len(results),
		ValidCount:     0,
		InvalidCount:   0,
		QualityScore:   0.0,
	}

	totalScore := 0.0
	for _, result := range results {
		if result.IsValid {
			stats.ValidCount++
		} else {
			stats.InvalidCount++
		}
		totalScore += result.Score
	}

	if len(results) > 0 {
		stats.QualityScore = totalScore / float64(len(results))
	}

	logger.Info("Scraper validation completed",
		"scraper", scraperName,
		"total", stats.TotalValidated,
		"valid", stats.ValidCount,
		"invalid", stats.InvalidCount,
		"quality_score", stats.QualityScore)

	return stats, nil
}

// CheckDataFreshnessActivity checks if scrapers are producing fresh data
func CheckDataFreshnessActivity(ctx context.Context) (map[string]time.Duration, error) {
	logger.Info("Checking data freshness")

	deps := GetActivityDependencies()
	if deps == nil || deps.Repository == nil {
		return nil, fmt.Errorf("activity dependencies not set")
	}

	freshnessChecker := validation.NewFreshnessChecker()

	// List of all scrapers to check
	scrapers := []string{
		"bayut-Dubai", "bayut-Sharjah", "bayut-Ajman", "bayut-Abu Dhabi",
		"dubizzle-Dubai-apartmentflat", "dubizzle-Sharjah-apartmentflat",
		"dubizzle-Ajman-apartmentflat", "dubizzle-Abu Dhabi-apartmentflat",
		"dubizzle-Dubai-bedspace", "dubizzle-Dubai-roomspace",
		"dewa", "sewa", "aadc", "rta", "careem",
	}

	freshness := make(map[string]time.Duration)
	staleCount := 0

	// Get all data points (we'll find the latest for each scraper)
	filter := repository.ListFilter{
		Limit:  10000,
		Offset: 0,
	}
	allPoints, err := deps.Repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get data points: %w", err)
	}

	// Group by scraper and find latest
	latestBySource := make(map[string]*models.CostDataPoint)
	for _, dp := range allPoints {
		existing, exists := latestBySource[dp.Source]
		if !exists || dp.RecordedAt.After(existing.RecordedAt) {
			latestBySource[dp.Source] = dp
		}
	}

	for _, scraper := range scrapers {
		latest, exists := latestBySource[scraper]
		if !exists {
			logger.Warn("No data found for scraper", "scraper", scraper)
			freshness[scraper] = -1 // Indicates no data
			continue
		}

		age := time.Since(latest.RecordedAt)
		freshness[scraper] = age

		// Check if stale
		status := freshnessChecker.CheckFreshness(scraper, latest.RecordedAt)
		if status == validation.FreshnessStale {
			staleCount++
			logger.Warn("Stale data detected",
				"scraper", scraper,
				"age", age,
				"last_update", latest.RecordedAt)
		}
	}

	logger.Info("Freshness check completed",
		"total_scrapers", len(scrapers),
		"stale_count", staleCount)

	return freshness, nil
}

// DetectOutliersActivity detects statistical outliers in recent data
func DetectOutliersActivity(ctx context.Context, category string) ([]string, error) {
	logger.Info("Detecting outliers", "category", category)

	deps := GetActivityDependencies()
	if deps == nil || deps.Repository == nil {
		return nil, fmt.Errorf("activity dependencies not set")
	}

	// Get recent data for the category
	since := time.Now().Add(-7 * 24 * time.Hour) // Last week
	filter := repository.ListFilter{
		Category:  category,
		StartDate: &since,
		Limit:     10000,
		Offset:    0,
	}
	dataPoints, err := deps.Repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get category data: %w", err)
	}

	if len(dataPoints) == 0 {
		return []string{}, nil
	}

	// Detect outliers
	detector := validation.NewOutlierDetector(validation.DetectionMethodIQR, 1.5)
	outlierIndices := detector.DetectOutliers(dataPoints)

	// Get outlier IDs
	outlierIDs := make([]string, 0, len(outlierIndices))
	for _, idx := range outlierIndices {
		if idx < len(dataPoints) {
			outlierIDs = append(outlierIDs, dataPoints[idx].ID)
		}
	}

	logger.Info("Outlier detection completed",
		"category", category,
		"total_points", len(dataPoints),
		"outliers", len(outlierIDs))

	return outlierIDs, nil
}

// CheckDuplicatesActivity checks for duplicate data points
func CheckDuplicatesActivity(ctx context.Context) (int, error) {
	logger.Info("Checking for duplicates")

	deps := GetActivityDependencies()
	if deps == nil || deps.Repository == nil {
		return 0, fmt.Errorf("activity dependencies not set")
	}

	// Get recent data points
	since := time.Now().Add(-24 * time.Hour)
	filter := repository.ListFilter{
		StartDate: &since,
		Limit:     10000,
		Offset:    0,
	}
	dataPoints, err := deps.Repository.List(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent data: %w", err)
	}

	if len(dataPoints) == 0 {
		return 0, nil
	}

	// Check for duplicates
	duplicateChecker := validation.NewDuplicateChecker(24 * time.Hour)
	duplicateGroups := duplicateChecker.DetectDuplicates(dataPoints)

	totalDuplicates := 0
	for _, group := range duplicateGroups {
		totalDuplicates += len(group.Indices) - 1 // Subtract 1 for the original
	}

	logger.Info("Duplicate check completed",
		"total_points", len(dataPoints),
		"duplicate_groups", len(duplicateGroups),
		"total_duplicates", totalDuplicates)

	return totalDuplicates, nil
}
