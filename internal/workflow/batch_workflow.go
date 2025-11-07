package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// BatchScraperWorkflowInput contains configuration for batch scraper execution
type BatchScraperWorkflowInput struct {
	ScraperNames []string      // Specific scrapers to run, empty means all
	Category     string        // Filter by category (housing, utilities, transportation)
	MaxRetries   int           // Maximum retry attempts per scraper
	Timeout      time.Duration // Timeout per scraper
	Sequential   bool          // Run scrapers sequentially (default: parallel)
	ValidateData bool          // Enable validation after scraping
}

// BatchScraperWorkflowResult contains the results of batch execution
type BatchScraperWorkflowResult struct {
	TotalScrapers   int
	SuccessCount    int
	FailedCount     int
	TotalItems      int
	TotalSaved      int
	Duration        time.Duration
	ScraperResults  []ScraperWorkflowResult
	ValidationStats *ValidationStats
	CompletedAt     time.Time
}

// ValidationStats contains validation statistics
type ValidationStats struct {
	TotalValidated int
	ValidCount     int
	InvalidCount   int
	QualityScore   float64
}

// BatchScraperWorkflow executes multiple scrapers in batch
// Can run scrapers in parallel or sequentially based on configuration
func BatchScraperWorkflow(ctx workflow.Context, input BatchScraperWorkflowInput) (*BatchScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting batch scraper workflow", "scrapers", len(input.ScraperNames))

	startTime := workflow.Now(ctx)

	// Set defaults
	if input.MaxRetries == 0 {
		input.MaxRetries = 3
	}
	if input.Timeout == 0 {
		input.Timeout = 5 * time.Minute
	}

	result := &BatchScraperWorkflowResult{
		TotalScrapers:  len(input.ScraperNames),
		ScraperResults: make([]ScraperWorkflowResult, 0, len(input.ScraperNames)),
		CompletedAt:    workflow.Now(ctx),
	}

	// Execute scrapers based on mode
	var scraperResults []ScraperWorkflowResult
	var err error

	if input.Sequential {
		scraperResults, err = runScrapersSequentially(ctx, input)
	} else {
		scraperResults, err = runScrapersParallel(ctx, input)
	}

	if err != nil {
		logger.Error("Batch scraper workflow failed", "error", err)
		return result, err
	}

	// Aggregate results
	for _, sr := range scraperResults {
		result.ScraperResults = append(result.ScraperResults, sr)
		result.TotalItems += sr.ItemsScraped
		result.TotalSaved += sr.ItemsSaved

		if len(sr.Errors) > 0 {
			result.FailedCount++
		} else {
			result.SuccessCount++
		}
	}

	// Run validation if enabled
	if input.ValidateData {
		validationStats, err := runBatchValidation(ctx)
		if err != nil {
			logger.Warn("Batch validation failed", "error", err)
		} else {
			result.ValidationStats = validationStats
		}
	}

	result.Duration = workflow.Now(ctx).Sub(startTime)

	logger.Info("Batch scraper workflow completed",
		"total", result.TotalScrapers,
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"items", result.TotalItems,
		"saved", result.TotalSaved,
		"duration", result.Duration)

	return result, nil
}

// runScrapersSequentially executes scrapers one at a time
func runScrapersSequentially(ctx workflow.Context, input BatchScraperWorkflowInput) ([]ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	results := make([]ScraperWorkflowResult, 0, len(input.ScraperNames))

	for _, scraperName := range input.ScraperNames {
		logger.Info("Running scraper sequentially", "name", scraperName)

		// Configure activity options
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: input.Timeout,
			HeartbeatTimeout:    30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    2 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    2 * time.Minute,
				MaximumAttempts:    int32(input.MaxRetries),
			},
		}
		activityCtx := workflow.WithActivityOptions(ctx, ao)

		// Execute scraper activity
		var activityResult ScraperActivityResult
		err := workflow.ExecuteActivity(activityCtx, RunScraperActivity, scraperName).Get(ctx, &activityResult)

		workflowResult := ScraperWorkflowResult{
			ScraperName:  scraperName,
			ItemsScraped: activityResult.ItemsScraped,
			ItemsSaved:   activityResult.ItemsSaved,
			CompletedAt:  workflow.Now(ctx),
		}

		if err != nil {
			logger.Error("Scraper failed", "name", scraperName, "error", err)
			workflowResult.Errors = []string{err.Error()}
			// Continue with next scraper even if this one fails
		}

		results = append(results, workflowResult)
	}

	return results, nil
}

// runScrapersParallel executes scrapers concurrently
func runScrapersParallel(ctx workflow.Context, input BatchScraperWorkflowInput) ([]ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Running scrapers in parallel", "count", len(input.ScraperNames))

	// Configure activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: input.Timeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    int32(input.MaxRetries),
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Create futures for parallel execution
	futures := make([]workflow.Future, 0, len(input.ScraperNames))
	scraperNames := make([]string, 0, len(input.ScraperNames))

	for _, scraperName := range input.ScraperNames {
		future := workflow.ExecuteActivity(activityCtx, RunScraperActivity, scraperName)
		futures = append(futures, future)
		scraperNames = append(scraperNames, scraperName)
	}

	// Wait for all scrapers to complete
	results := make([]ScraperWorkflowResult, 0, len(futures))
	for i, future := range futures {
		var activityResult ScraperActivityResult
		err := future.Get(ctx, &activityResult)

		workflowResult := ScraperWorkflowResult{
			ScraperName:  scraperNames[i],
			ItemsScraped: activityResult.ItemsScraped,
			ItemsSaved:   activityResult.ItemsSaved,
			CompletedAt:  workflow.Now(ctx),
		}

		if err != nil {
			logger.Error("Scraper failed", "name", scraperNames[i], "error", err)
			workflowResult.Errors = []string{err.Error()}
		}

		results = append(results, workflowResult)
	}

	return results, nil
}

// runBatchValidation runs validation on recently scraped data
func runBatchValidation(ctx workflow.Context) (*ValidationStats, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Running batch validation")

	// Configure validation activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	validationCtx := workflow.WithActivityOptions(ctx, ao)

	// Execute validation activity
	var stats ValidationStats
	err := workflow.ExecuteActivity(validationCtx, ValidateRecentDataActivity).Get(ctx, &stats)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	logger.Info("Batch validation completed",
		"total", stats.TotalValidated,
		"valid", stats.ValidCount,
		"invalid", stats.InvalidCount,
		"quality_score", stats.QualityScore)

	return &stats, nil
}

// DailyScraperWorkflow runs daily scrapers (housing data)
func DailyScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting daily scraper workflow")

	// Housing scrapers - run daily
	housingScrapers := []string{
		"bayut-Dubai",
		"bayut-Sharjah",
		"bayut-Ajman",
		"bayut-Abu Dhabi",
		"dubizzle-Dubai-apartmentflat",
		"dubizzle-Sharjah-apartmentflat",
		"dubizzle-Ajman-apartmentflat",
		"dubizzle-Abu Dhabi-apartmentflat",
		"dubizzle-Dubai-bedspace",
		"dubizzle-Dubai-roomspace",
	}

	input := BatchScraperWorkflowInput{
		ScraperNames: housingScrapers,
		Category:     "housing",
		MaxRetries:   3,
		Timeout:      5 * time.Minute,
		Sequential:   false, // Run in parallel for speed
		ValidateData: true,
	}

	result, err := BatchScraperWorkflow(ctx, input)
	if err != nil {
		logger.Error("Daily scraper workflow failed", "error", err)
		return err
	}

	logger.Info("Daily scraper workflow completed",
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"items", result.TotalSaved)

	return nil
}

// WeeklyScraperWorkflow runs weekly scrapers (utilities and transport)
func WeeklyScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting weekly scraper workflow")

	// Utility and transport scrapers - run weekly
	weeklyScrapers := []string{
		"dewa",
		"sewa",
		"aadc",
		"rta",
	}

	input := BatchScraperWorkflowInput{
		ScraperNames: weeklyScrapers,
		MaxRetries:   3,
		Timeout:      5 * time.Minute,
		Sequential:   true, // Run sequentially to avoid overwhelming official sites
		ValidateData: true,
	}

	result, err := BatchScraperWorkflow(ctx, input)
	if err != nil {
		logger.Error("Weekly scraper workflow failed", "error", err)
		return err
	}

	logger.Info("Weekly scraper workflow completed",
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"items", result.TotalSaved)

	return nil
}

// MonthlyScraperWorkflow runs monthly scrapers (Careem rates)
func MonthlyScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting monthly scraper workflow")

	input := BatchScraperWorkflowInput{
		ScraperNames: []string{"careem"},
		MaxRetries:   3,
		Timeout:      5 * time.Minute,
		Sequential:   true,
		ValidateData: true,
	}

	result, err := BatchScraperWorkflow(ctx, input)
	if err != nil {
		logger.Error("Monthly scraper workflow failed", "error", err)
		return err
	}

	logger.Info("Monthly scraper workflow completed",
		"success", result.SuccessCount,
		"items", result.TotalSaved)

	return nil
}
