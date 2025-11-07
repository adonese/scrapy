package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/aadc"
	"github.com/adonese/cost-of-living/internal/scrapers/bayut"
	"github.com/adonese/cost-of-living/internal/scrapers/careem"
	"github.com/adonese/cost-of-living/internal/scrapers/dewa"
	"github.com/adonese/cost-of-living/internal/scrapers/dubizzle"
	"github.com/adonese/cost-of-living/internal/scrapers/rta"
	"github.com/adonese/cost-of-living/internal/scrapers/sewa"
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

	// Create scraper service with validation enabled
	scraperService := services.NewScraperService(repo)

	// Configure scrapers
	scraperConfig := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; UAECostOfLiving/1.0)",
		RateLimit:  1,
		Timeout:    30,
		MaxRetries: 3,
	}

	// Register all scrapers
	allScrapers := registerAllScrapers(scraperService, scraperConfig)

	logger.Info("All scrapers registered", "total", allScrapers)

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

	// Register core workflows
	w.RegisterWorkflow(workflow.HelloWorkflow)
	w.RegisterWorkflow(workflow.ScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledScraperWorkflow)

	// Register batch and scheduler workflows
	w.RegisterWorkflow(workflow.BatchScraperWorkflow)
	w.RegisterWorkflow(workflow.DailyScraperWorkflow)
	w.RegisterWorkflow(workflow.WeeklyScraperWorkflow)
	w.RegisterWorkflow(workflow.MonthlyScraperWorkflow)
	w.RegisterWorkflow(workflow.MasterSchedulerWorkflow)
	w.RegisterWorkflow(workflow.CronSchedulerWorkflow)
	w.RegisterWorkflow(workflow.OnDemandScraperWorkflow)

	// Register utility scraper workflows
	w.RegisterWorkflow(workflow.DEWAScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledDEWAWorkflow)
	w.RegisterWorkflow(workflow.SEWAScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledSEWAWorkflow)
	w.RegisterWorkflow(workflow.AADCScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledAADCWorkflow)

	// Register transport scraper workflows
	w.RegisterWorkflow(workflow.RTAScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledRTAWorkflow)
	w.RegisterWorkflow(workflow.CareemScraperWorkflow)
	w.RegisterWorkflow(workflow.ScheduledCareemWorkflow)

	// Register core activities
	w.RegisterActivity(workflow.HelloActivity)
	w.RegisterActivity(workflow.RunScraperActivity)
	w.RegisterActivity(workflow.CompensateFailedScrapeActivity)

	// Register validation activities
	w.RegisterActivity(workflow.ValidateRecentDataActivity)
	w.RegisterActivity(workflow.ValidateScraperDataActivity)
	w.RegisterActivity(workflow.CheckDataFreshnessActivity)
	w.RegisterActivity(workflow.DetectOutliersActivity)
	w.RegisterActivity(workflow.CheckDuplicatesActivity)

	logger.Info("Worker starting...", "queue", "cost-of-living-task-queue")

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

// registerAllScrapers creates and registers all available scrapers
func registerAllScrapers(service *services.ScraperService, config scrapers.Config) int {
	count := 0
	emirates := []string{"Dubai", "Sharjah", "Ajman", "Abu Dhabi"}

	// Housing scrapers - Daily (run frequently, housing changes fast)
	// Register Bayut scrapers for all major emirates
	for _, emirate := range emirates {
		bayutScraper := bayut.NewBayutScraperForEmirate(config, emirate)
		service.RegisterScraper(bayutScraper)
		count++
	}

	// Register Dubizzle apartment scrapers for each emirate
	for _, emirate := range emirates {
		dubizzleScraper := dubizzle.NewDubizzleScraperFor(config, emirate, "apartmentflat")
		service.RegisterScraper(dubizzleScraper)
		count++
	}

	// Register shared accommodation scrapers (Dubai focus)
	dubizzleBedspace := dubizzle.NewDubizzleScraperFor(config, "Dubai", "bedspace")
	service.RegisterScraper(dubizzleBedspace)
	count++

	dubizzleRoomspace := dubizzle.NewDubizzleScraperFor(config, "Dubai", "roomspace")
	service.RegisterScraper(dubizzleRoomspace)
	count++

	// Utility scrapers - Weekly (rates change less frequently)
	// DEWA - Dubai Electricity and Water Authority
	dewaScraper := dewa.NewDEWAScraper(config)
	service.RegisterScraper(dewaScraper)
	count++

	// SEWA - Sharjah Electricity, Water and Gas Authority
	sewaScraper := sewa.NewSEWAScraper(config)
	service.RegisterScraper(sewaScraper)
	count++

	// AADC - Abu Dhabi Distribution Company
	aadcScraper := aadc.NewAADCScraper(config)
	service.RegisterScraper(aadcScraper)
	count++

	// Transportation scrapers
	// RTA - Dubai public transport (Weekly)
	rtaScraper := rta.NewRTAScraper(config)
	service.RegisterScraper(rtaScraper)
	count++

	// Careem - Ride-sharing (Monthly - rates rarely change)
	careemScraper := careem.NewCareemScraper(config)
	service.RegisterScraper(careemScraper)
	count++

	logger.Info("Scraper registration complete",
		"housing", 10,
		"utilities", 3,
		"transportation", 2,
		"total", count)

	return count
}
