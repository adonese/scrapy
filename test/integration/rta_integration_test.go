package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/rta"
)

func TestRTAScraperIntegration(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent:  "test-agent",
		RateLimit:  1,
		Timeout:    30,
		MaxRetries: 3,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Validate results
	if len(dataPoints) == 0 {
		t.Fatal("Expected data points but got none")
	}

	t.Logf("Extracted %d data points", len(dataPoints))

	// Count by transport mode
	metroCount := 0
	busCount := 0
	tramCount := 0
	taxiCount := 0

	for _, dp := range dataPoints {
		mode, ok := dp.Attributes["transport_mode"]
		if !ok && dp.SubCategory == "Taxi" {
			taxiCount++
			continue
		}

		switch mode {
		case "metro":
			metroCount++
		case "bus":
			busCount++
		case "tram":
			tramCount++
		case "taxi":
			taxiCount++
		}
	}

	t.Logf("Metro: %d, Bus: %d, Tram: %d, Taxi: %d", metroCount, busCount, tramCount, taxiCount)

	// Validate minimum counts
	if metroCount < 5 {
		t.Errorf("Expected at least 5 metro fares, got %d", metroCount)
	}

	if busCount < 3 {
		t.Errorf("Expected at least 3 bus fares, got %d", busCount)
	}

	if tramCount < 2 {
		t.Errorf("Expected at least 2 tram fares, got %d", tramCount)
	}

	if taxiCount < 2 {
		t.Errorf("Expected at least 2 taxi fares, got %d", taxiCount)
	}
}

func TestRTAScraperMetroFares(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Filter metro fares
	metroFares := filterByMode(dataPoints, "metro")

	if len(metroFares) == 0 {
		t.Fatal("Expected metro fares but got none")
	}

	// Validate metro fare structure
	hasSilver := false
	hasGold := false
	hasDayPass := false

	for _, dp := range metroFares {
		// Check category
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}

		if dp.SubCategory != "Public Transport" {
			t.Errorf("Expected subcategory Public Transport, got %s", dp.SubCategory)
		}

		// Check card types
		if cardType, ok := dp.Attributes["card_type"]; ok {
			switch cardType {
			case "silver":
				hasSilver = true
			case "gold":
				hasGold = true
			}
		}

		// Check for day pass
		if fareType, ok := dp.Attributes["fare_type"]; ok {
			if fareType == "day_pass" {
				hasDayPass = true
			}
		}

		// Validate price
		if dp.Price <= 0 {
			t.Errorf("Invalid price %f for %s", dp.Price, dp.ItemName)
		}

		// Validate location
		if dp.Location.Emirate != "Dubai" {
			t.Errorf("Expected emirate Dubai, got %s", dp.Location.Emirate)
		}
	}

	if !hasSilver {
		t.Error("Expected silver card metro fares")
	}

	if !hasGold {
		t.Error("Expected gold card metro fares")
	}

	if !hasDayPass {
		t.Error("Expected metro day pass fare")
	}
}

func TestRTAScraperBusFares(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Filter bus fares
	busFares := filterByMode(dataPoints, "bus")

	if len(busFares) == 0 {
		t.Fatal("Expected bus fares but got none")
	}

	// Validate bus fare structure
	hasZones := false
	hasDayPass := false

	for _, dp := range busFares {
		// Check zones
		if zones, ok := dp.Attributes["zones_crossed"]; ok && zones != 0 {
			hasZones = true
		}

		// Check for day pass
		if fareType, ok := dp.Attributes["fare_type"]; ok {
			if fareType == "day_pass" {
				hasDayPass = true
			}
		}

		// Validate price
		if dp.Price <= 0 {
			t.Errorf("Invalid price %f for %s", dp.Price, dp.ItemName)
		}
	}

	if !hasZones {
		t.Error("Expected zone-based bus fares")
	}

	if !hasDayPass {
		t.Error("Expected bus day pass fare")
	}
}

func TestRTAScraperTaxiFares(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Filter taxi fares
	taxiFares := filterBySubCategory(dataPoints, "Taxi")

	if len(taxiFares) == 0 {
		t.Fatal("Expected taxi fares but got none")
	}

	// Validate taxi fare types
	hasFlagDown := false
	hasPerKm := false
	hasMinimumFare := false

	for _, dp := range taxiFares {
		// Check fare types by item name
		itemName := dp.ItemName

		if rtaContains(itemName, "Flag Down") || rtaContains(itemName, "flag down") {
			hasFlagDown = true
		}

		if rtaContains(itemName, "Kilometer") || rtaContains(itemName, "kilometer") || rtaContains(itemName, "km") {
			hasPerKm = true
			// Validate unit
			if dp.Unit != "AED/km" {
				t.Errorf("Expected unit AED/km for per km fare, got %s", dp.Unit)
			}
		}

		if rtaContains(itemName, "Minimum") || rtaContains(itemName, "minimum") {
			hasMinimumFare = true
		}

		// Validate price
		if dp.Price <= 0 {
			t.Errorf("Invalid price %f for %s", dp.Price, dp.ItemName)
		}
	}

	if !hasFlagDown {
		t.Error("Expected flag down taxi fare")
	}

	if !hasPerKm {
		t.Error("Expected per kilometer taxi fare")
	}

	if !hasMinimumFare {
		t.Error("Expected minimum taxi fare")
	}
}

func TestRTAScraperDataQuality(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Validate each data point
	for i, dp := range dataPoints {
		// Required fields
		if dp.Category == "" {
			t.Errorf("Data point %d: missing category", i)
		}

		if dp.SubCategory == "" {
			t.Errorf("Data point %d: missing subcategory", i)
		}

		if dp.ItemName == "" {
			t.Errorf("Data point %d: missing item name", i)
		}

		if dp.Price <= 0 {
			t.Errorf("Data point %d: invalid price %f", i, dp.Price)
		}

		if dp.Unit == "" {
			t.Errorf("Data point %d: missing unit", i)
		}

		if dp.Source == "" {
			t.Errorf("Data point %d: missing source", i)
		}

		if dp.Source != "rta_official" {
			t.Errorf("Data point %d: expected source rta_official, got %s", i, dp.Source)
		}

		if dp.Confidence <= 0 || dp.Confidence > 1 {
			t.Errorf("Data point %d: invalid confidence %f", i, dp.Confidence)
		}

		// Location validation
		if dp.Location.Emirate == "" {
			t.Errorf("Data point %d: missing emirate", i)
		}

		if dp.Location.Emirate != "Dubai" {
			t.Errorf("Data point %d: expected emirate Dubai, got %s", i, dp.Location.Emirate)
		}

		// Attributes validation
		if len(dp.Attributes) == 0 && dp.SubCategory != "Taxi" {
			t.Errorf("Data point %d: missing attributes", i)
		}

		// Tags validation
		if len(dp.Tags) == 0 {
			t.Errorf("Data point %d: missing tags", i)
		}
	}

	// Validate using built-in validation
	err = rta.ValidateFareData(dataPoints)
	if err != nil {
		t.Errorf("Fare data validation failed: %v", err)
	}
}

func TestRTAScraperZoneConsistency(t *testing.T) {
	// Load fixture
	fixtureHTML, err := os.ReadFile("../fixtures/rta/fare_calculator.html")
	if err != nil {
		t.Fatalf("Failed to load fixture: %v", err)
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureHTML)
	}))
	defer server.Close()

	// Create scraper
	config := scrapers.Config{
		UserAgent: "test-agent",
		RateLimit: 1,
		Timeout:   30,
	}

	scraper := rta.NewRTAScraperWithURL(config, server.URL)

	// Execute scrape
	ctx := context.Background()
	dataPoints, err := scraper.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Group metro fares by card type and zones
	silverFares := make(map[int]float64)
	goldFares := make(map[int]float64)

	for _, dp := range dataPoints {
		mode, ok := dp.Attributes["transport_mode"]
		if !ok || mode != "metro" {
			continue
		}

		zones, hasZones := dp.Attributes["zones_crossed"]
		if !hasZones {
			continue
		}

		cardType, hasCard := dp.Attributes["card_type"]
		if !hasCard {
			continue
		}

		zonesInt := zones.(int)

		switch cardType {
		case "silver":
			silverFares[zonesInt] = dp.Price
		case "gold":
			goldFares[zonesInt] = dp.Price
		}
	}

	// Validate that fares increase with zones
	if len(silverFares) > 1 {
		prevZones := 0
		prevPrice := 0.0

		for zones, price := range silverFares {
			if prevZones > 0 && zones > prevZones {
				if price <= prevPrice {
					t.Errorf("Silver card fare for %d zones (%f) should be greater than %d zones (%f)",
						zones, price, prevZones, prevPrice)
				}
			}
			prevZones = zones
			prevPrice = price
		}
	}

	// Validate that gold is approximately 2x silver
	for zones := range silverFares {
		if goldPrice, hasGold := goldFares[zones]; hasGold {
			silverPrice := silverFares[zones]
			ratio := goldPrice / silverPrice

			if ratio < 1.8 || ratio > 2.2 {
				t.Errorf("Gold/Silver ratio for %d zones is %f (gold=%f, silver=%f), expected ~2.0",
					zones, ratio, goldPrice, silverPrice)
			}
		}
	}
}

// Helper functions

func filterByMode(dataPoints []*models.CostDataPoint, mode string) []*models.CostDataPoint {
	var filtered []*models.CostDataPoint
	for _, dp := range dataPoints {
		if m, ok := dp.Attributes["transport_mode"]; ok && m == mode {
			filtered = append(filtered, dp)
		}
	}
	return filtered
}

func rtaContains(s, substr string) bool {
	return len(s) >= len(substr) && rtaFindInString(s, substr) >= 0
}

func rtaFindInString(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
