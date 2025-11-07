package dewa

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

// DEWAScraper scrapes utility rates from Dubai Electricity and Water Authority
type DEWAScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	baseURL     string
}

// NewDEWAScraper creates a new DEWA scraper
func NewDEWAScraper(config scrapers.Config) *DEWAScraper {
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
		baseURL = "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff"
	}

	return &DEWAScraper{
		config:      config,
		client:      client,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), 1),
		baseURL:     baseURL,
	}
}

// Name returns the scraper identifier
func (s *DEWAScraper) Name() string {
	return "dewa"
}

// CanScrape checks if scraping is possible (rate limit)
func (s *DEWAScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches utility rates from DEWA
func (s *DEWAScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting DEWA scrape")

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	doc, err := s.fetchDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Extract data points
	dataPoints, err := s.extractRates(doc, s.baseURL)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("dewa", "parse").Inc()
		return nil, fmt.Errorf("extract rates: %w", err)
	}

	logger.Info("Completed DEWA scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues(s.Name()).Add(float64(len(dataPoints)))

	return dataPoints, nil
}

func (s *DEWAScraper) fetchDocument(ctx context.Context) (*goquery.Document, error) {
	maxRetries := s.config.EffectiveMaxRetries()
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying DEWA fetch", "attempt", attempt+1)
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
			metrics.ScraperErrorsTotal.WithLabelValues("dewa", "fetch").Inc()
			lastErr = fmt.Errorf("fetch page: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			metrics.ScraperErrorsTotal.WithLabelValues("dewa", "blocked").Inc()
			lastErr = fmt.Errorf("blocked by anti-bot (status %d)", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			metrics.ScraperErrorsTotal.WithLabelValues("dewa", "status").Inc()
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

// extractRates extracts all rate data from the DEWA page
func (s *DEWAScraper) extractRates(doc *goquery.Document, url string) ([]*models.CostDataPoint, error) {
	var dataPoints []*models.CostDataPoint

	// Parse electricity slabs
	electricitySlabs, err := parseElectricitySlabs(doc)
	if err != nil {
		logger.Warn("Failed to parse electricity slabs", "error", err)
	} else {
		for _, slab := range electricitySlabs {
			dp := slabToDataPoint(slab, "electricity", url)
			dataPoints = append(dataPoints, dp)
		}
	}

	// Parse water slabs
	waterSlabs, err := parseWaterSlabs(doc)
	if err != nil {
		logger.Warn("Failed to parse water slabs", "error", err)
	} else {
		for _, slab := range waterSlabs {
			dp := slabToDataPoint(slab, "water", url)
			dataPoints = append(dataPoints, dp)
		}
	}

	// Parse fuel surcharge
	if fuelSlab := parseFuelSurcharge(doc); fuelSlab != nil {
		dp := slabToDataPoint(*fuelSlab, "fuel_surcharge", url)
		dataPoints = append(dataPoints, dp)
	}

	if len(dataPoints) == 0 {
		return nil, fmt.Errorf("no rate data extracted")
	}

	return dataPoints, nil
}
