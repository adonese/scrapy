package services

import (
	"context"
	"errors"
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

// ValidationSummary captures how many data points passed validation and why
// others were discarded.
type ValidationSummary struct {
	Total      int
	Valid      int
	Invalid    int
	LowQuality int
	Skipped    bool
}

// ScrapeResult represents the outcome of running a scraper end-to-end.
type ScrapeResult struct {
	ScraperName  string
	Fetched      int
	Validation   ValidationSummary
	Saved        int
	SaveFailures int
	Duration     time.Duration
	Errors       []error
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

// RunScraper runs a specific scraper by name and returns a detailed summary.
func (s *ScraperService) RunScraper(ctx context.Context, scraperName string) (*ScrapeResult, error) {
	start := time.Now()
	result := &ScrapeResult{ScraperName: scraperName}

	// Find scraper
	var targetScraper scrapers.Scraper
	for _, scraper := range s.scrapers {
		if scraper.Name() == scraperName {
			targetScraper = scraper
			break
		}
	}

	if targetScraper == nil {
		err := fmt.Errorf("scraper not found: %s", scraperName)
		result.Errors = append(result.Errors, err)
		result.Duration = time.Since(start)
		return result, err
	}

	// Check if the scraper can run within rate limits
	if !targetScraper.CanScrape() {
		err := fmt.Errorf("rate limit exceeded")
		metrics.ScraperRunsTotal.WithLabelValues(scraperName, "error").Inc()
		result.Errors = append(result.Errors, err)
		result.Duration = time.Since(start)
		return result, err
	}

	logger.Info("Running scraper", "name", scraperName)

	// Execute scraper
	dataPoints, err := targetScraper.Scrape(ctx)
	result.Fetched = len(dataPoints)

	if err != nil {
		metrics.ScraperRunsTotal.WithLabelValues(scraperName, "error").Inc()
		wrapped := fmt.Errorf("scrape failed: %w", err)
		result.Errors = append(result.Errors, wrapped)
		result.Duration = time.Since(start)
		return result, wrapped
	}

	// Validate data points when enabled
	validatedPoints := dataPoints
	summary := ValidationSummary{Total: len(dataPoints), Valid: len(dataPoints), Skipped: true}
	if s.config.EnableValidation && s.config.ValidateBeforeSave {
		summary.Skipped = false
		var validationErr error
		validatedPoints, summary, validationErr = s.validateAndFilter(ctx, dataPoints, scraperName)
		logger.Info("Validation completed",
			"scraper", scraperName,
			"total", summary.Total,
			"valid", summary.Valid,
			"invalid", summary.Invalid,
			"low_quality", summary.LowQuality,
			"skipped", summary.Skipped)

		if validationErr != nil {
			metrics.ScraperRunsTotal.WithLabelValues(scraperName, "error").Inc()
			result.Validation = summary
			result.Errors = append(result.Errors, validationErr)
			result.Duration = time.Since(start)
			return result, validationErr
		}
	}
	result.Validation = summary

	// Persist validated data points
	saved := 0
	failed := 0
	for _, dp := range validatedPoints {
		if err := s.repo.Create(ctx, dp); err != nil {
			logger.Error("Failed to save data point", "error", err, "item", dp.ItemName)
			failed++
			result.Errors = append(result.Errors, err)
			continue
		}
		saved++
	}

	result.Saved = saved
	result.SaveFailures = failed

	if failed > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "save_failed").Add(float64(failed))
	}

	result.Duration = time.Since(start)
	metrics.ScraperDuration.WithLabelValues(scraperName).Observe(result.Duration.Seconds())

	metrics.ScraperRunsTotal.WithLabelValues(scraperName, "success").Inc()
	logger.Info("Scraper completed",
		"name", scraperName,
		"scraped", result.Fetched,
		"validated", result.Validation.Valid,
		"dropped_invalid", result.Validation.Invalid,
		"dropped_low_quality", result.Validation.LowQuality,
		"saved", result.Saved,
		"failed", result.SaveFailures,
		"duration", result.Duration)

	return result, nil
}

// RunAllScrapers runs all registered scrapers
func (s *ScraperService) RunAllScrapers(ctx context.Context) ([]*ScrapeResult, error) {
	results := make([]*ScrapeResult, 0, len(s.scrapers))
	var errs []error

	for _, scraper := range s.scrapers {
		result, err := s.RunScraper(ctx, scraper.Name())
		if result != nil {
			results = append(results, result)
		}
		if err != nil {
			logger.Error("Scraper failed", "name", scraper.Name(), "error", err)
			errs = append(errs, fmt.Errorf("%s: %w", scraper.Name(), err))
		}
	}

	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}

	return results, nil
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
func (s *ScraperService) validateAndFilter(ctx context.Context, dataPoints []*models.CostDataPoint, scraperName string) ([]*models.CostDataPoint, ValidationSummary, error) {
	summary := ValidationSummary{Total: len(dataPoints)}

	if len(dataPoints) == 0 {
		summary.Valid = 0
		summary.Skipped = true
		return dataPoints, summary, nil
	}

	if !s.config.EnableValidation || !s.config.ValidateBeforeSave {
		summary.Valid = len(dataPoints)
		summary.Skipped = true
		return dataPoints, summary, nil
	}

	// Validate batch
	results, err := s.validator.ValidateBatch(ctx, dataPoints)
	if err != nil {
		logger.Error("Validation failed", "scraper", scraperName, "error", err)
		if !s.config.FailOnValidation {
			summary.Valid = len(dataPoints)
			summary.Skipped = true
			return dataPoints, summary, nil
		}

		summary.Invalid = len(dataPoints)
		return nil, summary, fmt.Errorf("validation failed: %w", err)
	}

	// Filter valid data points
	validPoints := make([]*models.CostDataPoint, 0, len(dataPoints))

	for i, result := range results {
		if result.IsValid && result.Score >= s.config.MinQualityScore {
			validPoints = append(validPoints, dataPoints[i])
			continue
		}

		if !result.IsValid {
			summary.Invalid++
			logger.Warn("Invalid data point",
				"scraper", scraperName,
				"item", dataPoints[i].ItemName,
				"errors", len(result.Errors))
		} else {
			summary.LowQuality++
			logger.Warn("Low quality data point",
				"scraper", scraperName,
				"item", dataPoints[i].ItemName,
				"score", result.Score,
				"min_required", s.config.MinQualityScore)
		}
	}

	summary.Valid = len(validPoints)

	// Record validation metrics
	if summary.Invalid > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "validation_failed").Add(float64(summary.Invalid))
	}
	if summary.LowQuality > 0 {
		metrics.ScraperErrorsTotal.WithLabelValues(scraperName, "low_quality").Add(float64(summary.LowQuality))
	}

	logger.Info("Validation filtering complete",
		"scraper", scraperName,
		"total", summary.Total,
		"valid", summary.Valid,
		"invalid", summary.Invalid,
		"low_quality", summary.LowQuality,
		"skipped", summary.Skipped)

	return validPoints, summary, nil
}
