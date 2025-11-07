package scrapers

import (
	"context"
	"net/http"
	"time"

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
	UserAgent string
	// UserAgents allows providing a pool of User-Agent strings that will be
	// rotated for every outbound request. When empty, UserAgent or a default
	// desktop browser string will be used instead.
	UserAgents []string
	RateLimit  int // requests per second
	Timeout    int // seconds
	MaxRetries int
	ProxyURL   string // optional
	// ExtraHeaders are added to every outbound request. They can be used to
	// inject cookies or other negotiated headers required by a target site.
	ExtraHeaders map[string]string

	// MinDelayBetweenRequests defines the minimum amount of jitter to wait
	// before issuing a new HTTP request. This helps mimic human browsing
	// behaviour and reduces the risk of triggering rate-limiters when multiple
	// requests are needed during a scrape.
	MinDelayBetweenRequests time.Duration

	// MaxDelayBetweenRequests defines the upper bound of random jitter between
	// requests. When zero, MinDelayBetweenRequests is used as the fixed delay.
	// If both are zero, no additional delay is introduced beyond explicit rate
	// limiters in each scraper.
	MaxDelayBetweenRequests time.Duration

	// RetryBaseDelay defines the base duration used when backing off between
	// retry attempts. Exponential backoff with jitter is applied on top of this
	// base delay. When zero, a sensible default is used.
	RetryBaseDelay time.Duration

	// BaseURL allows overriding the default host that a scraper targets.
	// When empty, each scraper falls back to its production domain. Tests
	// can supply the address of an httptest.Server to exercise the full
	// scraping flow without real network calls.
	BaseURL string

	// HTTPClient lets callers provide a custom HTTP client (for example
	// with specialized transports or to hook into httptest servers). When
	// nil, scrapers instantiate a standard http.Client using the Timeout
	// defined above.
	HTTPClient *http.Client
}
