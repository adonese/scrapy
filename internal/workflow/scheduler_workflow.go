package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// ScraperSchedule defines the schedule for a scraper
type ScraperSchedule struct {
	Name      string
	Frequency time.Duration
	Priority  int // Lower number = higher priority
	Enabled   bool
}

// SchedulerConfig contains configuration for the master scheduler
type SchedulerConfig struct {
	Schedules []ScraperSchedule
}

// DefaultSchedulerConfig returns the recommended scheduler configuration
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		Schedules: []ScraperSchedule{
			// Housing scrapers - Daily (high priority)
			{Name: "bayut-Dubai", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "bayut-Sharjah", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "bayut-Ajman", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "bayut-Abu Dhabi", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "dubizzle-Dubai-apartmentflat", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "dubizzle-Sharjah-apartmentflat", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "dubizzle-Ajman-apartmentflat", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "dubizzle-Abu Dhabi-apartmentflat", Frequency: 24 * time.Hour, Priority: 1, Enabled: true},
			{Name: "dubizzle-Dubai-bedspace", Frequency: 24 * time.Hour, Priority: 2, Enabled: true},
			{Name: "dubizzle-Dubai-roomspace", Frequency: 24 * time.Hour, Priority: 2, Enabled: true},

			// Utility scrapers - Weekly (medium priority)
			{Name: "dewa", Frequency: 7 * 24 * time.Hour, Priority: 3, Enabled: true},
			{Name: "sewa", Frequency: 7 * 24 * time.Hour, Priority: 3, Enabled: true},
			{Name: "aadc", Frequency: 7 * 24 * time.Hour, Priority: 3, Enabled: true},

			// Transport scrapers - Weekly (medium priority)
			{Name: "rta", Frequency: 7 * 24 * time.Hour, Priority: 4, Enabled: true},

			// Ride-sharing scrapers - Monthly (low priority)
			{Name: "careem", Frequency: 30 * 24 * time.Hour, Priority: 5, Enabled: true},
		},
	}
}

// MasterSchedulerWorkflow is a long-running workflow that manages scraper schedules
// This workflow runs continuously and triggers scrapers based on their schedules
func MasterSchedulerWorkflow(ctx workflow.Context, config *SchedulerConfig) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting master scheduler workflow")

	if config == nil {
		config = DefaultSchedulerConfig()
	}

	// Track last execution time for each scraper
	lastExecution := make(map[string]time.Time)
	for _, schedule := range config.Schedules {
		if schedule.Enabled {
			lastExecution[schedule.Name] = time.Time{} // Never executed
		}
	}

	// Main scheduling loop
	for {
		currentTime := workflow.Now(ctx)

		// Check each scraper schedule
		for _, schedule := range config.Schedules {
			if !schedule.Enabled {
				continue
			}

			lastRun, exists := lastExecution[schedule.Name]
			if !exists || currentTime.Sub(lastRun) >= schedule.Frequency {
				// Time to run this scraper
				logger.Info("Triggering scraper",
					"name", schedule.Name,
					"frequency", schedule.Frequency,
					"last_run", lastRun)

				// Execute scraper as child workflow
				err := executeScheduledScraper(ctx, schedule.Name)
				if err != nil {
					logger.Error("Scheduled scraper failed",
						"name", schedule.Name,
						"error", err)
					// Continue with other scrapers
				}

				// Update last execution time
				lastExecution[schedule.Name] = currentTime
			}
		}

		// Sleep for a reasonable interval (check every hour)
		err := workflow.Sleep(ctx, 1*time.Hour)
		if err != nil {
			return err
		}
	}
}

// executeScheduledScraper triggers a scraper execution as a child workflow
func executeScheduledScraper(ctx workflow.Context, scraperName string) error {
	logger := workflow.GetLogger(ctx)

	// Configure child workflow options
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:            "scraper-" + scraperName + "-" + workflow.Now(ctx).Format("20060102-150405"),
		WorkflowExecutionTimeout: 30 * time.Minute,
		WorkflowTaskTimeout:      5 * time.Minute,
	}
	childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

	// Execute scraper workflow based on type
	// Use batch workflow with single scraper
	input := BatchScraperWorkflowInput{
		ScraperNames: []string{scraperName},
		MaxRetries:   3,
		Timeout:      10 * time.Minute,
		Sequential:   true,
		ValidateData: true,
	}

	var result BatchScraperWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, BatchScraperWorkflow, input).Get(ctx, &result)
	if err != nil {
		return err
	}

	logger.Info("Scheduled scraper completed",
		"name", scraperName,
		"items", result.TotalSaved,
		"duration", result.Duration)

	return nil
}

// CronSchedulerWorkflow uses cron-like scheduling for scrapers
// This is an alternative to MasterSchedulerWorkflow using fixed times
func CronSchedulerWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting cron scheduler workflow")

	for {
		currentTime := workflow.Now(ctx)
		hour := currentTime.Hour()

		// Daily scrapers at 2 AM
		if hour == 2 {
			logger.Info("Triggering daily scrapers")
			err := workflow.ExecuteChildWorkflow(
				workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
					WorkflowID: "daily-scrapers-" + currentTime.Format("20060102"),
				}),
				DailyScraperWorkflow,
			).Get(ctx, nil)
			if err != nil {
				logger.Error("Daily scrapers failed", "error", err)
			}
		}

		// Weekly scrapers on Monday at 3 AM
		if currentTime.Weekday() == time.Monday && hour == 3 {
			logger.Info("Triggering weekly scrapers")
			err := workflow.ExecuteChildWorkflow(
				workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
					WorkflowID: "weekly-scrapers-" + currentTime.Format("20060102"),
				}),
				WeeklyScraperWorkflow,
			).Get(ctx, nil)
			if err != nil {
				logger.Error("Weekly scrapers failed", "error", err)
			}
		}

		// Monthly scrapers on 1st of month at 4 AM
		if currentTime.Day() == 1 && hour == 4 {
			logger.Info("Triggering monthly scrapers")
			err := workflow.ExecuteChildWorkflow(
				workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
					WorkflowID: "monthly-scrapers-" + currentTime.Format("200601"),
				}),
				MonthlyScraperWorkflow,
			).Get(ctx, nil)
			if err != nil {
				logger.Error("Monthly scrapers failed", "error", err)
			}
		}

		// Sleep until next hour
		nextHour := currentTime.Add(time.Hour).Truncate(time.Hour)
		sleepDuration := nextHour.Sub(currentTime)
		err := workflow.Sleep(ctx, sleepDuration)
		if err != nil {
			return err
		}
	}
}

// OnDemandScraperWorkflow allows manual triggering of any scraper or group
type OnDemandScraperInput struct {
	ScraperNames []string // Scrapers to run
	Category     string   // Or run by category
	Priority     bool     // If true, run with higher priority
}

func OnDemandScraperWorkflow(ctx workflow.Context, input OnDemandScraperInput) (*BatchScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting on-demand scraper workflow",
		"scrapers", input.ScraperNames,
		"category", input.Category,
		"priority", input.Priority)

	// Build batch input
	batchInput := BatchScraperWorkflowInput{
		ScraperNames: input.ScraperNames,
		Category:     input.Category,
		MaxRetries:   3,
		Timeout:      10 * time.Minute,
		Sequential:   !input.Priority, // Run in parallel if priority
		ValidateData: true,
	}

	// Execute batch workflow
	result, err := BatchScraperWorkflow(ctx, batchInput)
	if err != nil {
		logger.Error("On-demand scraper workflow failed", "error", err)
		return nil, err
	}

	logger.Info("On-demand scraper workflow completed",
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"items", result.TotalSaved)

	return result, nil
}
