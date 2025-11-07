package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/aadc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAADCIntegration_WithFixture(t *testing.T) {
	// Load fixture HTML
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err, "Failed to load AADC fixture")

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent:  "test-agent",
		RateLimit:  10,
		Timeout:    30,
		MaxRetries: 3,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	// Verify results
	require.NoError(t, err)
	require.NotEmpty(t, dataPoints)

	t.Logf("Scraped %d data points from AADC fixture", len(dataPoints))

	// Should have 10 electricity rates + 2 water rates = 12 total
	assert.GreaterOrEqual(t, len(dataPoints), 10)

	// Verify all data points have required fields
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
}

func TestAADCIntegration_ElectricityRates(t *testing.T) {
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)

	// Filter electricity rates
	electricityRates := filterBySubCategory(dataPoints, "Electricity")
	assert.NotEmpty(t, electricityRates)

	t.Logf("Found %d electricity rates", len(electricityRates))

	// Check for national rates
	nationalRates := filterByCustomerType(electricityRates, "national")
	assert.NotEmpty(t, nationalRates, "Should have national electricity rates")

	// Check for expatriate rates
	expatRates := filterByCustomerType(electricityRates, "expatriate")
	assert.NotEmpty(t, expatRates, "Should have expatriate electricity rates")

	// Verify structure of electricity rates
	for _, dp := range electricityRates {
		assert.Equal(t, "Utilities", dp.Category)
		assert.Equal(t, "Electricity", dp.SubCategory)
		assert.Contains(t, dp.ItemName, "AADC Electricity")
		assert.Greater(t, dp.Price, 0.0)
		assert.Less(t, dp.Price, 1.0) // All electricity rates should be < 1 AED per kWh
		assert.Equal(t, "AED per kWh", dp.Unit)
		assert.Contains(t, dp.Tags, "electricity")

		// Check attributes
		assert.Contains(t, dp.Attributes, "customer_type")
		assert.Contains(t, dp.Attributes, "rate_type")
		assert.Contains(t, dp.Attributes, "fils_rate")
		assert.Contains(t, dp.Attributes, "tier_min_kwh")
		assert.Contains(t, dp.Attributes, "tier_max_kwh")
	}
}

func TestAADCIntegration_WaterRates(t *testing.T) {
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)

	// Filter water rates
	waterRates := filterBySubCategory(dataPoints, "Water")
	assert.NotEmpty(t, waterRates)

	t.Logf("Found %d water rates", len(waterRates))

	// Should have 2 water rates (national + expatriate)
	assert.Equal(t, 2, len(waterRates))

	// Check for both customer types
	nationalRates := filterByCustomerType(waterRates, "national")
	assert.Len(t, nationalRates, 1, "Should have 1 national water rate")

	expatRates := filterByCustomerType(waterRates, "expatriate")
	assert.Len(t, expatRates, 1, "Should have 1 expatriate water rate")

	// Verify structure of water rates
	for _, dp := range waterRates {
		assert.Equal(t, "Utilities", dp.Category)
		assert.Equal(t, "Water", dp.SubCategory)
		assert.Contains(t, dp.ItemName, "AADC Water")
		assert.Greater(t, dp.Price, 0.0)
		assert.Equal(t, "AED per 1000 IG", dp.Unit)
		assert.Contains(t, dp.Tags, "water")

		// Check attributes
		assert.Contains(t, dp.Attributes, "customer_type")
		assert.Contains(t, dp.Attributes, "rate_type")
		assert.Equal(t, "flat", dp.Attributes["rate_type"])
	}

	// Verify national rate is lower than expatriate rate
	nationalPrice := nationalRates[0].Price
	expatPrice := expatRates[0].Price
	assert.Less(t, nationalPrice, expatPrice, "National water rate should be lower than expatriate rate")
}

func TestAADCIntegration_CustomerTypeDifferentiation(t *testing.T) {
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)

	// Count customer types
	nationalCount := 0
	expatCount := 0

	for _, dp := range dataPoints {
		if customerType, ok := dp.Attributes["customer_type"].(string); ok {
			if customerType == "national" {
				nationalCount++
				assert.Contains(t, dp.ItemName, "National")
			} else if customerType == "expatriate" {
				expatCount++
				assert.Contains(t, dp.ItemName, "Expatriate")
			}
		}
	}

	assert.Greater(t, nationalCount, 0, "Should have national rates")
	assert.Greater(t, expatCount, 0, "Should have expatriate rates")

	t.Logf("National rates: %d, Expatriate rates: %d", nationalCount, expatCount)
}

func TestAADCIntegration_TierValidation(t *testing.T) {
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)

	// Check electricity tiers
	electricityRates := filterBySubCategory(dataPoints, "Electricity")

	for _, dp := range electricityRates {
		tierMin, hasMin := dp.Attributes["tier_min_kwh"]
		tierMax, hasMax := dp.Attributes["tier_max_kwh"]

		assert.True(t, hasMin, "Should have tier_min_kwh")
		assert.True(t, hasMax, "Should have tier_max_kwh")

		// Verify tier min is valid
		if minVal, ok := tierMin.(int); ok {
			assert.GreaterOrEqual(t, minVal, 0)
		}

		// Verify tier max is valid (either int or "unlimited")
		if maxVal, ok := tierMax.(string); ok {
			assert.Equal(t, "unlimited", maxVal)
		} else if maxVal, ok := tierMax.(int); ok {
			assert.Greater(t, maxVal, 0)
		}
	}
}

func TestAADCIntegration_PriceConversion(t *testing.T) {
	fixtureHTML, err := loadAADCFixture()
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureHTML))
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	require.NoError(t, err)

	// Check electricity prices (fils to AED conversion)
	electricityRates := filterBySubCategory(dataPoints, "Electricity")

	for _, dp := range electricityRates {
		// Check that fils rate is stored in attributes
		filsRate, ok := dp.Attributes["fils_rate"].(float64)
		require.True(t, ok, "Should have fils_rate in attributes")

		// Verify conversion: Price (AED) = fils / 100
		expectedPrice := filsRate / 100.0
		assert.InDelta(t, expectedPrice, dp.Price, 0.001, "Price should be correctly converted from fils to AED")
	}
}

func TestAADCIntegration_ErrorHandling_EmptyPage(t *testing.T) {
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
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	assert.Error(t, err)
	assert.Nil(t, dataPoints)
}

func TestAADCIntegration_ErrorHandling_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := aadc.NewAADCScraperWithURL(config, server.URL)

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	assert.Error(t, err)
	assert.Nil(t, dataPoints)
	assert.Contains(t, err.Error(), "bad status")
}

// Helper functions

func loadAADCFixture() (string, error) {
	// Get project root
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Navigate up to project root if needed
	projectRoot := cwd
	for i := 0; i < 3; i++ {
		fixturePath := filepath.Join(projectRoot, "test", "fixtures", "aadc", "rates.html")
		if _, err := os.Stat(fixturePath); err == nil {
			content, err := os.ReadFile(fixturePath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
		projectRoot = filepath.Dir(projectRoot)
	}

	return "", os.ErrNotExist
}

func filterBySubCategory(dataPoints []*models.CostDataPoint, subCategory string) []*models.CostDataPoint {
	filtered := []*models.CostDataPoint{}
	for _, dp := range dataPoints {
		if dp.SubCategory == subCategory {
			filtered = append(filtered, dp)
		}
	}
	return filtered
}

func filterByCustomerType(dataPoints []*models.CostDataPoint, customerType string) []*models.CostDataPoint {
	filtered := []*models.CostDataPoint{}
	for _, dp := range dataPoints {
		if ct, ok := dp.Attributes["customer_type"].(string); ok && ct == customerType {
			filtered = append(filtered, dp)
		}
	}
	return filtered
}
