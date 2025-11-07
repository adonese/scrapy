package careem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
)

func init() {
	// Initialize logger for tests
	logger.Init()
}

// MockRateSource is a mock implementation of RateSource for testing
type MockRateSource struct {
	name       string
	rates      *CareemRates
	err        error
	confidence float32
	available  bool
}

func (m *MockRateSource) Name() string {
	return m.name
}

func (m *MockRateSource) FetchRates(ctx context.Context) (*CareemRates, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rates, nil
}

func (m *MockRateSource) Confidence() float32 {
	return m.confidence
}

func (m *MockRateSource) IsAvailable(ctx context.Context) bool {
	return m.available
}

func createTestRates() *CareemRates {
	return &CareemRates{
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
		Source:                  "test",
		LastUpdated:             time.Now(),
		Confidence:              0.85,
	}
}

func TestNewCareemScraper(t *testing.T) {
	config := scrapers.Config{
		UserAgent:  "test-agent",
		RateLimit:  1,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := NewCareemScraper(config)

	if scraper == nil {
		t.Fatal("NewCareemScraper() returned nil")
	}

	if scraper.Name() != "careem" {
		t.Errorf("Name() = %v, want careem", scraper.Name())
	}

	if !scraper.CanScrape() {
		t.Error("CanScrape() = false, want true")
	}
}

func TestNewCareemScraperWithSources(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      createTestRates(),
		confidence: 0.9,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})

	if scraper == nil {
		t.Fatal("NewCareemScraperWithSources() returned nil")
	}

	if scraper.aggregator == nil {
		t.Error("aggregator is nil")
	}
}

func TestCareemScraper_Scrape(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      createTestRates(),
		confidence: 0.85,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape() error = %v", err)
	}

	if len(dataPoints) == 0 {
		t.Error("Scrape() returned no data points")
	}

	// Verify data point structure
	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Category = %v, want Transportation", dp.Category)
		}
		if dp.SubCategory != "Ride Sharing" {
			t.Errorf("SubCategory = %v, want Ride Sharing", dp.SubCategory)
		}
		if dp.Unit != "AED" {
			t.Errorf("Unit = %v, want AED", dp.Unit)
		}
		if dp.Price <= 0 {
			t.Errorf("Price = %v, want > 0", dp.Price)
		}
	}
}

func TestCareemScraper_Scrape_NoSourceAvailable(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:      "mock",
		available: false, // Not available
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	_, err := scraper.Scrape(ctx)
	if err == nil {
		t.Error("Scrape() expected error when no source available, got nil")
	}
}

func TestCareemScraper_Scrape_InvalidRates(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	invalidRates := &CareemRates{
		BaseFare:    0, // Invalid: must be > 0
		PerKm:       1.97,
		MinimumFare: 12.0,
		Confidence:  0.8,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      invalidRates,
		confidence: 0.8,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	_, err := scraper.Scrape(ctx)
	if err == nil {
		t.Error("Scrape() expected error for invalid rates, got nil")
	}
}

func TestCareemScraper_CheckRateChanges(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	initialRates := createTestRates()
	initialRates.BaseFare = 8.0

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      initialRates,
		confidence: 0.85,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	// First scrape
	_, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("First scrape error = %v", err)
	}

	// Update rates with significant change
	newRates := createTestRates()
	newRates.BaseFare = 10.0 // +25% change
	mockSource.rates = newRates

	// Second scrape - should detect change
	_, err = scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Second scrape error = %v", err)
	}

	// Verify last rates were updated
	lastRates := scraper.GetLastRates()
	if lastRates == nil {
		t.Fatal("GetLastRates() returned nil")
	}
	if lastRates.BaseFare != 10.0 {
		t.Errorf("BaseFare = %v, want 10.0", lastRates.BaseFare)
	}
}

func TestCareemScraper_GetRatesSummary(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      createTestRates(),
		confidence: 0.85,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	// Before scraping
	_, err := scraper.GetRatesSummary()
	if err == nil {
		t.Error("GetRatesSummary() expected error before scraping, got nil")
	}

	// After scraping
	_, err = scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape() error = %v", err)
	}

	summary, err := scraper.GetRatesSummary()
	if err != nil {
		t.Fatalf("GetRatesSummary() error = %v", err)
	}

	if summary == "" {
		t.Error("GetRatesSummary() returned empty string")
	}

	// Check that summary contains key information
	if !contains(summary, "Base Fare") {
		t.Error("Summary doesn't contain 'Base Fare'")
	}
	if !contains(summary, "Dubai") {
		t.Error("Summary doesn't contain 'Dubai'")
	}
}

func TestCareemScraper_EstimateFare(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      createTestRates(),
		confidence: 0.85,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	// Before scraping
	_, err := scraper.EstimateFare(10.0, 5.0, false, false, 0)
	if err == nil {
		t.Error("EstimateFare() expected error before scraping, got nil")
	}

	// After scraping
	_, err = scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape() error = %v", err)
	}

	fare, err := scraper.EstimateFare(10.0, 5.0, false, false, 0)
	if err != nil {
		t.Fatalf("EstimateFare() error = %v", err)
	}

	if fare <= 0 {
		t.Errorf("EstimateFare() = %v, want > 0", fare)
	}

	// Expected: 8 (base) + 19.7 (10km) + 2.5 (5min) = 30.2
	if fare < 25 || fare > 35 {
		t.Errorf("EstimateFare() = %v, expected between 25 and 35", fare)
	}
}

func TestCareemScraper_RefreshRates(t *testing.T) {
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 10,
		Timeout:   30,
	}

	mockSource := &MockRateSource{
		name:       "mock",
		rates:      createTestRates(),
		confidence: 0.85,
		available:  true,
	}

	scraper := NewCareemScraperWithSources(config, []RateSource{mockSource})
	ctx := context.Background()

	err := scraper.RefreshRates(ctx)
	if err != nil {
		t.Fatalf("RefreshRates() error = %v", err)
	}

	lastRates := scraper.GetLastRates()
	if lastRates == nil {
		t.Fatal("GetLastRates() returned nil after refresh")
	}
}

func TestStaticSource(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_rates.json")

	testData := `{
		"service_type": "careem_go",
		"emirate": "Dubai",
		"base_fare": 8.0,
		"per_km": 1.97,
		"per_minute_wait": 0.5,
		"minimum_fare": 12.0,
		"peak_surcharge_multiplier": 1.5,
		"airport_surcharge": 20.0,
		"salik_toll": 5.0,
		"effective_date": "2025-01-01",
		"source": "test"
	}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	source := NewStaticSource(testFile)
	ctx := context.Background()

	if !source.IsAvailable(ctx) {
		t.Error("IsAvailable() = false, want true")
	}

	rates, err := source.FetchRates(ctx)
	if err != nil {
		t.Fatalf("FetchRates() error = %v", err)
	}

	if rates.BaseFare != 8.0 {
		t.Errorf("BaseFare = %v, want 8.0", rates.BaseFare)
	}

	if rates.Source != "static_file" {
		t.Errorf("Source = %v, want static_file", rates.Source)
	}
}

func TestStaticSource_FileNotFound(t *testing.T) {
	source := NewStaticSource("/nonexistent/file.json")
	ctx := context.Background()

	if source.IsAvailable(ctx) {
		t.Error("IsAvailable() = true for nonexistent file, want false")
	}

	_, err := source.FetchRates(ctx)
	if err == nil {
		t.Error("FetchRates() expected error for nonexistent file, got nil")
	}
}

func TestSourceAggregator(t *testing.T) {
	ctx := context.Background()

	// Create sources with different confidence levels
	lowConfidence := &MockRateSource{
		name:       "low",
		rates:      createTestRates(),
		confidence: 0.5,
		available:  true,
	}
	lowConfidence.rates.Confidence = 0.5

	highConfidence := &MockRateSource{
		name:       "high",
		rates:      createTestRates(),
		confidence: 0.9,
		available:  true,
	}
	highConfidence.rates.Confidence = 0.9

	aggregator := NewSourceAggregator([]RateSource{lowConfidence, highConfidence})

	rates, err := aggregator.FetchBestRates(ctx)
	if err != nil {
		t.Fatalf("FetchBestRates() error = %v", err)
	}

	// Should return high confidence source
	if rates.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9 (highest)", rates.Confidence)
	}
}

func TestSourceAggregator_NoSourcesAvailable(t *testing.T) {
	ctx := context.Background()

	unavailableSource := &MockRateSource{
		name:      "unavailable",
		available: false,
	}

	aggregator := NewSourceAggregator([]RateSource{unavailableSource})

	_, err := aggregator.FetchBestRates(ctx)
	if err == nil {
		t.Error("FetchBestRates() expected error when no sources available, got nil")
	}
}

func TestGetDefaultStaticSourcePath(t *testing.T) {
	path := GetDefaultStaticSourcePath()
	if path == "" {
		t.Error("GetDefaultStaticSourcePath() returned empty string")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) > len(substr) && s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
