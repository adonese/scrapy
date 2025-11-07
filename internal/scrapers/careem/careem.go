package careem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

// CareemScraper scrapes Careem ride-sharing rates from multiple sources
type CareemScraper struct {
	config      scrapers.Config
	aggregator  *SourceAggregator
	rateLimiter *rate.Limiter
	lastRates   *CareemRates
	mu          sync.RWMutex
}

// NewCareemScraper creates a new Careem scraper with multiple sources
func NewCareemScraper(config scrapers.Config) *CareemScraper {
	return NewCareemScraperWithSources(config, nil)
}

// NewCareemScraperWithSources creates a new Careem scraper with custom sources
func NewCareemScraperWithSources(config scrapers.Config, customSources []RateSource) *CareemScraper {
	var sources []RateSource

	if customSources != nil && len(customSources) > 0 {
		sources = customSources
	} else {
		// Default sources in priority order
		sources = []RateSource{
			// Highest priority: Official API (if available)
			NewAPISource(""), // No API key by default

			// High priority: Help center
			NewHelpCenterSource(config.UserAgent),

			// Medium priority: News sources
			NewNewsSource(config.UserAgent),

			// Fallback: Static file with known rates
			NewStaticSource(GetDefaultStaticSourcePath()),
		}
	}

	return &CareemScraper{
		config:      config,
		aggregator:  NewSourceAggregator(sources),
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
	}
}

// Name returns the scraper identifier
func (s *CareemScraper) Name() string {
	return "careem"
}

// CanScrape checks if scraping is possible (rate limit)
func (s *CareemScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches Careem rates from multiple sources and returns cost data points
func (s *CareemScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting Careem scrape")
	startTime := time.Now()

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch rates from best available source
	rates, err := s.aggregator.FetchBestRates(ctx)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("careem", "fetch").Inc()
		return nil, fmt.Errorf("fetch rates: %w", err)
	}

	// Validate rates
	if err := ValidateRates(rates); err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("careem", "validation").Inc()
		return nil, fmt.Errorf("validate rates: %w", err)
	}

	// Check for significant rate changes
	s.checkRateChanges(rates)

	// Store current rates for change detection
	s.mu.Lock()
	s.lastRates = rates
	s.mu.Unlock()

	// Convert rates to data points
	dataPoints, err := ParseRatesToDataPoints(rates)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("careem", "parse").Inc()
		return nil, fmt.Errorf("parse rates: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Completed Careem scrape",
		"count", len(dataPoints),
		"source", rates.Source,
		"confidence", rates.Confidence,
		"duration", duration)

	metrics.ScraperItemsScraped.WithLabelValues("careem").Add(float64(len(dataPoints)))
	metrics.ScraperDuration.WithLabelValues("careem").Observe(duration.Seconds())

	return dataPoints, nil
}

// checkRateChanges detects and logs significant rate changes
func (s *CareemScraper) checkRateChanges(newRates *CareemRates) {
	s.mu.RLock()
	oldRates := s.lastRates
	s.mu.RUnlock()

	if oldRates == nil {
		logger.Info("No previous rates to compare")
		return
	}

	// Detect changes > 10%
	changes := DetectRateChange(oldRates, newRates, 10.0)
	if len(changes) > 0 {
		logger.Info("Significant rate changes detected", "changes", changes)
		for _, change := range changes {
			logger.Info("Rate change", "detail", change)
		}
	}
}

// GetLastRates returns the last fetched rates (for testing and monitoring)
func (s *CareemScraper) GetLastRates() *CareemRates {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRates
}

// EstimateFare provides a fare estimate based on current rates
func (s *CareemScraper) EstimateFare(distanceKm float64, waitTimeMinutes float64, isPeakHour bool, isAirport bool, salikGates int) (float64, error) {
	s.mu.RLock()
	rates := s.lastRates
	s.mu.RUnlock()

	if rates == nil {
		return 0, fmt.Errorf("no rates available, please scrape first")
	}

	fare := EstimateFare(rates, distanceKm, waitTimeMinutes, isPeakHour, isAirport, salikGates)
	return fare, nil
}

// GetRatesSummary returns a human-readable summary of current rates
func (s *CareemScraper) GetRatesSummary() (string, error) {
	s.mu.RLock()
	rates := s.lastRates
	s.mu.RUnlock()

	if rates == nil {
		return "", fmt.Errorf("no rates available")
	}

	summary := fmt.Sprintf(`Careem Rates Summary
=====================
Service: %s
Emirate: %s
Source: %s (Confidence: %.1f%%)
Effective Date: %s

Base Rates:
- Base Fare: %.2f AED
- Per Kilometer: %.2f AED
- Per Minute Wait: %.2f AED
- Minimum Fare: %.2f AED

Surcharges:
- Peak Hour Multiplier: %.2fx
- Airport Surcharge: %.2f AED
- Salik Toll (per gate): %.2f AED

Last Updated: %s
`,
		rates.ServiceType,
		rates.Emirate,
		rates.Source,
		rates.Confidence*100,
		rates.EffectiveDate,
		rates.BaseFare,
		rates.PerKm,
		rates.PerMinuteWait,
		rates.MinimumFare,
		rates.PeakSurchargeMultiplier,
		rates.AirportSurcharge,
		rates.SalikToll,
		rates.LastUpdated.Format("2006-01-02 15:04:05"),
	)

	// Add service-specific rates if available
	if len(rates.Rates) > 0 {
		summary += "\nService-Specific Rates:\n"
		for _, serviceRate := range rates.Rates {
			summary += fmt.Sprintf("\n%s (%s):\n", serviceRate.ServiceType, serviceRate.Description)
			summary += fmt.Sprintf("  Base: %.2f AED, Per km: %.2f AED, Min: %.2f AED\n",
				serviceRate.BaseFare, serviceRate.PerKm, serviceRate.MinimumFare)
		}
	}

	return summary, nil
}

// RefreshRates forces a refresh of rates from sources
func (s *CareemScraper) RefreshRates(ctx context.Context) error {
	logger.Info("Forcing rate refresh")
	_, err := s.Scrape(ctx)
	return err
}
