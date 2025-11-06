package services

import (
	"context"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

// ScraperService manages and runs scrapers
type ScraperService struct {
	scrapers []scrapers.Scraper
	repo     repository.CostDataPointRepository
}

// NewScraperService creates a new scraper service
func NewScraperService(repo repository.CostDataPointRepository) *ScraperService {
	return &ScraperService{
		scrapers: []scrapers.Scraper{},
		repo:     repo,
	}
}

// RegisterScraper adds a scraper to the service
func (s *ScraperService) RegisterScraper(scraper scrapers.Scraper) {
	s.scrapers = append(s.scrapers, scraper)
	logger.Info("Registered scraper", "name", scraper.Name())
}

// RunScraper runs a specific scraper by name
func (s *ScraperService) RunScraper(ctx context.Context, scraperName string) error {
	start := time.Now()

	// Find scraper
	var targetScraper scrapers.Scraper
	for _, scraper := range s.scrapers {
		if scraper.Name() == scraperName {
			targetScraper = scraper
			break
		}
	}

	if targetScraper == nil {
		return fmt.Errorf("scraper not found: %s", scraperName)
	}

	// Check if can scrape
	if !targetScraper.CanScrape() {
		return fmt.Errorf("rate limit exceeded")
	}

	// Run scraper
	logger.Info("Running scraper", "name", scraperName)
	dataPoints, err := targetScraper.Scrape(ctx)

	// Record duration
	duration := time.Since(start).Seconds()
	metrics.ScraperDuration.WithLabelValues(scraperName).Observe(duration)

	if err != nil {
		metrics.ScraperRunsTotal.WithLabelValues(scraperName, "error").Inc()
		return fmt.Errorf("scrape failed: %w", err)
	}

	// Save to database
	saved := 0
	for _, dp := range dataPoints {
		if err := s.repo.Create(ctx, dp); err != nil {
			logger.Error("Failed to save data point", "error", err, "item", dp.ItemName)
			continue
		}
		saved++
	}

	metrics.ScraperRunsTotal.WithLabelValues(scraperName, "success").Inc()
	logger.Info("Scraper completed", "name", scraperName, "scraped", len(dataPoints), "saved", saved)

	return nil
}

// RunAllScrapers runs all registered scrapers
func (s *ScraperService) RunAllScrapers(ctx context.Context) error {
	for _, scraper := range s.scrapers {
		if err := s.RunScraper(ctx, scraper.Name()); err != nil {
			logger.Error("Scraper failed", "name", scraper.Name(), "error", err)
			// Continue with other scrapers
		}
	}
	return nil
}

// ListScrapers returns the names of all registered scrapers
func (s *ScraperService) ListScrapers() []string {
	names := make([]string, len(s.scrapers))
	for i, scraper := range s.scrapers {
		names[i] = scraper.Name()
	}
	return names
}
