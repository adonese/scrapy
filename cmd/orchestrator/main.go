package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/adonese/cost-of-living/internal/workflow"
	"github.com/adonese/cost-of-living/pkg/logger"
)

func main() {
	logger.Init()

	// Define command flags
	var (
		command      = flag.String("command", "", "Command to execute: trigger, validate, status, schedule")
		scrapers     = flag.String("scrapers", "", "Comma-separated list of scraper names (empty = all)")
		category     = flag.String("category", "", "Scraper category: housing, utilities, transportation")
		schedule     = flag.String("schedule", "", "Schedule type: daily, weekly, monthly, cron")
		sequential   = flag.Bool("sequential", false, "Run scrapers sequentially instead of parallel")
		validate     = flag.Bool("validate", true, "Enable validation after scraping")
		workflowID   = flag.String("workflow-id", "", "Custom workflow ID (auto-generated if empty)")
		temporalAddr = flag.String("temporal", os.Getenv("TEMPORAL_ADDRESS"), "Temporal server address")
	)

	flag.Parse()

	// Set default Temporal address
	if *temporalAddr == "" {
		*temporalAddr = "localhost:7233"
	}

	// Validate command
	if *command == "" {
		fmt.Println("Usage: orchestrator -command <trigger|validate|status|schedule> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: *temporalAddr,
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// Execute command
	switch *command {
	case "trigger":
		err = triggerScrapers(ctx, c, *scrapers, *category, *sequential, *validate, *workflowID)
	case "validate":
		err = validateData(ctx, c, *scrapers)
	case "status":
		err = checkStatus(ctx, c, *workflowID)
	case "schedule":
		err = startScheduler(ctx, c, *schedule)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}

	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

// triggerScrapers triggers scraper execution
func triggerScrapers(ctx context.Context, c client.Client, scraperNames, category string, sequential, validate bool, workflowID string) error {
	logger.Info("Triggering scrapers", "names", scraperNames, "category", category)

	// Parse scraper names
	var scraperList []string
	if scraperNames != "" {
		scraperList = strings.Split(scraperNames, ",")
		// Trim whitespace
		for i := range scraperList {
			scraperList[i] = strings.TrimSpace(scraperList[i])
		}
	}

	// Create workflow input
	input := workflow.BatchScraperWorkflowInput{
		ScraperNames: scraperList,
		Category:     category,
		MaxRetries:   3,
		Timeout:      10 * time.Minute,
		Sequential:   sequential,
		ValidateData: validate,
	}

	// Generate workflow ID if not provided
	if workflowID == "" {
		workflowID = fmt.Sprintf("batch-scraper-%s", time.Now().Format("20060102-150405"))
	}

	// Start workflow
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "cost-of-living-task-queue",
	}

	we, err := c.ExecuteWorkflow(ctx, options, workflow.BatchScraperWorkflow, input)
	if err != nil {
		return fmt.Errorf("failed to start workflow: %w", err)
	}

	logger.Info("Workflow started", "workflow_id", we.GetID(), "run_id", we.GetRunID())
	fmt.Printf("Workflow started successfully!\n")
	fmt.Printf("Workflow ID: %s\n", we.GetID())
	fmt.Printf("Run ID: %s\n", we.GetRunID())
	fmt.Printf("\nMonitor progress:\n")
	fmt.Printf("  temporal workflow show --workflow-id %s\n", we.GetID())

	// Wait for result
	fmt.Println("\nWaiting for workflow to complete...")
	var result workflow.BatchScraperWorkflowResult
	err = we.Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Print results
	fmt.Println("\n=== Scraper Execution Results ===")
	fmt.Printf("Total Scrapers: %d\n", result.TotalScrapers)
	fmt.Printf("Successful: %d\n", result.SuccessCount)
	fmt.Printf("Failed: %d\n", result.FailedCount)
	fmt.Printf("Total Items Scraped: %d\n", result.TotalItems)
	fmt.Printf("Total Items Saved: %d\n", result.TotalSaved)
	fmt.Printf("Duration: %s\n", result.Duration)

	if result.ValidationStats != nil {
		fmt.Println("\n=== Validation Results ===")
		fmt.Printf("Total Validated: %d\n", result.ValidationStats.TotalValidated)
		fmt.Printf("Valid: %d\n", result.ValidationStats.ValidCount)
		fmt.Printf("Invalid: %d\n", result.ValidationStats.InvalidCount)
		fmt.Printf("Quality Score: %.2f\n", result.ValidationStats.QualityScore)
	}

	fmt.Println("\n=== Individual Scraper Results ===")
	for _, sr := range result.ScraperResults {
		status := "✓ SUCCESS"
		if len(sr.Errors) > 0 {
			status = "✗ FAILED"
		}
		fmt.Printf("%s - %s: scraped=%d validated=%d saved=%d save_failures=%d duration=%s\n",
			status, sr.ScraperName, sr.ItemsScraped, sr.Validation.Valid, sr.ItemsSaved, sr.SaveFailures, sr.Duration)
		if len(sr.Errors) > 0 {
			fmt.Printf("  Error: %s\n", sr.Errors[0])
		}
	}

	return nil
}

// validateData triggers data validation
func validateData(ctx context.Context, c client.Client, scraperNames string) error {
	logger.Info("Triggering data validation", "scrapers", scraperNames)

	workflowID := fmt.Sprintf("validation-%s", time.Now().Format("20060102-150405"))

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "cost-of-living-task-queue",
	}

	// Execute validation workflow (we'll create a simple one-shot validation workflow)
	// For now, we'll use a batch workflow with validation enabled
	input := workflow.BatchScraperWorkflowInput{
		ScraperNames: []string{}, // Empty means validate all recent data
		ValidateData: true,
		MaxRetries:   1,
		Timeout:      5 * time.Minute,
	}

	we, err := c.ExecuteWorkflow(ctx, options, workflow.BatchScraperWorkflow, input)

	if err != nil {
		return fmt.Errorf("failed to start validation: %w", err)
	}

	logger.Info("Validation started", "workflow_id", we.GetID())
	fmt.Printf("Validation started: %s\n", we.GetID())

	// Wait for results
	var result workflow.BatchScraperWorkflowResult
	err = we.Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Print results
	if result.ValidationStats != nil {
		fmt.Println("\n=== Validation Results ===")
		fmt.Printf("Total Validated: %d\n", result.ValidationStats.TotalValidated)
		fmt.Printf("Valid: %d\n", result.ValidationStats.ValidCount)
		fmt.Printf("Invalid: %d\n", result.ValidationStats.InvalidCount)
		fmt.Printf("Quality Score: %.2f\n", result.ValidationStats.QualityScore)
	}

	return nil
}

// checkStatus checks the status of a workflow
func checkStatus(ctx context.Context, c client.Client, workflowID string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow-id is required for status command")
	}

	logger.Info("Checking workflow status", "workflow_id", workflowID)

	// Query workflow
	resp, err := c.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		return fmt.Errorf("failed to describe workflow: %w", err)
	}

	fmt.Println("\n=== Workflow Status ===")
	fmt.Printf("Workflow ID: %s\n", workflowID)
	fmt.Printf("Status: %s\n", resp.WorkflowExecutionInfo.Status)
	fmt.Printf("Type: %s\n", resp.WorkflowExecutionInfo.Type.Name)
	fmt.Printf("Start Time: %s\n", resp.WorkflowExecutionInfo.StartTime)

	if resp.WorkflowExecutionInfo.CloseTime != nil {
		fmt.Printf("Close Time: %s\n", resp.WorkflowExecutionInfo.CloseTime)
		// Duration calculation for timestamppb is more complex, skip for now
		// fmt.Printf("Duration: %s\n", duration)
	}

	return nil
}

// startScheduler starts a scheduler workflow
func startScheduler(ctx context.Context, c client.Client, scheduleType string) error {
	logger.Info("Starting scheduler", "type", scheduleType)

	var workflowFunc interface{}
	var workflowID string

	switch scheduleType {
	case "daily":
		workflowFunc = workflow.DailyScraperWorkflow
		workflowID = "daily-scheduler"
	case "weekly":
		workflowFunc = workflow.WeeklyScraperWorkflow
		workflowID = "weekly-scheduler"
	case "monthly":
		workflowFunc = workflow.MonthlyScraperWorkflow
		workflowID = "monthly-scheduler"
	case "cron":
		workflowFunc = workflow.CronSchedulerWorkflow
		workflowID = "cron-scheduler"
	case "master":
		workflowFunc = workflow.MasterSchedulerWorkflow
		config := workflow.DefaultSchedulerConfig()
		workflowID = "master-scheduler"

		options := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "cost-of-living-task-queue",
		}

		we, err := c.ExecuteWorkflow(ctx, options, workflowFunc, config)
		if err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}

		fmt.Printf("Master scheduler started: %s\n", we.GetID())
		fmt.Println("This is a long-running workflow that will manage all scraper schedules.")
		return nil
	default:
		return fmt.Errorf("unknown schedule type: %s (use: daily, weekly, monthly, cron, master)", scheduleType)
	}

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "cost-of-living-task-queue",
	}

	we, err := c.ExecuteWorkflow(ctx, options, workflowFunc)
	if err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	logger.Info("Scheduler started", "workflow_id", we.GetID())
	fmt.Printf("Scheduler started: %s\n", we.GetID())

	return nil
}
