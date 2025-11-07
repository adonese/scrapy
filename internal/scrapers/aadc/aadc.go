package aadc

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
	// DefaultAADCURL is the default AADC tariff page
	DefaultAADCURL = "https://www.aadc.ae/en/pages/maintarrif.aspx"
)

// AADCScraper scrapes utility rates from AADC
type AADCScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	url         string
}

// NewAADCScraper creates a new AADC scraper
func NewAADCScraper(config scrapers.Config) *AADCScraper {
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
		baseURL = DefaultAADCURL
	}

	return &AADCScraper{
		config:      config,
		url:         baseURL,
		client:      client,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), 1),
	}
}

// NewAADCScraperWithURL creates an AADC scraper with a custom URL (useful for testing)
func NewAADCScraperWithURL(config scrapers.Config, url string) *AADCScraper {
	scraper := NewAADCScraper(config)
	scraper.url = url
	return scraper
}

// Name returns the scraper identifier
func (s *AADCScraper) Name() string {
	return "aadc"
}

// CanScrape checks if scraping is possible (rate limit)
func (s *AADCScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches utility rates from AADC
func (s *AADCScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting AADC scrape", "url", s.url)

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	doc, err := s.fetchDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Parse rates
	dataPoints, err := s.parseRates(doc)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("aadc", "parse").Inc()
		return nil, fmt.Errorf("parse rates: %w", err)
	}

	logger.Info("Completed AADC scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues(s.Name()).Add(float64(len(dataPoints)))

	return dataPoints, nil
}

func (s *AADCScraper) fetchDocument(ctx context.Context) (*goquery.Document, error) {
	maxRetries := s.config.EffectiveMaxRetries()
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying AADC fetch", "attempt", attempt+1)
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
			metrics.ScraperErrorsTotal.WithLabelValues("aadc", "fetch").Inc()
			lastErr = fmt.Errorf("fetch page: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			metrics.ScraperErrorsTotal.WithLabelValues("aadc", "blocked").Inc()
			lastErr = fmt.Errorf("blocked by anti-bot (status %d)", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			metrics.ScraperErrorsTotal.WithLabelValues("aadc", "status").Inc()
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

// parseRates extracts all utility rates from the document
func (s *AADCScraper) parseRates(doc *goquery.Document) ([]*models.CostDataPoint, error) {
	dataPoints := []*models.CostDataPoint{}

	// Parse electricity rates
	electricityRates, err := parseElectricityRates(doc)
	if err != nil {
		logger.Warn("Failed to parse electricity rates", "error", err)
	} else {
		logger.Info("Parsed electricity rates", "count", len(electricityRates))
		electricityPoints := convertElectricityToDataPoints(electricityRates, s.url)
		dataPoints = append(dataPoints, electricityPoints...)
	}

	// Parse water rates
	waterRates, err := parseWaterRates(doc)
	if err != nil {
		logger.Warn("Failed to parse water rates", "error", err)
	} else {
		logger.Info("Parsed water rates", "count", len(waterRates))
		waterPoints := convertWaterToDataPoints(waterRates, s.url)
		dataPoints = append(dataPoints, waterPoints...)
	}

	if len(dataPoints) == 0 {
		return nil, fmt.Errorf("no rates extracted from page")
	}

	return dataPoints, nil
}
