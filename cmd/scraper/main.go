package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/bayut"
	"github.com/adonese/cost-of-living/internal/scrapers/dubizzle"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/adonese/cost-of-living/pkg/logger"
)

func main() {
	scraperName := flag.String("scraper", "bayut", "Scraper to run (bayut, dubizzle, all)")
	emirate := flag.String("emirate", "all", "Emirate to scrape (Dubai, Sharjah, Ajman, Abu Dhabi, all)")
	flag.Parse()

	// Initialize logger
	logger.Init()
	logger.Info("Starting scraper CLI", "scraper", *scraperName, "emirate", *emirate)

	// Connect to database
	config := database.NewConfigFromEnv()
	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Connected to database successfully")

	// Create repository
	repo := postgres.NewCostDataPointRepository(db.GetConn())

	// Create scraper service
	service := services.NewScraperService(repo)

	// Register scrapers
	scraperConfig := scrapers.Config{
		UserAgent: "Mozilla/5.0 (compatible; UAECostOfLiving/1.0; +http://localhost)",
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
		RateLimit:               1, // 1 request per second
		Timeout:                 30,
		MaxRetries:              3,
		MinDelayBetweenRequests: 500 * time.Millisecond,
		MaxDelayBetweenRequests: 2 * time.Second,
		RetryBaseDelay:          2 * time.Second,
	}

	// Determine which emirates to register
	emirates := []string{}
	if *emirate == "all" {
		emirates = []string{"Dubai", "Sharjah", "Ajman", "Abu Dhabi"}
	} else {
		emirates = []string{*emirate}
	}

	// Register Bayut scrapers for requested emirates
	if *scraperName == "bayut" || *scraperName == "all" {
		for _, em := range emirates {
			bayutScraper := bayut.NewBayutScraperForEmirate(scraperConfig, em)
			service.RegisterScraper(bayutScraper)
			logger.Info("Registered Bayut scraper", "emirate", em)
		}
	}

	// Register Dubizzle scrapers for requested emirates
	if *scraperName == "dubizzle" || *scraperName == "all" {
		// Regular apartments
		for _, em := range emirates {
			dubizzleScraper := dubizzle.NewDubizzleScraperFor(scraperConfig, em, "apartmentflat")
			service.RegisterScraper(dubizzleScraper)
			logger.Info("Registered Dubizzle scraper", "emirate", em, "category", "apartments")
		}

		// Shared accommodations (bedspace and roomspace) - Dubai only initially
		dubizzleBedspace := dubizzle.NewDubizzleScraperFor(scraperConfig, "Dubai", "bedspace")
		service.RegisterScraper(dubizzleBedspace)
		logger.Info("Registered Dubizzle scraper", "emirate", "Dubai", "category", "bedspace")

		dubizzleRoomspace := dubizzle.NewDubizzleScraperFor(scraperConfig, "Dubai", "roomspace")
		service.RegisterScraper(dubizzleRoomspace)
		logger.Info("Registered Dubizzle scraper", "emirate", "Dubai", "category", "roomspace")
	}

	logger.Info("Registered scrapers", "count", len(service.ListScrapers()))

	// Run scraper
	ctx := context.Background()

	if *scraperName == "all" {
		logger.Info("Running all scrapers")
		results, runErr := service.RunAllScrapers(ctx)
		for _, res := range results {
			logScrapeResult(res)
		}
		if runErr != nil {
			logger.Error("One or more scrapers failed", "error", runErr)
			os.Exit(1)
		}
	} else {
		logger.Info("Running specific scraper", "name", *scraperName)
		result, runErr := service.RunScraper(ctx, *scraperName)
		if result != nil {
			logScrapeResult(result)
		}
		if runErr != nil {
			logger.Error("Scraper failed", "error", runErr)
			os.Exit(1)
		}
	}

	logger.Info("Scraping completed successfully")
}

func logScrapeResult(res *services.ScrapeResult) {
	if res == nil {
		return
	}

	logger.Info("Scraper summary",
		"scraper", res.ScraperName,
		"fetched", res.Fetched,
		"validated", res.Validation.Valid,
		"dropped_invalid", res.Validation.Invalid,
		"dropped_low_quality", res.Validation.LowQuality,
		"validation_skipped", res.Validation.Skipped,
		"saved", res.Saved,
		"save_failures", res.SaveFailures,
		"duration", res.Duration)

	for _, err := range res.Errors {
		logger.Warn("Scraper encountered issue",
			"scraper", res.ScraperName,
			"error", err)
	}
}
