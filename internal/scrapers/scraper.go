package scrapers

import (
	"context"

	"github.com/adonese/cost-of-living/internal/models"
)

// Scraper defines the interface for all scrapers
type Scraper interface {
	// Name returns the scraper identifier
	Name() string

	// Scrape fetches data and returns cost data points
	Scrape(ctx context.Context) ([]*models.CostDataPoint, error)

	// CanScrape checks if scraping is possible (rate limit, etc)
	CanScrape() bool
}

// Config holds common scraper configuration
type Config struct {
	UserAgent  string
	RateLimit  int    // requests per second
	Timeout    int    // seconds
	MaxRetries int
	ProxyURL   string // optional
}
