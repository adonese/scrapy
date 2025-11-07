package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/careem"
)

func TestCareemIntegration_WithStaticSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use the fixture file
	fixturesDir := getFixturesDir(t)
	fixturePath := filepath.Join(fixturesDir, "careem", "rates.json")

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	// Create static source
	staticSource := careem.NewStaticSource(fixturePath)

	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{staticSource})

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(dataPoints) == 0 {
		t.Fatal("Expected data points, got 0")
	}

	t.Logf("Successfully scraped %d data points", len(dataPoints))

	// Verify data points
	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}
		if dp.SubCategory != "Ride Sharing" {
			t.Errorf("Expected subcategory Ride Sharing, got %s", dp.SubCategory)
		}
		if dp.Unit != "AED" {
			t.Errorf("Expected unit AED, got %s", dp.Unit)
		}
		if dp.Price <= 0 {
			t.Errorf("Expected positive price, got %f", dp.Price)
		}
		if dp.Location.Emirate == "" {
			t.Error("Expected emirate to be set")
		}
	}
}

func TestCareemIntegration_WithMockHTTPSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock HTTP server
	ratesData := careem.CareemRates{
		ServiceType:             "careem_go",
		Emirate:                 "Dubai",
		BaseFare:                8.0,
		PerKm:                   1.97,
		PerMinuteWait:           0.5,
		MinimumFare:             12.0,
		PeakSurchargeMultiplier: 1.5,
		AirportSurcharge:        20.0,
		SalikToll:               5.0,
		EffectiveDate:           "2025-01-01",
		Source:                  "mock_server",
		Confidence:              0.8,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ratesData)
	}))
	defer server.Close()

	// Create custom mock source
	mockSource := &MockHTTPSource{
		url:        server.URL,
		rates:      &ratesData,
		confidence: 0.8,
	}

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{mockSource})

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(dataPoints) == 0 {
		t.Fatal("Expected data points, got 0")
	}

	t.Logf("Successfully scraped %d data points from mock server", len(dataPoints))
}

func TestCareemIntegration_MultipleSourceFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create sources: failing source + working static source
	fixturesDir := getFixturesDir(t)
	fixturePath := filepath.Join(fixturesDir, "careem", "rates.json")

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	// First source that will fail
	failingSource := &MockFailingSource{}

	// Second source that will work
	staticSource := careem.NewStaticSource(fixturePath)

	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{
		failingSource,
		staticSource,
	})

	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)

	if err != nil {
		t.Fatalf("Scrape failed even with fallback: %v", err)
	}

	if len(dataPoints) == 0 {
		t.Fatal("Expected data points from fallback source, got 0")
	}

	t.Logf("Successfully fell back to static source, got %d data points", len(dataPoints))
}

func TestCareemIntegration_RateValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create source with invalid rates
	invalidRates := &careem.CareemRates{
		ServiceType: "careem_go",
		Emirate:     "Dubai",
		BaseFare:    0, // Invalid!
		PerKm:       1.97,
		MinimumFare: 12.0,
		Confidence:  0.8,
	}

	mockSource := &MockHTTPSource{
		rates:      invalidRates,
		confidence: 0.8,
	}

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{mockSource})

	ctx := context.Background()
	_, err := scraper.Scrape(ctx)

	if err == nil {
		t.Fatal("Expected validation error for invalid rates, got nil")
	}

	t.Logf("Correctly rejected invalid rates: %v", err)
}

func TestCareemIntegration_RateChangeDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  10, // Higher rate limit for multiple scrapes
		Timeout:    30,
		MaxRetries: 3,
	}

	// Initial rates
	initialRates := &careem.CareemRates{
		ServiceType:             "careem_go",
		Emirate:                 "Dubai",
		BaseFare:                8.0,
		PerKm:                   1.97,
		PerMinuteWait:           0.5,
		MinimumFare:             12.0,
		PeakSurchargeMultiplier: 1.5,
		Confidence:              0.8,
	}

	mockSource := &MockHTTPSource{
		rates:      initialRates,
		confidence: 0.8,
	}

	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{mockSource})

	ctx := context.Background()

	// First scrape
	_, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("First scrape failed: %v", err)
	}

	// Update rates significantly
	updatedRates := &careem.CareemRates{
		ServiceType:             "careem_go",
		Emirate:                 "Dubai",
		BaseFare:                10.0, // +25%
		PerKm:                   2.5,  // +27%
		PerMinuteWait:           0.5,
		MinimumFare:             15.0, // +25%
		PeakSurchargeMultiplier: 1.5,
		Confidence:              0.8,
	}

	mockSource.rates = updatedRates

	// Second scrape - should detect changes
	_, err = scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Second scrape failed: %v", err)
	}

	// Verify rates were updated
	lastRates := scraper.GetLastRates()
	if lastRates == nil {
		t.Fatal("Expected last rates to be set")
	}

	if lastRates.BaseFare != 10.0 {
		t.Errorf("Expected BaseFare=10.0, got %f", lastRates.BaseFare)
	}

	t.Log("Successfully detected rate changes")
}

func TestCareemIntegration_FareEstimation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fixturesDir := getFixturesDir(t)
	fixturePath := filepath.Join(fixturesDir, "careem", "rates.json")

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	staticSource := careem.NewStaticSource(fixturePath)
	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{staticSource})

	ctx := context.Background()

	// Scrape to load rates
	_, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Test fare estimation
	tests := []struct {
		name            string
		distanceKm      float64
		waitTimeMinutes float64
		isPeakHour      bool
		isAirport       bool
		salikGates      int
		minExpected     float64
		maxExpected     float64
	}{
		{
			name:            "short trip",
			distanceKm:      2.0,
			waitTimeMinutes: 0,
			isPeakHour:      false,
			isAirport:       false,
			salikGates:      0,
			minExpected:     12.0, // minimum fare
			maxExpected:     15.0,
		},
		{
			name:            "normal trip",
			distanceKm:      10.0,
			waitTimeMinutes: 5.0,
			isPeakHour:      false,
			isAirport:       false,
			salikGates:      1,
			minExpected:     30.0,
			maxExpected:     40.0,
		},
		{
			name:            "airport trip with peak",
			distanceKm:      15.0,
			waitTimeMinutes: 10.0,
			isPeakHour:      true,
			isAirport:       true,
			salikGates:      2,
			minExpected:     70.0,
			maxExpected:     90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fare, err := scraper.EstimateFare(
				tt.distanceKm,
				tt.waitTimeMinutes,
				tt.isPeakHour,
				tt.isAirport,
				tt.salikGates,
			)

			if err != nil {
				t.Fatalf("EstimateFare failed: %v", err)
			}

			if fare < tt.minExpected || fare > tt.maxExpected {
				t.Errorf("Fare=%f, expected between %f and %f",
					fare, tt.minExpected, tt.maxExpected)
			}

			t.Logf("Estimated fare: %.2f AED", fare)
		})
	}
}

func TestCareemIntegration_GetRatesSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fixturesDir := getFixturesDir(t)
	fixturePath := filepath.Join(fixturesDir, "careem", "rates.json")

	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (compatible; CostOfLivingBot/1.0)",
		RateLimit:  2,
		Timeout:    30,
		MaxRetries: 3,
	}

	staticSource := careem.NewStaticSource(fixturePath)
	scraper := careem.NewCareemScraperWithSources(config, []careem.RateSource{staticSource})

	ctx := context.Background()

	// Scrape to load rates
	_, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Get summary
	summary, err := scraper.GetRatesSummary()
	if err != nil {
		t.Fatalf("GetRatesSummary failed: %v", err)
	}

	if summary == "" {
		t.Fatal("Expected non-empty summary")
	}

	// Verify summary contains key information
	requiredStrings := []string{
		"Base Fare",
		"Per Kilometer",
		"Minimum Fare",
		"Dubai",
		"AED",
	}

	for _, str := range requiredStrings {
		if !contains(summary, str) {
			t.Errorf("Summary missing required string: %s", str)
		}
	}

	t.Log("Successfully generated rates summary")
	t.Logf("\n%s", summary)
}

// Mock sources for testing

type MockHTTPSource struct {
	url        string
	rates      *careem.CareemRates
	confidence float32
}

func (m *MockHTTPSource) Name() string {
	return "mock_http"
}

func (m *MockHTTPSource) FetchRates(ctx context.Context) (*careem.CareemRates, error) {
	m.rates.Source = m.Name()
	m.rates.Confidence = m.confidence
	return m.rates, nil
}

func (m *MockHTTPSource) Confidence() float32 {
	return m.confidence
}

func (m *MockHTTPSource) IsAvailable(ctx context.Context) bool {
	return true
}

type MockFailingSource struct{}

func (m *MockFailingSource) Name() string {
	return "mock_failing"
}

func (m *MockFailingSource) FetchRates(ctx context.Context) (*careem.CareemRates, error) {
	return nil, http.ErrServerClosed
}

func (m *MockFailingSource) Confidence() float32 {
	return 0.9
}

func (m *MockFailingSource) IsAvailable(ctx context.Context) bool {
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
