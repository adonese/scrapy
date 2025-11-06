package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/bayut"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/internal/workflow"
	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/adonese/cost-of-living/pkg/logger"
)

func main() {
	logger.Init()

	// Connect to database
	config := database.NewConfigFromEnv()
	db, err := database.Connect(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create repository
	repo := postgres.NewCostDataPointRepository(db.GetConn())

	// Create scraper service
	scraperService := services.NewScraperService(repo)

	// Register scrapers
	bayutConfig := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; UAECostOfLiving/1.0)",
		RateLimit:  1,
		Timeout:    30,
		MaxRetries: 3,
	}
	bayutScraper := bayut.NewBayutScraper(bayutConfig)
	scraperService.RegisterScraper(bayutScraper)

	// Set activity dependencies
	workflow.SetActivityDependencies(&workflow.ScraperActivityDependencies{
		ScraperService: scraperService,
		Repository:     repo,
	})

	// Get Temporal address from env
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: temporalAddress,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, "cost-of-living-task-queue", worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflow.HelloWorkflow)
	w.RegisterWorkflow(workflow.ScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledScraperWorkflow)

	// Register activities
	w.RegisterActivity(workflow.HelloActivity)
	w.RegisterActivity(workflow.RunScraperActivity)
	w.RegisterActivity(workflow.CompensateFailedScrapeActivity)

	logger.Info("Worker starting...", "queue", "cost-of-living-task-queue")

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
