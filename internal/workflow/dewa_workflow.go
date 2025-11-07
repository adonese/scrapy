package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DEWAScraperWorkflowInput contains configuration for DEWA scraper workflow
type DEWAScraperWorkflowInput struct {
	MaxRetries int
	Timeout    time.Duration
}

// DEWAScraperWorkflow executes the DEWA utility rates scraper
// This workflow is designed to run on a weekly schedule to capture
// any changes in utility rates for Dubai
func DEWAScraperWorkflow(ctx workflow.Context, input DEWAScraperWorkflowInput) (*ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DEWA scraper workflow")

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
	// DEWA is an official government site, so we can be more aggressive with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    int32(maxRetries),
			NonRetryableErrorTypes: []string{
				"InvalidHTMLError", // Don't retry if HTML structure changed
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute DEWA scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, "dewa").Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &ScraperWorkflowResult{
		ScraperName:  "dewa",
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("DEWA scraper workflow failed", "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, "dewa").Get(compensationCtx, &compensated)

		return result, err
	}

	// Validate expected data volume
	// DEWA should return approximately 7-8 data points (4 electricity + 3 water + fuel surcharge)
	if activityResult.ItemsScraped < 6 {
		logger.Warn("DEWA scraper returned fewer items than expected",
			"expected_min", 6,
			"actual", activityResult.ItemsScraped)
	}

	logger.Info("DEWA scraper workflow completed",
		"scraped", activityResult.ItemsScraped,
		"saved", activityResult.ItemsSaved,
		"duration", duration)

	return result, nil
}

// ScheduledDEWAWorkflow runs the DEWA scraper on a weekly schedule
// Utility rates don't change frequently, so weekly checks are sufficient
func ScheduledDEWAWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled DEWA workflow")

	input := DEWAScraperWorkflowInput{
		MaxRetries: 3,
		Timeout:    5 * time.Minute,
	}

	result, err := DEWAScraperWorkflow(ctx, input)

	if err != nil {
		logger.Error("Scheduled DEWA workflow failed", "error", err)
		return err
	}

	logger.Info("Scheduled DEWA workflow completed",
		"scraped", result.ItemsScraped,
		"saved", result.ItemsSaved)

	return nil
}
