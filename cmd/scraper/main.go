package main

import (
	"context"
	"flag"
	"log"
	"os"

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
		UserAgent:  "Mozilla/5.0 (compatible; UAECostOfLiving/1.0; +http://localhost)",
		RateLimit:  1, // 1 request per second
		Timeout:    30,
		MaxRetries: 3,
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

	// Register Dubizzle scrapers (still only Dubai for now)
	if *scraperName == "dubizzle" || *scraperName == "all" {
		dubizzleScraper := dubizzle.NewDubizzleScraper(scraperConfig)
		service.RegisterScraper(dubizzleScraper)
		logger.Info("Registered Dubizzle scraper", "emirate", "Dubai")
	}

	logger.Info("Registered scrapers", "count", len(service.ListScrapers()))

	// Run scraper
	ctx := context.Background()

	if *scraperName == "all" {
		logger.Info("Running all scrapers")
		err = service.RunAllScrapers(ctx)
	} else {
		logger.Info("Running specific scraper", "name", *scraperName)
		err = service.RunScraper(ctx, *scraperName)
	}

	if err != nil {
		logger.Error("Scraper failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Scraping completed successfully")
}
