package rta

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

const (
	// DefaultRTAURL is the default URL for RTA fare information
	DefaultRTAURL = "https://www.rta.ae/wps/portal/rta/ae/public-transport/fares"
)

// RTAScraper scrapes public transport fare data from Dubai RTA
type RTAScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	url         string
}

// NewRTAScraper creates a new RTA scraper
func NewRTAScraper(config scrapers.Config) *RTAScraper {
	return &RTAScraper{
		config: config,
		url:    DefaultRTAURL,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
	}
}

// NewRTAScraperWithURL creates a new RTA scraper with a custom URL
// This is useful for testing with mock servers
func NewRTAScraperWithURL(config scrapers.Config, url string) *RTAScraper {
	scraper := NewRTAScraper(config)
	scraper.url = url
	return scraper
}

// Name returns the scraper identifier
func (s *RTAScraper) Name() string {
	return "rta"
}

// CanScrape checks if scraping is possible (rate limit)
func (s *RTAScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches public transport fare data from RTA
func (s *RTAScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting RTA scrape", "url", s.url)

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch the page
	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("rta", "fetch").Inc()
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		metrics.ScraperErrorsTotal.WithLabelValues("rta", "status").Inc()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Parse fares from the document
	dataPoints, err := ParseFares(doc, s.url)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("rta", "parse").Inc()
		return nil, fmt.Errorf("parse fares: %w", err)
	}

	// Validate we got reasonable data
	if len(dataPoints) == 0 {
		metrics.ScraperErrorsTotal.WithLabelValues("rta", "no_data").Inc()
		return nil, fmt.Errorf("no fare data extracted")
	}

	logger.Info("Completed RTA scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues("rta").Add(float64(len(dataPoints)))

	return dataPoints, nil
}

// ScrapeWithRetry performs scraping with retry logic
func (s *RTAScraper) ScrapeWithRetry(ctx context.Context) ([]*models.CostDataPoint, error) {
	var lastErr error
	maxRetries := s.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info("RTA scrape attempt", "attempt", attempt, "max_retries", maxRetries)

		dataPoints, err := s.Scrape(ctx)
		if err == nil {
			return dataPoints, nil
		}

		lastErr = err
		logger.Warn("RTA scrape failed", "attempt", attempt, "error", err)

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff before retry
		if attempt < maxRetries {
			backoff := time.Duration(attempt*attempt) * time.Second
			logger.Info("Retrying after backoff", "backoff", backoff)

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("scraping failed after %d attempts: %w", maxRetries, lastErr)
}

// ValidateFareData validates that the extracted fare data makes sense
func ValidateFareData(dataPoints []*models.CostDataPoint) error {
	if len(dataPoints) == 0 {
		return fmt.Errorf("no data points")
	}

	// Count by category
	metroCount := 0
	busCount := 0
	tramCount := 0
	taxiCount := 0

	for _, dp := range dataPoints {
		if dp.Price <= 0 {
			return fmt.Errorf("invalid price %f for item %s", dp.Price, dp.ItemName)
		}

		if dp.Category != "Transportation" {
			return fmt.Errorf("invalid category %s for RTA data", dp.Category)
		}

		// Count by mode (check item name)
		itemLower := dp.ItemName
		if containsIgnoreCase(itemLower, "metro") {
			metroCount++
		} else if containsIgnoreCase(itemLower, "bus") {
			busCount++
		} else if containsIgnoreCase(itemLower, "tram") {
			tramCount++
		} else if containsIgnoreCase(itemLower, "taxi") {
			taxiCount++
		}
	}

	// We expect at least some metro and bus fares
	if metroCount == 0 {
		return fmt.Errorf("no metro fares found")
	}

	logger.Info("Fare data validation passed",
		"total", len(dataPoints),
		"metro", metroCount,
		"bus", busCount,
		"tram", tramCount,
		"taxi", taxiCount,
	)

	return nil
}

// containsIgnoreCase checks if a string contains a substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// contains checks if string s contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring finds the index of substr in s, or -1 if not found
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
