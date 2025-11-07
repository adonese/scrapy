package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/adonese/cost-of-living/internal/services"
)

type ScraperWorkflowInput struct {
	ScraperName string
	MaxRetries  int
}

type ScraperWorkflowResult struct {
	ScraperName  string
	ItemsScraped int
	ItemsSaved   int
	SaveFailures int
	Validation   services.ValidationSummary
	Errors       []string
	Duration     time.Duration
	CompletedAt  time.Time
}

// ScraperWorkflow executes a single scraper with retry logic and compensation
func ScraperWorkflow(ctx workflow.Context, input ScraperWorkflowInput) (*ScraperWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scraper workflow", "scraper", input.ScraperName)

	startTime := workflow.Now(ctx)

	// Configure activity options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    int32(input.MaxRetries),
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute scraper activity
	var activityResult ScraperActivityResult
	err := workflow.ExecuteActivity(ctx, RunScraperActivity, input.ScraperName).Get(ctx, &activityResult)

	duration := workflow.Now(ctx).Sub(startTime)

	result := &ScraperWorkflowResult{
		ScraperName:  input.ScraperName,
		ItemsScraped: activityResult.ItemsScraped,
		ItemsSaved:   activityResult.ItemsSaved,
		SaveFailures: activityResult.SaveFailures,
		Validation:   activityResult.Validation,
		Duration:     duration,
		CompletedAt:  workflow.Now(ctx),
	}

	if err != nil {
		logger.Error("Scraper workflow failed", "scraper", input.ScraperName, "error", err)
		result.Errors = []string{err.Error()}

		// Execute compensation activity
		compensationAO := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
		}
		compensationCtx := workflow.WithActivityOptions(ctx, compensationAO)

		var compensated bool
		workflow.ExecuteActivity(compensationCtx, CompensateFailedScrapeActivity, input.ScraperName).Get(compensationCtx, &compensated)

		return result, err
	}

	logger.Info("Scraper workflow completed",
		"scraper", input.ScraperName,
		"scraped", activityResult.ItemsScraped,
		"validated", activityResult.ItemsValidated,
		"saved", activityResult.ItemsSaved,
		"save_failures", activityResult.SaveFailures,
		"duration", duration)

	return result, nil
}

// ScheduledScraperWorkflow runs all scrapers periodically
func ScheduledScraperWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled scraper workflow")

	scrapers := dailyScraperNames()

	// Run scrapers in parallel
	futures := []workflow.Future{}

	for _, scraperName := range scrapers {
		input := ScraperWorkflowInput{
			ScraperName: scraperName,
			MaxRetries:  3,
		}

		// Execute as child workflow
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("scraper-%s-%d", scraperName, workflow.Now(ctx).Unix()),
		})

		future := workflow.ExecuteChildWorkflow(childCtx, ScraperWorkflow, input)
		futures = append(futures, future)
	}

	// Wait for all scrapers to complete
	for i, future := range futures {
		var result ScraperWorkflowResult
		if err := future.Get(ctx, &result); err != nil {
			logger.Error("Child workflow failed", "scraper", scrapers[i], "error", err)
			// Continue with other scrapers
		} else {
			logger.Info("Child workflow completed",
				"scraper", scrapers[i],
				"scraped", result.ItemsScraped,
				"saved", result.ItemsSaved)
		}
	}

	logger.Info("Scheduled scraper workflow completed")
	return nil
}
