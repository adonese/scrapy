package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/pkg/logger"
)

type ScraperActivityResult struct {
	ScraperName    string
	ItemsFetched   int
	ItemsScraped   int
	ItemsValidated int
	ItemsSaved     int
	SaveFailures   int
	Duration       time.Duration
	Validation     services.ValidationSummary
	Errors         []string
}

type ScraperActivityDependencies struct {
	ScraperService *services.ScraperService
	Repository     repository.CostDataPointRepository
}

var dependencies *ScraperActivityDependencies

func SetActivityDependencies(deps *ScraperActivityDependencies) {
	dependencies = deps
}

func GetActivityDependencies() *ScraperActivityDependencies {
	return dependencies
}

// RunScraperActivity executes a scraper and stores the results
func RunScraperActivity(ctx context.Context, scraperName string) (*ScraperActivityResult, error) {
	logger.Info("Running scraper activity", "scraper", scraperName)

	activity.RecordHeartbeat(ctx, "starting")

	start := time.Now()

	serviceResult, err := dependencies.ScraperService.RunScraper(ctx, scraperName)

	result := &ScraperActivityResult{
		ScraperName: scraperName,
		Duration:    time.Since(start),
	}

	if serviceResult != nil {
		result.ItemsFetched = serviceResult.Fetched
		result.ItemsScraped = serviceResult.Fetched
		result.ItemsValidated = serviceResult.Validation.Valid
		result.ItemsSaved = serviceResult.Saved
		result.SaveFailures = serviceResult.SaveFailures
		result.Validation = serviceResult.Validation
		result.Duration = serviceResult.Duration
		result.Errors = make([]string, 0, len(serviceResult.Errors))
		for _, serviceErr := range serviceResult.Errors {
			if serviceErr != nil {
				result.Errors = append(result.Errors, serviceErr.Error())
			}
		}
	}

	activity.RecordHeartbeat(ctx, "completed")

	if err != nil {
		logger.Error("Scraper activity failed",
			"scraper", scraperName,
			"duration", result.Duration,
			"error", err)
		return result, fmt.Errorf("scraper failed: %w", err)
	}

	logger.Info("Scraper activity completed",
		"scraper", scraperName,
		"duration", result.Duration,
		"fetched", result.ItemsFetched,
		"scraped", result.ItemsScraped,
		"validated", result.ItemsValidated,
		"saved", result.ItemsSaved,
		"save_failures", result.SaveFailures)

	return result, nil
}

// CompensateFailedScrapeActivity performs compensation actions when a scrape fails
func CompensateFailedScrapeActivity(ctx context.Context, scraperName string) (bool, error) {
	logger.Info("Compensating failed scrape", "scraper", scraperName)

	// Compensation logic:
	// 1. Mark failed scrape in database (could add a scrape_runs table)
	// 2. Clean up partial data if needed
	// 3. Send alert notification

	// For now, just log the failure
	logger.Warn("Compensation executed for failed scrape",
		"scraper", scraperName,
		"timestamp", time.Now())

	// In a production system, you might want to:
	// - Record the failure in a failures table
	// - Delete any partial data from this scrape run
	// - Send notifications to operators
	// - Update monitoring dashboards

	return true, nil
}
