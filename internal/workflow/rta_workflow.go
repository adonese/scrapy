package workflow

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RTAWorkflowInput contains input parameters for RTA scraper workflow
type RTAWorkflowInput struct {
	MaxRetries int
}

// RTAWorkflowResult contains the results of RTA scraper workflow execution
type RTAWorkflowResult struct {
	ItemsScraped int
	ItemsSaved   int
	Errors       []string
	Duration     time.Duration
	CompletedAt  time.Time
}

// RTAScraperWorkflow executes the RTA scraper with retry logic
// RTA fare data changes infrequently, so this workflow is designed to run weekly
func RTAScraperWorkflow(ctx workflow.Context, input RTAWorkflowInput) (*RTAWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting RTA scraper workflow")

	startTime := workflow.Now(ctx)

	// Configure activity options with retries
	// RTA is official government site, so we can be more aggressive with timeouts
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute, // RTA site should be fast
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 2,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 2,
			MaximumAttempts:    int32(input.MaxRetries),
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute RTA scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, "rta").Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &RTAWorkflowResult{
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("RTA scraper workflow failed", "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, "rta").Get(compensationCtx, &compensated)

		return result, err
	}

	// Validate we got reasonable data
	if activityResult.ItemsScraped < 20 {
		logger.Warn("RTA scraper extracted fewer items than expected",
			"expected", 20,
			"actual", activityResult.ItemsScraped)
		result.Errors = append(result.Errors, "Low item count - possible parsing issue")
	}

	logger.Info("RTA scraper workflow completed",
		"scraped", activityResult.ItemsScraped,
		"saved", activityResult.ItemsSaved,
		"duration", duration)

	return result, nil
}

// ScheduledRTAWorkflow runs the RTA scraper on a weekly schedule
// RTA fares change infrequently (typically quarterly or annually)
func ScheduledRTAWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled RTA scraper workflow")

	// Execute RTA workflow
	input := RTAWorkflowInput{
		MaxRetries: 3,
	}

	var result RTAWorkflowResult
	err := workflow.ExecuteActivity(ctx, RunRTAScraperActivity, input).Get(ctx, &result)

	if err != nil {
		logger.Error("Scheduled RTA workflow failed", "error", err)
		return err
	}

	logger.Info("Scheduled RTA workflow completed",
		"scraped", result.ItemsScraped,
		"saved", result.ItemsSaved,
		"duration", result.Duration)

	// Schedule next run (7 days)
	return workflow.NewTimer(ctx, 7*24*time.Hour).Get(ctx, nil)
}

// RunRTAScraperActivity is the activity function that performs RTA scraping
// This wraps the standard RunScraperActivity with RTA-specific logic
func RunRTAScraperActivity(ctx context.Context, input RTAWorkflowInput) (RTAWorkflowResult, error) {
	// Use the generic scraper activity
	activityResult, err := RunScraperActivity(ctx, "rta")

	result := RTAWorkflowResult{
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		CompletedAt:  time.Now(),
	}

	if err != nil {
		result.Errors = []string{err.Error()}
		return result, err
	}

	return result, nil
}
