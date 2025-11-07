package aadc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Initialize logger for tests
	logger.Init()
}

func TestNewAADCScraper(t *testing.T) {
	config := scrapers.Config{
		UserAgent:  "test-agent",
		RateLimit:  1,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := NewAADCScraper(config)
	require.NotNil(t, scraper)
	assert.Equal(t, config, scraper.config)
	assert.Equal(t, DefaultAADCURL, scraper.url)
	assert.NotNil(t, scraper.client)
	assert.NotNil(t, scraper.rateLimiter)
}

func TestNewAADCScraperWithURL(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}
	customURL := "http://example.com/test"

	scraper := NewAADCScraperWithURL(config, customURL)
	require.NotNil(t, scraper)
	assert.Equal(t, customURL, scraper.url)
}

func TestAADCScraper_Name(t *testing.T) {
	config := scrapers.Config{RateLimit: 1, Timeout: 30}
	scraper := NewAADCScraper(config)
	assert.Equal(t, "aadc", scraper.Name())
}

func TestAADCScraper_CanScrape(t *testing.T) {
	config := scrapers.Config{
		RateLimit: 100, // 100 requests per second (very permissive for testing)
		Timeout:   30,
	}
	scraper := NewAADCScraper(config)

	// First call should succeed
	assert.True(t, scraper.CanScrape())

	// Immediate subsequent calls should also succeed with high rate limit
	for i := 0; i < 5; i++ {
		result := scraper.CanScrape()
		if !result {
			t.Logf("Rate limit hit on iteration %d", i)
		}
		// We can't assert true here as CanScrape returns false if rate limited
		// Just verify the method doesn't panic
	}
}

func TestAADCScraper_Scrape_Success(t *testing.T) {
	// Create a test server with valid AADC HTML
	html := getValidAADCHTML()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)
	require.NotEmpty(t, dataPoints)

	// Should have both electricity and water rates
	// Expected: ~10 electricity rates (2 national + 8 expatriate) + 2 water rates
	assert.GreaterOrEqual(t, len(dataPoints), 4)

	// Verify structure of data points
	for _, dp := range dataPoints {
		assert.Equal(t, "Utilities", dp.Category)
		assert.Contains(t, []string{"Electricity", "Water"}, dp.SubCategory)
		assert.NotEmpty(t, dp.ItemName)
		assert.Greater(t, dp.Price, 0.0)
		assert.Equal(t, "Abu Dhabi", dp.Location.Emirate)
		assert.Equal(t, "aadc_official", dp.Source)
		assert.Equal(t, server.URL, dp.SourceURL)
		assert.Equal(t, float32(0.98), dp.Confidence)
		assert.NotEmpty(t, dp.Tags)
		assert.NotEmpty(t, dp.Attributes)
	}

	// Check for specific rate types
	hasElectricity := false
	hasWater := false
	hasNational := false
	hasExpatriate := false

	for _, dp := range dataPoints {
		if dp.SubCategory == "Electricity" {
			hasElectricity = true
		}
		if dp.SubCategory == "Water" {
			hasWater = true
		}
		if customerType, ok := dp.Attributes["customer_type"].(string); ok {
			if customerType == "national" {
				hasNational = true
			}
			if customerType == "expatriate" {
				hasExpatriate = true
			}
		}
	}

	assert.True(t, hasElectricity, "Should have electricity rates")
	assert.True(t, hasWater, "Should have water rates")
	assert.True(t, hasNational, "Should have national rates")
	assert.True(t, hasExpatriate, "Should have expatriate rates")
}

func TestAADCScraper_Scrape_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	assert.Error(t, err)
	assert.Nil(t, dataPoints)
	assert.Contains(t, err.Error(), "bad status")
}

func TestAADCScraper_Scrape_InvalidHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>No rates here</body></html>"))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	assert.Error(t, err)
	assert.Nil(t, dataPoints)
	assert.Contains(t, err.Error(), "parse rates")
}

func TestAADCScraper_Scrape_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getValidAADCHTML()))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewAADCScraperWithURL(config, server.URL)

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dataPoints, err := scraper.Scrape(ctx)

	assert.Error(t, err)
	assert.Nil(t, dataPoints)
}

func TestAADCScraper_Scrape_RateLimiting(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getValidAADCHTML()))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1, // 1 request per second
		Timeout:   30,
	}
	scraper := NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()

	// First scrape should succeed
	start := time.Now()
	_, err := scraper.Scrape(ctx)
	require.NoError(t, err)

	// Second scrape should be rate limited (delayed)
	_, err = scraper.Scrape(ctx)
	require.NoError(t, err)
	elapsed := time.Since(start)

	// Should take at least ~1 second due to rate limiting
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond)
	assert.Equal(t, 2, requestCount)
}

func TestAADCScraper_ParseRates(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewAADCScraper(config)

	html := getValidAADCHTML()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	scraper.url = server.URL

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)
	require.NotEmpty(t, dataPoints)

	// Verify all rates have required fields
	for _, dp := range dataPoints {
		assert.NotEmpty(t, dp.Category)
		assert.NotEmpty(t, dp.SubCategory)
		assert.NotEmpty(t, dp.ItemName)
		assert.Greater(t, dp.Price, 0.0)
		assert.NotEmpty(t, dp.Location.Emirate)
		assert.NotEmpty(t, dp.Source)
		assert.NotEmpty(t, dp.Unit)
		assert.NotZero(t, dp.RecordedAt)
		assert.NotZero(t, dp.ValidFrom)
	}
}

// getValidAADCHTML returns a valid AADC HTML fixture for testing
func getValidAADCHTML() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AADC Tariff Information</title>
</head>
<body>
    <div class="rates-container">
        <h1>AADC Tariff Information</h1>

        <section class="electricity-residential">
            <h2>Residential Electricity Tariff</h2>

            <h3>UAE Nationals</h3>
            <table class="tariff-table">
                <thead>
                    <tr>
                        <th>Monthly Consumption (kWh)</th>
                        <th>Rate (Fils/kWh)</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Up to 30,000</td>
                        <td>5.8 fils</td>
                    </tr>
                    <tr>
                        <td>Above 30,000</td>
                        <td>6.7 fils</td>
                    </tr>
                </tbody>
            </table>

            <h3>Expatriates</h3>
            <table class="tariff-table">
                <thead>
                    <tr>
                        <th>Monthly Consumption (kWh)</th>
                        <th>Rate (Fils/kWh)</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Up to 400</td>
                        <td>6.7 fils</td>
                    </tr>
                    <tr>
                        <td>401 - 700</td>
                        <td>7.6 fils</td>
                    </tr>
                    <tr>
                        <td>701 - 1,000</td>
                        <td>9.5 fils</td>
                    </tr>
                    <tr>
                        <td>1,001 - 2,000</td>
                        <td>11.5 fils</td>
                    </tr>
                    <tr>
                        <td>2,001 - 3,000</td>
                        <td>17.2 fils</td>
                    </tr>
                    <tr>
                        <td>3,001 - 4,000</td>
                        <td>20.6 fils</td>
                    </tr>
                    <tr>
                        <td>4,001 - 10,000</td>
                        <td>26.8 fils</td>
                    </tr>
                    <tr>
                        <td>Above 10,000</td>
                        <td>28.7 fils</td>
                    </tr>
                </tbody>
            </table>
        </section>

        <section class="water-residential">
            <h2>Residential Water Tariff</h2>

            <h3>UAE Nationals</h3>
            <table class="tariff-table">
                <tbody>
                    <tr>
                        <td>Rate per 1,000 IG</td>
                        <td>AED 2.09</td>
                    </tr>
                </tbody>
            </table>

            <h3>Expatriates</h3>
            <table class="tariff-table">
                <tbody>
                    <tr>
                        <td>Rate per 1,000 IG</td>
                        <td>AED 8.55</td>
                    </tr>
                </tbody>
            </table>
        </section>
    </div>
</body>
</html>
	`
}
