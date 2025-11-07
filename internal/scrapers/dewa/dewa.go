package dewa

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

// DEWAScraper scrapes utility rates from Dubai Electricity and Water Authority
type DEWAScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
}

// NewDEWAScraper creates a new DEWA scraper
func NewDEWAScraper(config scrapers.Config) *DEWAScraper {
	return &DEWAScraper{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
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

	// DEWA official tariff page
	url := "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff"

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch the page
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("dewa", "fetch").Inc()
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		metrics.ScraperErrorsTotal.WithLabelValues("dewa", "status").Inc()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Extract data points
	dataPoints, err := s.extractRates(doc, url)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("dewa", "parse").Inc()
		return nil, fmt.Errorf("extract rates: %w", err)
	}

	logger.Info("Completed DEWA scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues("dewa").Add(float64(len(dataPoints)))

	return dataPoints, nil
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
