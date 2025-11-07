package sewa

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
	// SEWAURL is the official SEWA tariff page
	SEWAURL = "https://www.sewa.gov.ae/en/content/tariff"
)

// SEWAScraper scrapes utility rates from SEWA (Sharjah Electricity, Water and Gas Authority)
type SEWAScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	baseURL     string
}

// NewSEWAScraper creates a new SEWA scraper
func NewSEWAScraper(config scrapers.Config) *SEWAScraper {
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
		baseURL = SEWAURL
	}

	return &SEWAScraper{
		config:      config,
		client:      client,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), 1),
		baseURL:     baseURL,
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

	doc, err := s.fetchDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Extract tariffs
	dataPoints, err := parseSEWATariffs(doc, s.baseURL)
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
	metrics.ScraperItemsScraped.WithLabelValues(s.Name()).Add(float64(len(dataPoints)))

	return dataPoints, nil
}

func (s *SEWAScraper) fetchDocument(ctx context.Context) (*goquery.Document, error) {
	maxRetries := s.config.EffectiveMaxRetries()
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying SEWA fetch", "attempt", attempt+1)
			if err := scrapers.WaitRetry(ctx, s.config, attempt-1); err != nil {
				return nil, err
			}
		}

		if err := scrapers.DelayBetweenRequests(ctx, s.config); err != nil {
			return nil, err
		}

		req, err := scrapers.PrepareRequest(ctx, http.MethodGet, s.baseURL, nil, s.config)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			metrics.ScraperErrorsTotal.WithLabelValues("sewa", "fetch").Inc()
			lastErr = fmt.Errorf("fetch page: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			metrics.ScraperErrorsTotal.WithLabelValues("sewa", "blocked").Inc()
			lastErr = fmt.Errorf("blocked by anti-bot (status %d)", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			metrics.ScraperErrorsTotal.WithLabelValues("sewa", "status").Inc()
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
