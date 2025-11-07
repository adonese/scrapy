package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// AADCScraperWorkflowInput contains configuration for AADC scraper workflow
type AADCScraperWorkflowInput struct {
	MaxRetries int
	Timeout    time.Duration
}

// AADCScraperWorkflow executes the AADC utility rates scraper
// This workflow is designed to run on a weekly schedule to capture
// any changes in utility rates for Abu Dhabi
func AADCScraperWorkflow(ctx workflow.Context, input AADCScraperWorkflowInput) (*ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting AADC scraper workflow")

	startTime := workflow.Now(ctx)

	// Set default timeout if not provided
	timeout := input.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Set default retries if not provided
	maxRetries := input.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	// Configure activity options with retries
	// AADC is an official government utility provider, so we can be aggressive with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,  // Start with 2s delay
			BackoffCoefficient: 2.0,              // Double the interval each time
			MaximumInterval:    2 * time.Minute,  // Cap at 2 minutes
			MaximumAttempts:    int32(maxRetries),
			NonRetryableErrorTypes: []string{
				"InvalidHTMLError", // Don't retry if HTML structure changed
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute AADC scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, "aadc").Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &ScraperWorkflowResult{
		ScraperName:  "aadc",
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("AADC scraper workflow failed", "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, "aadc").Get(compensationCtx, &compensated)

		// TODO: Send alert for failed AADC scrape (utility rates are critical)
		// SendAlertActivity needs to be implemented by Agent 10
		// alertAO := workflow.ActivityOptions{
		// 	StartToCloseTimeout: 30 * time.Second,
		// }
		// alertCtx := workflow.WithActivityOptions(ctx, alertAO)
		// workflow.ExecuteActivity(alertCtx, SendAlertActivity, map[string]interface{}{
		// 	"scraper": "aadc",
		// 	"error":   err.Error(),
		// 	"type":    "utility_scraper_failed",
		// }).Get(alertCtx, nil)

		return result, err
	}

	// Validate expected data volume
	// AADC should return approximately 12 data points:
	// - 2 electricity tiers for nationals (up to 30k, above 30k)
	// - 8 electricity tiers for expatriates (up to 400, 401-700, 701-1000, etc.)
	// - 2 water rates (national, expatriate)
	if activityResult.ItemsScraped < 10 {
		logger.Warn("AADC scraper returned fewer items than expected",
			"expected_min", 10,
			"actual", activityResult.ItemsScraped)

		// TODO: Send warning alert
		// SendAlertActivity needs to be implemented by Agent 10
		// alertAO := workflow.ActivityOptions{
		// 	StartToCloseTimeout: 30 * time.Second,
		// }
		// alertCtx := workflow.WithActivityOptions(ctx, alertAO)
		// workflow.ExecuteActivity(alertCtx, SendAlertActivity, map[string]interface{}{
		// 	"scraper":      "aadc",
		// 	"type":         "low_data_volume",
		// 	"expected_min": 10,
		// 	"actual":       activityResult.ItemsScraped,
		// }).Get(alertCtx, nil)
	}

	logger.Info("AADC scraper workflow completed",
		"scraped", activityResult.ItemsScraped,
		"saved", activityResult.ItemsSaved,
		"duration", duration)

	return result, nil
}

// ScheduledAADCWorkflow runs the AADC scraper on a weekly schedule
// Utility rates don't change frequently, so weekly checks are sufficient
func ScheduledAADCWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled AADC workflow")

	input := AADCScraperWorkflowInput{
		MaxRetries: 3,
		Timeout:    5 * time.Minute,
	}

	result, err := AADCScraperWorkflow(ctx, input)

	if err != nil {
		logger.Error("Scheduled AADC workflow failed", "error", err)
		return err
	}

	logger.Info("Scheduled AADC workflow completed",
		"scraped", result.ItemsScraped,
		"saved", result.ItemsSaved)

	return nil
}
