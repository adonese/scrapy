package dewa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Initialize logger for tests
	logger.Init()
}

func TestDEWAScraper_Integration_WithFixture(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "test", "fixtures", "dewa", "rates_table.html")
	fixtureData, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file")

	// Create mock server with fixture data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureData)
	}))
	defer server.Close()

	// Create scraper with test config
	config := scrapers.Config{
		UserAgent: "test-agent/1.0",
		RateLimit: 10,
		Timeout:   30,
	}
	scraper := NewDEWAScraper(config)

	// Fetch and parse the fixture
	resp, err := scraper.client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)

	// Extract rates using the parser
	dataPoints, err := scraper.extractRates(doc, server.URL)
	require.NoError(t, err)
	require.NotEmpty(t, dataPoints)

	// Should have 4 electricity slabs + 3 water slabs + 1 fuel surcharge = 8 data points
	assert.GreaterOrEqual(t, len(dataPoints), 7, "Should have at least 7 data points")

	// Validate electricity data points
	electricityPoints := filterBySubCategory(dataPoints, "Electricity")
	assert.Len(t, electricityPoints, 4, "Should have 4 electricity slabs")

	// Check first electricity slab (0-2000 kWh @ 23.0 fils)
	firstElectric := findByItemName(dataPoints, "DEWA Electricity Slab 0-2000 kWh")
	require.NotNil(t, firstElectric, "First electricity slab not found")
	assert.Equal(t, "Utilities", firstElectric.Category)
	assert.Equal(t, "Electricity", firstElectric.SubCategory)
	assert.Equal(t, 0.23, firstElectric.Price, "23 fils should convert to 0.23 AED")
	assert.Equal(t, "Dubai", firstElectric.Location.Emirate)
	assert.Equal(t, "dewa_official", firstElectric.Source)
	assert.Equal(t, float32(0.98), firstElectric.Confidence)
	assert.Equal(t, "AED", firstElectric.Unit)
	assert.Contains(t, firstElectric.Tags, "electricity")
	assert.Equal(t, "fils_per_kwh", firstElectric.Attributes["unit"])

	// Check consumption ranges (they are stored as int)
	minRange, ok := firstElectric.Attributes["consumption_range_min"]
	require.True(t, ok, "consumption_range_min should exist")
	assert.Equal(t, 0, minRange)

	maxRange, ok := firstElectric.Attributes["consumption_range_max"]
	require.True(t, ok, "consumption_range_max should exist")
	assert.Equal(t, 2000, maxRange)

	// Check last electricity slab (6000+ kWh @ 38.0 fils)
	lastElectric := findByItemName(dataPoints, "DEWA Electricity Slab 6001+ kWh")
	require.NotNil(t, lastElectric, "Last electricity slab not found")
	assert.Equal(t, 0.38, lastElectric.Price, "38 fils should convert to 0.38 AED")

	lastMinRange, ok := lastElectric.Attributes["consumption_range_min"]
	require.True(t, ok, "consumption_range_min should exist")
	assert.Equal(t, 6001, lastMinRange)

	_, hasMax := lastElectric.Attributes["consumption_range_max"]
	assert.False(t, hasMax, "Last slab should not have max range")

	// Validate water data points
	waterPoints := filterBySubCategory(dataPoints, "Water")
	assert.Len(t, waterPoints, 3, "Should have 3 water slabs")

	// Check first water slab (0-5000 IG @ 3.57 fils)
	firstWater := findByItemName(dataPoints, "DEWA Water Slab 0-5000 IG")
	require.NotNil(t, firstWater, "First water slab not found")
	assert.Equal(t, "Utilities", firstWater.Category)
	assert.Equal(t, "Water", firstWater.SubCategory)
	assert.InDelta(t, 0.0357, firstWater.Price, 0.0001, "3.57 fils should convert to 0.0357 AED")
	assert.Contains(t, firstWater.Tags, "water")
	assert.Equal(t, "fils_per_ig", firstWater.Attributes["unit"])

	waterMinRange, ok := firstWater.Attributes["consumption_range_min"]
	require.True(t, ok, "consumption_range_min should exist")
	assert.Equal(t, 0, waterMinRange)

	waterMaxRange, ok := firstWater.Attributes["consumption_range_max"]
	require.True(t, ok, "consumption_range_max should exist")
	assert.Equal(t, 5000, waterMaxRange)

	// Check fuel surcharge
	fuelSurcharge := findByItemName(dataPoints, "DEWA Fuel Surcharge")
	require.NotNil(t, fuelSurcharge, "Fuel surcharge not found")
	assert.Equal(t, "Utilities", fuelSurcharge.Category)
	assert.Equal(t, "Fuel Surcharge", fuelSurcharge.SubCategory)
	assert.Equal(t, 0.065, fuelSurcharge.Price, "6.5 fils should convert to 0.065 AED")

	// Verify all data points have required fields
	for _, dp := range dataPoints {
		assert.NotEmpty(t, dp.Category, "Category should not be empty")
		assert.NotEmpty(t, dp.SubCategory, "SubCategory should not be empty")
		assert.NotEmpty(t, dp.ItemName, "ItemName should not be empty")
		assert.Greater(t, dp.Price, 0.0, "Price should be positive")
		assert.Equal(t, "Dubai", dp.Location.Emirate)
		assert.Equal(t, "dewa_official", dp.Source)
		assert.NotEmpty(t, dp.SourceURL)
		assert.Greater(t, dp.Confidence, float32(0.9), "Confidence should be high for official source")
		assert.Equal(t, "AED", dp.Unit)
		assert.NotEmpty(t, dp.Tags)
	}

	// Test context cancellation
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	err = scraper.rateLimiter.Wait(cancelCtx)
	assert.Error(t, err, "Should error on cancelled context")
}

func TestDEWAScraper_Integration_ParseElectricitySlabs(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "test", "fixtures", "dewa", "rates_table.html")
	fixtureData, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(fixtureData)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)

	slabs, err := parseElectricitySlabs(doc)
	require.NoError(t, err)
	require.Len(t, slabs, 4)

	// Verify all slabs
	expected := []struct {
		min  int
		max  int
		rate float64
	}{
		{0, 2000, 23.0},
		{2001, 4000, 28.0},
		{4001, 6000, 32.0},
		{6001, -1, 38.0},
	}

	for i, exp := range expected {
		assert.Equal(t, exp.min, slabs[i].MinRange, "Slab %d min range mismatch", i)
		assert.Equal(t, exp.max, slabs[i].MaxRange, "Slab %d max range mismatch", i)
		assert.Equal(t, exp.rate, slabs[i].Rate, "Slab %d rate mismatch", i)
		assert.Equal(t, "fils_per_kwh", slabs[i].Unit)
	}
}

func TestDEWAScraper_Integration_ParseWaterSlabs(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "test", "fixtures", "dewa", "rates_table.html")
	fixtureData, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(fixtureData)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)

	slabs, err := parseWaterSlabs(doc)
	require.NoError(t, err)
	require.Len(t, slabs, 3)

	// Verify all slabs
	expected := []struct {
		min  int
		max  int
		rate float64
	}{
		{0, 5000, 3.57},
		{5001, 10000, 5.24},
		{10001, -1, 10.52},
	}

	for i, exp := range expected {
		assert.Equal(t, exp.min, slabs[i].MinRange, "Slab %d min range mismatch", i)
		assert.Equal(t, exp.max, slabs[i].MaxRange, "Slab %d max range mismatch", i)
		assert.Equal(t, exp.rate, slabs[i].Rate, "Slab %d rate mismatch", i)
		assert.Equal(t, "fils_per_ig", slabs[i].Unit)
	}
}

func TestDEWAScraper_Integration_ErrorCases(t *testing.T) {
	t.Run("HTTP 500 Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := scrapers.Config{
			UserAgent: "test-agent",
			RateLimit: 10,
			Timeout:   30,
		}
		scraper := NewDEWAScraper(config)

		resp, err := scraper.client.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Malformed HTML", func(t *testing.T) {
		malformedHTML := `<html><body><h1>Invalid`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(malformedHTML))
		}))
		defer server.Close()

		config := scrapers.Config{
			UserAgent: "test-agent",
			RateLimit: 10,
			Timeout:   30,
		}
		scraper := NewDEWAScraper(config)

		resp, err := scraper.client.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		require.NoError(t, err) // goquery is forgiving

		// Try to extract - should get an error
		_, err = scraper.extractRates(doc, server.URL)
		assert.Error(t, err)
	})

	t.Run("Empty Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		}))
		defer server.Close()

		config := scrapers.Config{
			UserAgent: "test-agent",
			RateLimit: 10,
			Timeout:   30,
		}
		scraper := NewDEWAScraper(config)

		resp, err := scraper.client.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		require.NoError(t, err)

		_, err = scraper.extractRates(doc, server.URL)
		assert.Error(t, err)
	})
}

// Helper functions

func filterBySubCategory(points []*models.CostDataPoint, subCategory string) []*models.CostDataPoint {
	var result []*models.CostDataPoint
	for _, p := range points {
		if p.SubCategory == subCategory {
			result = append(result, p)
		}
	}
	return result
}

func findByItemName(points []*models.CostDataPoint, itemName string) *models.CostDataPoint {
	for _, p := range points {
		if p.ItemName == itemName {
			return p
		}
	}
	return nil
}
