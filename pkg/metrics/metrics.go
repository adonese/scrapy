package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts the total number of HTTP requests by method, endpoint, and status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// ScraperRunsTotal counts the total number of scraper runs
	ScraperRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scraper_runs_total",
			Help: "Total number of scraper runs",
		},
		[]string{"scraper", "status"},
	)

	// ScraperItemsScraped counts the total number of items scraped
	ScraperItemsScraped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scraper_items_scraped_total",
			Help: "Total number of items scraped",
		},
		[]string{"scraper"},
	)

	// ScraperErrorsTotal counts the total number of scraper errors
	ScraperErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scraper_errors_total",
			Help: "Total number of scraper errors",
		},
		[]string{"scraper", "error_type"},
	)

	// ScraperDuration measures the duration of scraper runs
	ScraperDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scraper_duration_seconds",
			Help:    "Duration of scraper runs",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"scraper"},
	)
)
