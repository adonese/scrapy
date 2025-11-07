package sewa

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
	// SEWAURL is the official SEWA tariff page
	SEWAURL = "https://www.sewa.gov.ae/en/content/tariff"
)

// SEWAScraper scrapes utility rates from SEWA (Sharjah Electricity, Water and Gas Authority)
type SEWAScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
}

// NewSEWAScraper creates a new SEWA scraper
func NewSEWAScraper(config scrapers.Config) *SEWAScraper {
	return &SEWAScraper{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
	}
}

// Name returns the scraper identifier
func (s *SEWAScraper) Name() string {
	return "sewa"
}

// CanScrape checks if scraping is possible (rate limit)
func (s *SEWAScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches utility rates from SEWA
func (s *SEWAScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting SEWA scrape")

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch the page
	req, err := http.NewRequestWithContext(ctx, "GET", SEWAURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("sewa", "fetch").Inc()
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		metrics.ScraperErrorsTotal.WithLabelValues("sewa", "status").Inc()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Extract tariffs
	dataPoints, err := parseSEWATariffs(doc, SEWAURL)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("sewa", "parse").Inc()
		return nil, fmt.Errorf("parse tariffs: %w", err)
	}

	// Validate we got some data
	if len(dataPoints) == 0 {
		metrics.ScraperErrorsTotal.WithLabelValues("sewa", "no_data").Inc()
		return nil, fmt.Errorf("no tariff data found")
	}

	logger.Info("Completed SEWA scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues("sewa").Add(float64(len(dataPoints)))

	return dataPoints, nil
}

// ScrapeFromHTML is a helper method for testing that allows scraping from an HTML document
func (s *SEWAScraper) ScrapeFromHTML(doc *goquery.Document, sourceURL string) ([]*models.CostDataPoint, error) {
	dataPoints, err := parseSEWATariffs(doc, sourceURL)
	if err != nil {
		return nil, fmt.Errorf("parse tariffs: %w", err)
	}

	if len(dataPoints) == 0 {
		return nil, fmt.Errorf("no tariff data found")
	}

	return dataPoints, nil
}
