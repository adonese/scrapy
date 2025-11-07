package sewa

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/models"
)

func TestParseConsumptionRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMin  int
		wantMax  int
	}{
		{
			name:    "simple range",
			input:   "1 - 3,000",
			wantMin: 1,
			wantMax: 3000,
		},
		{
			name:    "range with spaces",
			input:   "3,001 - 5,000",
			wantMin: 3001,
			wantMax: 5000,
		},
		{
			name:    "above format",
			input:   "Above 10,000",
			wantMin: 10000,
			wantMax: -1,
		},
		{
			name:    "above format lowercase",
			input:   "above 5,000",
			wantMin: 5000,
			wantMax: -1,
		},
		{
			name:    "range without comma",
			input:   "5001 - 10000",
			wantMin: 5001,
			wantMax: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := parseConsumptionRange(tt.input)
			assert.Equal(t, tt.wantMin, min, "min consumption should match")
			assert.Equal(t, tt.wantMax, max, "max consumption should match")
		})
	}
}

func TestParseRateValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{
			name:  "integer fils",
			input: "14 fils",
			want:  14.0,
		},
		{
			name:  "decimal fils",
			input: "27.5 fils",
			want:  27.5,
		},
		{
			name:  "no unit",
			input: "32",
			want:  32.0,
		},
		{
			name:  "with extra spaces",
			input: "  18 fils  ",
			want:  18.0,
		},
		{
			name:  "uppercase",
			input: "38 FILS",
			want:  38.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRateValue(tt.input)
			assert.Equal(t, tt.want, got, "rate value should match")
		})
	}
}

func TestParseAEDValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{
			name:  "simple AED value",
			input: "AED 8.00",
			want:  8.0,
		},
		{
			name:  "no decimals",
			input: "AED 15",
			want:  15.0,
		},
		{
			name:  "lowercase aed",
			input: "aed 8.00",
			want:  8.0,
		},
		{
			name:  "just number",
			input: "15.00",
			want:  15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAEDValue(tt.input)
			assert.Equal(t, tt.want, got, "AED value should match")
		})
	}
}

func TestParseElectricityTariff(t *testing.T) {
	// Load fixture
	html := loadFixture(t, "../../../test/fixtures/sewa/tariff_page.html")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	t.Run("emirati electricity rates", func(t *testing.T) {
		rates := parseElectricityTariff(doc, TariffEmirati)

		// Should have 4 tiers
		assert.Len(t, rates, 4, "should have 4 electricity tiers for Emirati")

		// Check first tier
		assert.Equal(t, 1, rates[0].MinConsumption)
		assert.Equal(t, 3000, rates[0].MaxConsumption)
		assert.Equal(t, 14.0, rates[0].Rate)

		// Check second tier
		assert.Equal(t, 3001, rates[1].MinConsumption)
		assert.Equal(t, 5000, rates[1].MaxConsumption)
		assert.Equal(t, 18.0, rates[1].Rate)

		// Check third tier
		assert.Equal(t, 5001, rates[2].MinConsumption)
		assert.Equal(t, 10000, rates[2].MaxConsumption)
		assert.Equal(t, 27.5, rates[2].Rate)

		// Check fourth tier (above 10,000)
		assert.Equal(t, 10000, rates[3].MinConsumption)
		assert.Equal(t, -1, rates[3].MaxConsumption, "should be unlimited")
		assert.Equal(t, 32.0, rates[3].Rate)
	})

	t.Run("expatriate electricity rates", func(t *testing.T) {
		rates := parseElectricityTariff(doc, TariffExpatriate)

		// Should have 3 tiers
		assert.Len(t, rates, 3, "should have 3 electricity tiers for Expatriate")

		// Check first tier
		assert.Equal(t, 1, rates[0].MinConsumption)
		assert.Equal(t, 3000, rates[0].MaxConsumption)
		assert.Equal(t, 27.5, rates[0].Rate)

		// Check second tier
		assert.Equal(t, 3001, rates[1].MinConsumption)
		assert.Equal(t, 5000, rates[1].MaxConsumption)
		assert.Equal(t, 32.0, rates[1].Rate)

		// Check third tier (above 5,000)
		assert.Equal(t, 5000, rates[2].MinConsumption)
		assert.Equal(t, -1, rates[2].MaxConsumption, "should be unlimited")
		assert.Equal(t, 38.0, rates[2].Rate)
	})
}

func TestParseWaterTariff(t *testing.T) {
	// Load fixture
	html := loadFixture(t, "../../../test/fixtures/sewa/tariff_page.html")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	dataPoints := parseWaterTariff(doc)

	// Should have 2 water rates (Emirati and Expatriate)
	assert.Len(t, dataPoints, 2, "should have 2 water rates")

	// Find Emirati rate
	var emiratiRate, expatRate *models.CostDataPoint
	for _, dp := range dataPoints {
		customerType := dp.Attributes["customer_type"].(string)
		if customerType == "emirati" {
			emiratiRate = dp
		} else if customerType == "expatriate" {
			expatRate = dp
		}
	}

	require.NotNil(t, emiratiRate, "should have Emirati water rate")
	require.NotNil(t, expatRate, "should have Expatriate water rate")

	// Check Emirati water rate
	assert.Equal(t, "Utilities", emiratiRate.Category)
	assert.Equal(t, "Water", emiratiRate.SubCategory)
	assert.Contains(t, emiratiRate.ItemName, "Emirati")
	assert.Equal(t, 8.0, emiratiRate.Price)
	assert.Equal(t, "Sharjah", emiratiRate.Location.Emirate)
	assert.Equal(t, float32(0.98), emiratiRate.Confidence)

	// Check Expatriate water rate
	assert.Equal(t, "Utilities", expatRate.Category)
	assert.Equal(t, "Water", expatRate.SubCategory)
	assert.Contains(t, expatRate.ItemName, "Expatriate")
	assert.Equal(t, 15.0, expatRate.Price)
	assert.Equal(t, "Sharjah", expatRate.Location.Emirate)
}

func TestParseSewerageInfo(t *testing.T) {
	// Load fixture
	html := loadFixture(t, "../../../test/fixtures/sewa/tariff_page.html")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	now := time.Now()
	dataPoint := parseSewerageInfo(doc, "https://test.com", now)

	require.NotNil(t, dataPoint, "should have sewerage data point")

	assert.Equal(t, "Utilities", dataPoint.Category)
	assert.Equal(t, "Sewerage", dataPoint.SubCategory)
	assert.Equal(t, "SEWA Sewerage Charge", dataPoint.ItemName)
	assert.Equal(t, 0.50, dataPoint.Price, "50% should be stored as 0.50")
	assert.Equal(t, "Sharjah", dataPoint.Location.Emirate)
	assert.Equal(t, float32(0.98), dataPoint.Confidence)
	assert.Contains(t, dataPoint.Attributes, "calculation_method")
}

func TestParseSEWATariffs(t *testing.T) {
	// Load fixture
	html := loadFixture(t, "../../../test/fixtures/sewa/tariff_page.html")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	dataPoints, err := parseSEWATariffs(doc, "https://www.sewa.gov.ae/en/content/tariff")
	require.NoError(t, err)
	require.NotEmpty(t, dataPoints)

	// Expected breakdown:
	// - 4 Emirati electricity tiers
	// - 3 Expatriate electricity tiers
	// - 2 water rates (Emirati + Expatriate)
	// - 1 sewerage charge
	// Total: 10 data points
	assert.Len(t, dataPoints, 10, "should have 10 total data points")

	// Verify categories
	electricityCount := 0
	waterCount := 0
	sewerageCount := 0

	for _, dp := range dataPoints {
		assert.Equal(t, "Utilities", dp.Category, "all should be Utilities category")
		assert.Equal(t, "Sharjah", dp.Location.Emirate, "all should be in Sharjah")
		assert.Equal(t, "sewa_official", dp.Source, "all should have sewa_official source")
		assert.Equal(t, float32(0.98), dp.Confidence, "all should have high confidence")

		switch dp.SubCategory {
		case "Electricity":
			electricityCount++
		case "Water":
			waterCount++
		case "Sewerage":
			sewerageCount++
		}
	}

	assert.Equal(t, 7, electricityCount, "should have 7 electricity data points")
	assert.Equal(t, 2, waterCount, "should have 2 water data points")
	assert.Equal(t, 1, sewerageCount, "should have 1 sewerage data point")
}

func TestCreateElectricityDataPoint(t *testing.T) {
	rate := ElectricityRate{
		MinConsumption: 1,
		MaxConsumption: 3000,
		Rate:           14.0,
	}
	now := time.Now()

	dp := createElectricityDataPoint(rate, TariffEmirati, "https://test.com", now)

	assert.Equal(t, "Utilities", dp.Category)
	assert.Equal(t, "Electricity", dp.SubCategory)
	assert.Contains(t, dp.ItemName, "1-3000 kWh")
	assert.Contains(t, dp.ItemName, "Emirati")
	assert.Equal(t, 0.14, dp.Price, "14 fils should convert to 0.14 AED")
	assert.Equal(t, "Sharjah", dp.Location.Emirate)
	assert.Equal(t, "sewa_official", dp.Source)
	assert.Equal(t, float32(0.98), dp.Confidence)
	assert.Equal(t, "AED per kWh", dp.Unit)

	// Check attributes
	assert.Equal(t, 1, dp.Attributes["consumption_range_min"])
	assert.Equal(t, 3000, dp.Attributes["consumption_range_max"])
	assert.Equal(t, "emirati", dp.Attributes["customer_type"])
	assert.Equal(t, 14.0, dp.Attributes["rate_fils"])
}

func TestCreateElectricityDataPointUnlimited(t *testing.T) {
	rate := ElectricityRate{
		MinConsumption: 10000,
		MaxConsumption: -1,
		Rate:           32.0,
	}
	now := time.Now()

	dp := createElectricityDataPoint(rate, TariffExpatriate, "https://test.com", now)

	assert.Contains(t, dp.ItemName, "Above 10000 kWh")
	assert.Contains(t, dp.ItemName, "Expatriate")
	assert.Equal(t, 0.32, dp.Price, "32 fils should convert to 0.32 AED")

	// Should not have max consumption attribute for unlimited tier
	_, hasMax := dp.Attributes["consumption_range_max"]
	assert.False(t, hasMax, "unlimited tier should not have max consumption")
}

// loadFixture loads an HTML fixture file
func loadFixture(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture")
	return string(content)
}
