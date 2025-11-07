package smoke

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/repository/mock"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/aadc"
	"github.com/adonese/cost-of-living/internal/scrapers/bayut"
	"github.com/adonese/cost-of-living/internal/scrapers/dewa"
	"github.com/adonese/cost-of-living/internal/scrapers/dubizzle"
	"github.com/adonese/cost-of-living/internal/scrapers/rta"
	"github.com/adonese/cost-of-living/internal/scrapers/sewa"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/test/helpers"
)

func TestScraperSmokeSuite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Init()

	baseConfig := scrapers.Config{
		UserAgent: "SmokeTestBot/1.0",
		UserAgents: []string{
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		},
		RateLimit:               10,
		Timeout:                 10,
		MaxRetries:              3,
		MinDelayBetweenRequests: 5 * time.Millisecond,
		MaxDelayBetweenRequests: 15 * time.Millisecond,
		RetryBaseDelay:          10 * time.Millisecond,
	}

	type scraperCase struct {
		name  string
		setup func(t *testing.T) (*helpers.MockServer, string)
		build func(scrapers.Config) scrapers.Scraper
	}

	cases := []scraperCase{
		{
			name: "bayut",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server, err := helpers.NewBayutMockServer()
				require.NoError(t, err)
				return server, ""
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return bayut.NewBayutScraperForEmirate(cfg, "Dubai")
			},
		},
		{
			name: "dubizzle",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server, err := helpers.NewDubizzleMockServer()
				require.NoError(t, err)
				return server, ""
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return dubizzle.NewDubizzleScraperFor(cfg, "Dubai", "apartmentflat")
			},
		},
		{
			name: "dewa",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server := helpers.NewMockServer()
				err := server.AddFixture("/en/consumer/billing/slab-tariff", "dewa", "rates_table.html")
				require.NoError(t, err)
				return server, "/en/consumer/billing/slab-tariff"
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return dewa.NewDEWAScraper(cfg)
			},
		},
		{
			name: "sewa",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server := helpers.NewMockServer()
				err := server.AddFixture("/en/content/tariff", "sewa", "tariff_page.html")
				require.NoError(t, err)
				return server, "/en/content/tariff"
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return sewa.NewSEWAScraper(cfg)
			},
		},
		{
			name: "rta",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server := helpers.NewMockServer()
				err := server.AddFixture("/wps/portal/rta/ae/public-transport/fares", "rta", "fare_calculator.html")
				require.NoError(t, err)
				return server, "/wps/portal/rta/ae/public-transport/fares"
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return rta.NewRTAScraper(cfg)
			},
		},
		{
			name: "aadc",
			setup: func(t *testing.T) (*helpers.MockServer, string) {
				server := helpers.NewMockServer()
				err := server.AddFixture("/en/pages/maintarrif.aspx", "aadc", "rates.html")
				require.NoError(t, err)
				return server, "/en/pages/maintarrif.aspx"
			},
			build: func(cfg scrapers.Config) scrapers.Scraper {
				return aadc.NewAADCScraper(cfg)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server, basePath := tc.setup(t)
			defer server.Close()

			cfg := baseConfig
			cfg.BaseURL = strings.TrimRight(server.URL(), "/")
			if basePath != "" {
				cfg.BaseURL = cfg.BaseURL + basePath
			}
			cfg.HTTPClient = server.Server.Client()

			repo := mock.NewCostDataPointRepository()
			service := services.NewScraperServiceWithConfig(repo, &services.ScraperServiceConfig{
				EnableValidation:   false,
				ValidateBeforeSave: false,
			})
			scraper := tc.build(cfg)
			service.RegisterScraper(scraper)

			result, err := service.RunScraper(ctx, scraper.Name())
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Greater(t, result.Fetched, 0, "should fetch items")
			assert.Greater(t, result.Saved, 0, "should persist items")
			assert.Empty(t, result.Errors, "should not record errors")
			assert.Greater(t, repo.GetCallCount("Create"), 0, "repository should receive creations")
		})
	}
}
