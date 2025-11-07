package aadc

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
	return &AADCScraper{
		config: config,
		url:    DefaultAADCURL,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
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

	// Fetch the page
	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("aadc", "fetch").Inc()
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		metrics.ScraperErrorsTotal.WithLabelValues("aadc", "status").Inc()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Parse rates
	dataPoints, err := s.parseRates(doc)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("aadc", "parse").Inc()
		return nil, fmt.Errorf("parse rates: %w", err)
	}

	logger.Info("Completed AADC scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues("aadc").Add(float64(len(dataPoints)))

	return dataPoints, nil
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
