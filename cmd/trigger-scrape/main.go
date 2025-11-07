package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/adonese/cost-of-living/internal/workflow"
)

func main() {
	scraperName := flag.String("scraper", "bayut", "Scraper to run")
	scheduled := flag.Bool("scheduled", false, "Run all scrapers on schedule")
	flag.Parse()

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	if *scheduled {
		// Trigger scheduled workflow
		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("scheduled-scrape-%d", time.Now().Unix()),
			TaskQueue: "cost-of-living-task-queue",
		}

		we, err := c.ExecuteWorkflow(ctx, workflowOptions, workflow.ScheduledScraperWorkflow)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Started scheduled workflow: %s\n", we.GetID())

		// Wait for result
		err = we.Get(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Scheduled workflow completed successfully")
	} else {
		// Trigger single scraper workflow
		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("scrape-%s-%d", *scraperName, time.Now().Unix()),
			TaskQueue: "cost-of-living-task-queue",
		}

		input := workflow.ScraperWorkflowInput{
			ScraperName: *scraperName,
			MaxRetries:  3,
		}

		we, err := c.ExecuteWorkflow(ctx, workflowOptions, workflow.ScraperWorkflow, input)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Started scraper workflow: %s\n", we.GetID())

		// Wait for result
		var result workflow.ScraperWorkflowResult
		err = we.Get(ctx, &result)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nWorkflow completed:\n")
		fmt.Printf("  Scraper: %s\n", result.ScraperName)
		fmt.Printf("  Items Scraped: %d\n", result.ItemsScraped)
		fmt.Printf("  Items Validated: %d\n", result.Validation.Valid)
		fmt.Printf("  Items Saved: %d\n", result.ItemsSaved)
		fmt.Printf("  Save Failures: %d\n", result.SaveFailures)
		if result.Validation.Invalid > 0 || result.Validation.LowQuality > 0 {
			fmt.Printf("  Dropped (invalid/low_quality): %d/%d\n", result.Validation.Invalid, result.Validation.LowQuality)
		}
		fmt.Printf("  Duration: %s\n", result.Duration)
		fmt.Printf("  Completed: %s\n", result.CompletedAt)

		if len(result.Errors) > 0 {
			fmt.Printf("  Errors: %v\n", result.Errors)
		}
	}
}
