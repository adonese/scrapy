package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

type ScraperActivityResult struct {
	ItemsScraped int
	ItemsSaved   int
	Duration     time.Duration
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

	// Send heartbeat to let Temporal know we're alive
	activity.RecordHeartbeat(ctx, "starting")

	start := time.Now()

	// Get count of items before scraping
	filterBefore := repository.ListFilter{
		Limit:  100,
		Offset: 0,
	}
	itemsBefore, _ := dependencies.Repository.List(ctx, filterBefore)
	countBefore := len(itemsBefore)

	// Run the scraper
	err := dependencies.ScraperService.RunScraper(ctx, scraperName)

	duration := time.Since(start)

	// Record metrics
	if err != nil {
		metrics.ScraperRunsTotal.WithLabelValues(scraperName, "error").Inc()
		return nil, fmt.Errorf("scraper failed: %w", err)
	}

	// Get count of items after scraping (approximate)
	filterAfter := repository.ListFilter{
		Limit:  100,
		Offset: 0,
	}
	itemsAfter, _ := dependencies.Repository.List(ctx, filterAfter)
	countAfter := len(itemsAfter)

	// Calculate items scraped (this is approximate, in production you'd want to track this more precisely)
	itemsScraped := countAfter - countBefore
	if itemsScraped < 0 {
		itemsScraped = 0 // Safety check
	}

	result := &ScraperActivityResult{
		ItemsScraped: itemsScraped,
		ItemsSaved:   itemsScraped, // Simplified - assume all scraped items were saved
		Duration:     duration,
	}

	activity.RecordHeartbeat(ctx, "completed")

	logger.Info("Scraper activity completed",
		"scraper", scraperName,
		"duration", duration,
		"itemsScraped", itemsScraped)

	metrics.ScraperRunsTotal.WithLabelValues(scraperName, "success").Inc()

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
