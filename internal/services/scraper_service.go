package services

import (
	"context"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/validation"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

// ScraperService manages and runs scrapers
type ScraperService struct {
	scrapers  []scrapers.Scraper
	repo      repository.CostDataPointRepository
	validator validation.Validator
	config    *ScraperServiceConfig
}

// ScraperServiceConfig holds configuration for the scraper service
type ScraperServiceConfig struct {
	EnableValidation   bool
	MinQualityScore    float64 // Minimum quality score to save data (default: 0.7)
	FailOnValidation   bool    // Fail scrape if validation fails
	ValidateBeforeSave bool    // Validate data before saving to DB
}

// DefaultScraperServiceConfig returns default configuration
func DefaultScraperServiceConfig() *ScraperServiceConfig {
	return &ScraperServiceConfig{
		EnableValidation:   true,
		MinQualityScore:    0.7,
		FailOnValidation:   false,
		ValidateBeforeSave: true,
	}
}

// NewScraperService creates a new scraper service with default configuration
func NewScraperService(repo repository.CostDataPointRepository) *ScraperService {
	return NewScraperServiceWithConfig(repo, DefaultScraperServiceConfig())
}

// NewScraperServiceWithConfig creates a new scraper service with custom configuration
func NewScraperServiceWithConfig(repo repository.CostDataPointRepository, config *ScraperServiceConfig) *ScraperService {
	return &ScraperService{
		scrapers:  []scrapers.Scraper{},
		repo:      repo,
		validator: validation.NewValidator(),
		config:    config,
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

	// Validate data points if enabled
	validatedPoints := dataPoints
	if s.config.EnableValidation && s.config.ValidateBeforeSave {
		validatedPoints = s.validateAndFilter(ctx, dataPoints, scraperName)
		logger.Info("Validation completed",
			"scraper", scraperName,
			"original", len(dataPoints),
			"validated", len(validatedPoints))
	}

	// Save to database
	saved := 0
	failed := 0
	for _, dp := range validatedPoints {
		if err := s.repo.Create(ctx, dp); err != nil {
			logger.Error("Failed to save data point", "error", err, "item", dp.ItemName)
			failed++
			continue
		}
		saved++
	}

	// Record metrics
	if failed > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "save_failed").Add(float64(failed))
	}

	metrics.ScraperRunsTotal.WithLabelValues(scraperName, "success").Inc()
	logger.Info("Scraper completed",
		"name", scraperName,
		"scraped", len(dataPoints),
		"validated", len(validatedPoints),
		"saved", saved,
		"failed", failed)

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

// validateAndFilter validates data points and filters out invalid ones
func (s *ScraperService) validateAndFilter(ctx context.Context, dataPoints []*models.CostDataPoint, scraperName string) []*models.CostDataPoint {
	if len(dataPoints) == 0 {
		return dataPoints
	}

	// Validate batch
	results, err := s.validator.ValidateBatch(ctx, dataPoints)
	if err != nil {
		logger.Error("Validation failed", "scraper", scraperName, "error", err)
		// Return original data points if validation fails and FailOnValidation is false
		if !s.config.FailOnValidation {
			return dataPoints
		}
		return nil
	}

	// Filter valid data points
	validPoints := make([]*models.CostDataPoint, 0, len(dataPoints))
	invalidCount := 0
	lowQualityCount := 0

	for i, result := range results {
		if result.IsValid && result.Score >= s.config.MinQualityScore {
			validPoints = append(validPoints, dataPoints[i])
		} else {
			if !result.IsValid {
				invalidCount++
				logger.Warn("Invalid data point",
					"scraper", scraperName,
					"item", dataPoints[i].ItemName,
					"errors", len(result.Errors))
			} else {
				lowQualityCount++
				logger.Warn("Low quality data point",
					"scraper", scraperName,
					"item", dataPoints[i].ItemName,
					"score", result.Score,
					"min_required", s.config.MinQualityScore)
			}
		}
	}

	// Record validation metrics
	if invalidCount > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "validation_failed").Add(float64(invalidCount))
	}
	if lowQualityCount > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "low_quality").Add(float64(lowQualityCount))
	}

	logger.Info("Validation filtering complete",
		"scraper", scraperName,
		"total", len(dataPoints),
		"valid", len(validPoints),
		"invalid", invalidCount,
		"low_quality", lowQualityCount)

	return validPoints
}
