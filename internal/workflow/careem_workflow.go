package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// CareemScraperWorkflowInput contains input for the Careem scraper workflow
type CareemScraperWorkflowInput struct {
	MaxRetries        int
	RefreshInterval   time.Duration // How often to refresh rates
	AlertOnChangePerc float64       // Alert if rate changes by this percentage
}

// CareemScraperWorkflowResult contains the result of the Careem scraper workflow
type CareemScraperWorkflowResult struct {
	ItemsScraped  int
	ItemsSaved    int
	Source        string
	Confidence    float32
	RateChanges   []string
	Errors        []string
	Duration      time.Duration
	CompletedAt   time.Time
	NextExecution time.Time
}

// CareemScraperWorkflow executes the Careem scraper with special handling for rate sources
func CareemScraperWorkflow(ctx workflow.Context, input CareemScraperWorkflowInput) (*CareemScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Careem scraper workflow")

	startTime := workflow.Now(ctx)

	// Configure activity options with longer timeout for Careem
	// Careem may need to try multiple sources
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute, // Longer timeout for multiple source attempts
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second, // Longer initial interval
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    int32(input.MaxRetries),
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, "careem").Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &CareemScraperWorkflowResult{
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("Careem scraper workflow failed", "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, "careem").Get(compensationCtx, &compensated)

		return result, err
	}

	logger.Info("Careem scraper workflow completed",
		"scraped", activityResult.ItemsScraped,
		"saved", activityResult.ItemsSaved,
		"duration", duration)

	return result, nil
}

// ScheduledCareemWorkflow runs the Careem scraper on a schedule
// Careem rates change infrequently, so we run monthly instead of daily
func ScheduledCareemWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled Careem workflow")

	// Careem rates are relatively stable, check monthly
	refreshInterval := 30 * 24 * time.Hour // 30 days

	for {
		input := CareemScraperWorkflowInput{
			MaxRetries:        5, // More retries since we try multiple sources
			RefreshInterval:   refreshInterval,
			AlertOnChangePerc: 10.0, // Alert on 10% rate change
		}

		// Execute Careem scraper
		var result CareemScraperWorkflowResult
		err := workflow.ExecuteChildWorkflow(ctx, CareemScraperWorkflow, input).Get(ctx, &result)

		if err != nil {
			logger.Error("Careem scraper failed", "error", err)
		} else {
			logger.Info("Careem scraper completed",
				"scraped", result.ItemsScraped,
				"saved", result.ItemsSaved,
				"source", result.Source,
				"confidence", result.Confidence)

			// Log rate changes if any
			if len(result.RateChanges) > 0 {
				logger.Warn("Significant rate changes detected", "changes", result.RateChanges)
			}
		}

		// Wait for next execution
		err = workflow.Sleep(ctx, refreshInterval)
		if err != nil {
			return err
		}
	}
}

// ValidateCareemRatesWorkflow validates freshness and quality of Careem rate data
func ValidateCareemRatesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Careem rates validation workflow")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Activity would check:
	// 1. Data freshness (< 60 days old)
	// 2. Rate reasonableness (within expected ranges)
	// 3. Completeness (all required rate components present)
	// 4. Consistency (minimum fare >= base fare, etc.)

	var validationResult struct {
		IsValid      bool
		Issues       []string
		DataAge      time.Duration
		LastSource   string
		Confidence   float32
		MissingRates []string
	}

	err := workflow.ExecuteActivity(ctx, "ValidateCareemRatesActivity").Get(ctx, &validationResult)
	if err != nil {
		logger.Error("Validation failed", "error", err)
		return err
	}

	if !validationResult.IsValid {
		logger.Warn("Careem rates validation issues found",
			"issues", validationResult.Issues,
			"data_age", validationResult.DataAge,
			"confidence", validationResult.Confidence)

		// If data is too old or confidence too low, trigger refresh
		if validationResult.DataAge > 90*24*time.Hour || validationResult.Confidence < 0.6 {
			logger.Info("Triggering rate refresh due to validation issues")

			input := CareemScraperWorkflowInput{
				MaxRetries:        5,
				RefreshInterval:   30 * 24 * time.Hour,
				AlertOnChangePerc: 10.0,
			}

			var refreshResult CareemScraperWorkflowResult
			err = workflow.ExecuteChildWorkflow(ctx, CareemScraperWorkflow, input).Get(ctx, &refreshResult)
			if err != nil {
				logger.Error("Refresh failed", "error", err)
				return err
			}
		}
	}

	logger.Info("Careem rates validation completed", "valid", validationResult.IsValid)
	return nil
}
