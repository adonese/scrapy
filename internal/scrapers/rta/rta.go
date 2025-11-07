package rta

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	rateLimit := 1
	if config.RateLimit > 0 {
		rateLimit = config.RateLimit
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	client := scrapers.BuildHTTPClient(config)

	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = DefaultRTAURL
	}

	return &RTAScraper{
		config:      config,
		url:         baseURL,
		client:      client,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), 1),
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

	doc, err := s.fetchDocument(ctx)
	if err != nil {
		return nil, err
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
	metrics.ScraperItemsScraped.WithLabelValues(s.Name()).Add(float64(len(dataPoints)))

	return dataPoints, nil
}

func (s *RTAScraper) fetchDocument(ctx context.Context) (*goquery.Document, error) {
	maxRetries := s.config.EffectiveMaxRetries()
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying RTA fetch", "attempt", attempt+1)
			if err := scrapers.WaitRetry(ctx, s.config, attempt-1); err != nil {
				return nil, err
			}
		}

		if err := scrapers.DelayBetweenRequests(ctx, s.config); err != nil {
			return nil, err
		}

		req, err := scrapers.PrepareRequest(ctx, http.MethodGet, s.url, nil, s.config)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			metrics.ScraperErrorsTotal.WithLabelValues("rta", "fetch").Inc()
			lastErr = fmt.Errorf("fetch page: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			metrics.ScraperErrorsTotal.WithLabelValues("rta", "blocked").Inc()
			lastErr = fmt.Errorf("blocked by anti-bot (status %d)", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			metrics.ScraperErrorsTotal.WithLabelValues("rta", "status").Inc()
			lastErr = fmt.Errorf("bad status: %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("parse html: %w", err)
			continue
		}

		return doc, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("failed after %d attempts", maxRetries)
}

// ScrapeWithRetry performs scraping with retry logic
func (s *RTAScraper) ScrapeWithRetry(ctx context.Context) ([]*models.CostDataPoint, error) {
	var lastErr error
	maxRetries := s.config.EffectiveMaxRetries()

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
			logger.Info("Retrying after backoff", "attempt", attempt)
			if err := scrapers.WaitRetry(ctx, s.config, attempt-1); err != nil {
				return nil, err
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
