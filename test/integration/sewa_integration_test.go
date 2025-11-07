package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/sewa"
)

func TestSEWAScraper_Integration(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/sewa/tariff_page.html")
	require.NoError(t, err)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (Test)",
		RateLimit:  10,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := sewa.NewSEWAScraper(config)

	t.Run("scraper metadata", func(t *testing.T) {
		assert.Equal(t, "sewa", scraper.Name())
		assert.True(t, scraper.CanScrape())
	})

	t.Run("scrape from fixture", func(t *testing.T) {
		// Use ScrapeFromHTML helper for testing
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")

		require.NoError(t, err)
		require.NotEmpty(t, dataPoints)

		// Verify expected number of data points
		assert.Len(t, dataPoints, 10, "should extract 10 data points")
	})

	t.Run("electricity tiers extracted correctly", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		electricityPoints := filterBySubCategorySEWA(dataPoints, "Electricity")
		assert.Len(t, electricityPoints, 7, "should have 7 electricity tiers (4 Emirati + 3 Expatriate)")

		emiratiTiers := 0
		expatTiers := 0

		for _, dp := range electricityPoints {
			customerType, ok := dp.Attributes["customer_type"].(string)
			require.True(t, ok, "customer_type attribute should exist")

			if customerType == "emirati" {
				emiratiTiers++
			} else if customerType == "expatriate" {
				expatTiers++
			}

			// Verify common fields
			assert.Equal(t, "Utilities", dp.Category)
			assert.Equal(t, "Electricity", dp.SubCategory)
			assert.Equal(t, "Sharjah", dp.Location.Emirate)
			assert.Equal(t, "sewa_official", dp.Source)
			assert.Greater(t, dp.Price, 0.0)
			assert.Equal(t, float32(0.98), dp.Confidence)
		}

		assert.Equal(t, 4, emiratiTiers, "should have 4 Emirati electricity tiers")
		assert.Equal(t, 3, expatTiers, "should have 3 Expatriate electricity tiers")
	})

	t.Run("water rates extracted correctly", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		waterPoints := filterBySubCategorySEWA(dataPoints, "Water")
		assert.Len(t, waterPoints, 2, "should have 2 water rates")

		var emiratiWater, expatWater *models.CostDataPoint
		for _, dp := range waterPoints {
			customerType := dp.Attributes["customer_type"].(string)
			if customerType == "emirati" {
				emiratiWater = dp
			} else {
				expatWater = dp
			}
		}

		require.NotNil(t, emiratiWater, "should have Emirati water rate")
		require.NotNil(t, expatWater, "should have Expatriate water rate")

		// Verify Emirati water rate
		assert.Equal(t, 8.0, emiratiWater.Price)
		assert.Equal(t, "AED per 1000 gallons", emiratiWater.Unit)
		assert.Contains(t, emiratiWater.ItemName, "Emirati")

		// Verify Expatriate water rate
		assert.Equal(t, 15.0, expatWater.Price)
		assert.Equal(t, "AED per 1000 gallons", expatWater.Unit)
		assert.Contains(t, expatWater.ItemName, "Expatriate")
	})

	t.Run("sewerage charge extracted correctly", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		seweragePoints := filterBySubCategorySEWA(dataPoints, "Sewerage")
		require.Len(t, seweragePoints, 1, "should have 1 sewerage charge")

		sewerage := seweragePoints[0]
		assert.Equal(t, "Utilities", sewerage.Category)
		assert.Equal(t, "SEWA Sewerage Charge", sewerage.ItemName)
		assert.Equal(t, 0.50, sewerage.Price, "50% stored as decimal")
		assert.Equal(t, "percentage of water charge", sewerage.Unit)
		assert.Contains(t, sewerage.Attributes, "calculation_method")
	})

	t.Run("price conversions", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		electricityPoints := filterBySubCategorySEWA(dataPoints, "Electricity")

		// Find the 14 fils tier and verify conversion
		for _, dp := range electricityPoints {
			rateFils, ok := dp.Attributes["rate_fils"].(float64)
			if ok && rateFils == 14.0 {
				assert.Equal(t, 0.14, dp.Price, "14 fils should convert to 0.14 AED")
				break
			}
		}

		// Find the 27.5 fils tier and verify decimal conversion
		for _, dp := range electricityPoints {
			rateFils, ok := dp.Attributes["rate_fils"].(float64)
			if ok && rateFils == 27.5 {
				assert.Equal(t, 0.275, dp.Price, "27.5 fils should convert to 0.275 AED")
				break
			}
		}
	})

	t.Run("customer type differentiation", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		// Verify that Emirati and Expatriate have different rates
		electricityPoints := filterBySubCategorySEWA(dataPoints, "Electricity")

		var emiratiFirstTier, expatFirstTier float64
		for _, dp := range electricityPoints {
			customerType := dp.Attributes["customer_type"].(string)
			minConsumption, ok := dp.Attributes["consumption_range_min"].(int)
			if !ok || minConsumption != 1 {
				continue
			}

			if customerType == "emirati" {
				emiratiFirstTier = dp.Price
			} else if customerType == "expatriate" {
				expatFirstTier = dp.Price
			}
		}

		assert.Greater(t, expatFirstTier, emiratiFirstTier, "Expatriate rates should be higher than Emirati rates")
	})

	t.Run("all required attributes present", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(fixtureHTML)))
		require.NoError(t, err)

		dataPoints, err := scraper.ScrapeFromHTML(doc, "https://www.sewa.gov.ae/en/content/tariff")
		require.NoError(t, err)

		for i, dp := range dataPoints {
			// Common required fields
			assert.NotEmpty(t, dp.Category, "data point %d: category should not be empty", i)
			assert.NotEmpty(t, dp.SubCategory, "data point %d: subcategory should not be empty", i)
			assert.NotEmpty(t, dp.ItemName, "data point %d: item name should not be empty", i)
			assert.Greater(t, dp.Price, 0.0, "data point %d: price should be positive", i)
			assert.Equal(t, "Sharjah", dp.Location.Emirate, "data point %d: should be in Sharjah", i)
			assert.Equal(t, "sewa_official", dp.Source, "data point %d: source should be sewa_official", i)
			assert.NotEmpty(t, dp.Unit, "data point %d: unit should not be empty", i)
			assert.Greater(t, dp.Confidence, float32(0.0), "data point %d: confidence should be positive", i)
			assert.NotEmpty(t, dp.Tags, "data point %d: tags should not be empty", i)

			// Timestamps
			assert.False(t, dp.RecordedAt.IsZero(), "data point %d: recorded_at should be set", i)
			assert.False(t, dp.ValidFrom.IsZero(), "data point %d: valid_from should be set", i)
		}
	})

	t.Run("error handling - empty page", func(t *testing.T) {
		emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html><html><body></body></html>`))
		}))
		defer emptyServer.Close()

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<!DOCTYPE html><html><body></body></html>`))
		require.NoError(t, err)

		_, err = scraper.ScrapeFromHTML(doc, emptyServer.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tariff data found")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := scraper.Scrape(ctx)
		assert.Error(t, err)
	})
}

// filterBySubCategory is a helper function to filter data points by subcategory
func filterBySubCategorySEWA(dataPoints []*models.CostDataPoint, subCategory string) []*models.CostDataPoint {
	var filtered []*models.CostDataPoint
	for _, dp := range dataPoints {
		if dp.SubCategory == subCategory {
			filtered = append(filtered, dp)
		}
	}
	return filtered
}
