package workflow

import (
	"fmt"
	"strings"
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
	TotalScrapers     int
	SuccessCount      int
	FailedCount       int
	TotalItems        int
	TotalSaved        int
	TotalValidated    int
	TotalSaveFailures int
	Duration          time.Duration
	ScraperResults    []ScraperWorkflowResult
	ValidationStats   *ValidationStats
	CompletedAt       time.Time
}

// ValidationStats contains validation statistics
type ValidationStats struct {
	TotalValidated int
	ValidCount     int
	InvalidCount   int
	QualityScore   float64
}

// resolveScraperNames canonicalises the requested scraper list and falls back to scheduler defaults
// when the caller does not provide explicit names. This keeps workflow inputs resilient to
// naming changes and ensures category-based runs execute the expected scrapers.
func resolveScraperNames(input *BatchScraperWorkflowInput) []string {
	if input == nil {
		return nil
	}

	if len(input.ScraperNames) > 0 {
		seen := make(map[string]struct{}, len(input.ScraperNames))
		resolved := make([]string, 0, len(input.ScraperNames))
		for _, name := range input.ScraperNames {
			canonical := strings.TrimSpace(strings.ToLower(name))
			if canonical == "" {
				continue
			}
			if _, exists := seen[canonical]; exists {
				continue
			}
			resolved = append(resolved, canonical)
			seen[canonical] = struct{}{}
		}
		return resolved
	}

	// Fallback to scheduler defaults filtered by category when provided.
	resolved := getEnabledScraperNames(input.Category, 0, 0)
	if len(resolved) > 0 {
		return resolved
	}

	// No category-specific scrapers found; return every enabled scraper so the workflow still executes.
	return getEnabledScraperNames("", 0, 0)
}

// BatchScraperWorkflow executes multiple scrapers in batch
// Can run scrapers in parallel or sequentially based on configuration
func BatchScraperWorkflow(ctx workflow.Context, input BatchScraperWorkflowInput) (*BatchScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)

	startTime := workflow.Now(ctx)

	// Resolve requested scrapers using scheduler defaults when none provided.
	resolvedNames := resolveScraperNames(&input)
	logger.Info("Starting batch scraper workflow",
		"requested", len(input.ScraperNames),
		"resolved", len(resolvedNames),
		"category", input.Category)
	if len(resolvedNames) == 0 {
		logger.Warn("No scrapers resolved for batch execution", "category", input.Category)
	}

	input.ScraperNames = resolvedNames

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
		result.TotalValidated += sr.Validation.Valid
		result.TotalSaveFailures += sr.SaveFailures

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
		"validated", result.TotalValidated,
		"saved", result.TotalSaved,
		"save_failures", result.TotalSaveFailures,
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
			SaveFailures: activityResult.SaveFailures,
			Validation:   activityResult.Validation,
			Duration:     activityResult.Duration,
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
			SaveFailures: activityResult.SaveFailures,
			Validation:   activityResult.Validation,
			Duration:     activityResult.Duration,
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

	housingScrapers := dailyScraperNames()

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
		"items_scraped", result.TotalItems,
		"validated", result.TotalValidated,
		"saved", result.TotalSaved,
		"save_failures", result.TotalSaveFailures)

	return nil
}

// WeeklyScraperWorkflow runs weekly scrapers (utilities and transport)
func WeeklyScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting weekly scraper workflow")

	weeklyScrapers := weeklyScraperNames()

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
		"items_scraped", result.TotalItems,
		"validated", result.TotalValidated,
		"saved", result.TotalSaved,
		"save_failures", result.TotalSaveFailures)

	return nil
}

// MonthlyScraperWorkflow runs monthly scrapers (Careem rates)
func MonthlyScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting monthly scraper workflow")

	monthlyScrapers := monthlyScraperNames()

	input := BatchScraperWorkflowInput{
		ScraperNames: monthlyScrapers,
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
		"items_scraped", result.TotalItems,
		"validated", result.TotalValidated,
		"saved", result.TotalSaved,
		"save_failures", result.TotalSaveFailures)

	return nil
}

func dailyScraperNames() []string {
	return getEnabledScraperNames("housing", 24*time.Hour, 0)
}

func weeklyScraperNames() []string {
	utilities := getEnabledScraperNames("utilities", 7*24*time.Hour, 24*time.Hour)
	transportation := getEnabledScraperNames("transportation", 7*24*time.Hour, 24*time.Hour)
	return appendUnique(utilities, transportation...)
}

func monthlyScraperNames() []string {
	rideshare := getEnabledScraperNames("rideshare", 0, 0)
	if len(rideshare) > 0 {
		return rideshare
	}

	// Fallback to any scrapers scheduled less frequently than weekly.
	return getEnabledScraperNames("", 0, 8*24*time.Hour)
}

func appendUnique(base []string, additional ...string) []string {
	if len(additional) == 0 {
		return base
	}

	seen := make(map[string]struct{}, len(base)+len(additional))
	for _, name := range base {
		seen[name] = struct{}{}
	}

	for _, name := range additional {
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		base = append(base, name)
		seen[name] = struct{}{}
	}

	return base
}
