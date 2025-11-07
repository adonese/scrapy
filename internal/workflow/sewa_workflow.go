package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SEWAScraperWorkflowInput contains configuration for SEWA scraper workflow
type SEWAScraperWorkflowInput struct {
	MaxRetries int
	Timeout    time.Duration
}

// SEWAScraperWorkflow executes the SEWA utility rates scraper
// This workflow is designed to run on a weekly schedule to capture
// any changes in utility rates for Sharjah
func SEWAScraperWorkflow(ctx workflow.Context, input SEWAScraperWorkflowInput) (*ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SEWA scraper workflow")

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
	// SEWA is an official government site, so we can be more aggressive with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,     // Start with 2s delay
			BackoffCoefficient: 2.0,                 // Double the interval each time
			MaximumInterval:    2 * time.Minute,     // Cap at 2 minutes
			MaximumAttempts:    int32(maxRetries),
			NonRetryableErrorTypes: []string{
				"InvalidHTMLError",  // Don't retry if HTML structure changed
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute SEWA scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, "sewa").Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &ScraperWorkflowResult{
		ScraperName:  "sewa",
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("SEWA scraper workflow failed", "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, "sewa").Get(compensationCtx, &compensated)

		// TODO: Send alert for failed SEWA scrape (utility rates are critical)
		// SendAlertActivity needs to be implemented by Agent 10
		// alertAO := workflow.ActivityOptions{
		// 	StartToCloseTimeout: 30 * time.Second,
		// }
		// alertCtx := workflow.WithActivityOptions(ctx, alertAO)
		// workflow.ExecuteActivity(alertCtx, SendAlertActivity, map[string]interface{}{
		// 	"scraper": "sewa",
		// 	"error":   err.Error(),
		// 	"type":    "utility_scraper_failed",
		// }).Get(alertCtx, nil)

		return result, err
	}

	// Validate expected data volume
	// SEWA should return approximately 10 data points (7 electricity + 2 water + 1 sewerage)
	if activityResult.ItemsScraped < 8 {
		logger.Warn("SEWA scraper returned fewer items than expected",
			"expected_min", 8,
			"actual", activityResult.ItemsScraped)

		// TODO: Send warning alert
		// SendAlertActivity needs to be implemented by Agent 10
		// alertAO := workflow.ActivityOptions{
		// 	StartToCloseTimeout: 30 * time.Second,
		// }
		// alertCtx := workflow.WithActivityOptions(ctx, alertAO)
		// workflow.ExecuteActivity(alertCtx, SendAlertActivity, map[string]interface{}{
		// 	"scraper":      "sewa",
		// 	"type":         "low_data_volume",
		// 	"expected_min": 8,
		// 	"actual":       activityResult.ItemsScraped,
		// }).Get(alertCtx, nil)
	}

	logger.Info("SEWA scraper workflow completed",
		"scraped", activityResult.ItemsScraped,
		"saved", activityResult.ItemsSaved,
		"duration", duration)

	return result, nil
}

// ScheduledSEWAWorkflow runs the SEWA scraper on a weekly schedule
// Utility rates don't change frequently, so weekly checks are sufficient
func ScheduledSEWAWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled SEWA workflow")

	input := SEWAScraperWorkflowInput{
		MaxRetries: 3,
		Timeout:    5 * time.Minute,
	}

	result, err := SEWAScraperWorkflow(ctx, input)

	if err != nil {
		logger.Error("Scheduled SEWA workflow failed", "error", err)
		return err
	}

	logger.Info("Scheduled SEWA workflow completed",
		"scraped", result.ItemsScraped,
		"saved", result.ItemsSaved)

	return nil
}
