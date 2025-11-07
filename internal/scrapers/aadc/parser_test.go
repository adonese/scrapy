package aadc

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFilsRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"with fils suffix", "6.7 fils", 6.7},
		{"with Fils capitalized", "6.7 Fils", 6.7},
		{"without suffix", "6.7", 6.7},
		{"integer value", "5 fils", 5.0},
		{"with comma thousands separator", "26,800 fils", 26800.0},
		{"with extra spaces", "  6.7  fils  ", 6.7},
		{"empty string", "", 0},
		{"invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFilsRate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAEDRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"with AED prefix", "AED 2.09", 2.09},
		{"without prefix", "8.55", 8.55},
		{"with aed lowercase", "aed 2.09", 2.09},
		{"integer value", "AED 5", 5.0},
		{"with comma thousands separator", "AED 2,000", 2000.0},
		{"with extra spaces", "  AED  2.09  ", 2.09},
		{"empty string", "", 0},
		{"invalid", "xyz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAEDRate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseElectricityTier(t *testing.T) {
	tests := []struct {
		name         string
		consumption  string
		rateText     string
		customerType string
		expectedMin  int
		expectedMax  int
		expectedRate float64
		expectNil    bool
		isUnlimited  bool
	}{
		{
			name:         "up to tier",
			consumption:  "Up to 30,000",
			rateText:     "5.8 fils",
			customerType: "national",
			expectedMin:  0,
			expectedMax:  30000,
			expectedRate: 5.8,
		},
		{
			name:         "above tier",
			consumption:  "Above 30,000",
			rateText:     "6.7 fils",
			customerType: "national",
			expectedMin:  30001,
			expectedMax:  0,
			expectedRate: 6.7,
			isUnlimited:  true,
		},
		{
			name:         "range tier",
			consumption:  "401 - 700",
			rateText:     "7.6 fils",
			customerType: "expatriate",
			expectedMin:  401,
			expectedMax:  700,
			expectedRate: 7.6,
		},
		{
			name:         "range with spaces",
			consumption:  "1,001 - 2,000",
			rateText:     "11.5 fils",
			customerType: "expatriate",
			expectedMin:  1001,
			expectedMax:  2000,
			expectedRate: 11.5,
		},
		{
			name:         "invalid rate",
			consumption:  "Up to 400",
			rateText:     "invalid",
			customerType: "expatriate",
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseElectricityTier(tt.consumption, tt.rateText, tt.customerType)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedMin, result.ConsumptionMin)
			assert.Equal(t, tt.expectedMax, result.ConsumptionMax)
			assert.Equal(t, tt.expectedRate, result.RateFils)
			assert.Equal(t, tt.customerType, result.CustomerType)
			assert.Equal(t, tt.isUnlimited, result.IsUnlimitedMax)
		})
	}
}

func TestParseElectricityRates(t *testing.T) {
	html := `
	<html>
		<body>
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
					<tbody>
						<tr>
							<td>Up to 400</td>
							<td>6.7 fils</td>
						</tr>
						<tr>
							<td>401 - 700</td>
							<td>7.6 fils</td>
						</tr>
					</tbody>
				</table>
			</section>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	rates, err := parseElectricityRates(doc)
	require.NoError(t, err)
	require.NotEmpty(t, rates)

	// Should find 4 rates total
	assert.Equal(t, 4, len(rates))

	// Check national rates
	nationals := filterByCustomerType(rates, "national")
	assert.Len(t, nationals, 2)
	assert.Equal(t, 5.8, nationals[0].RateFils)
	assert.Equal(t, 6.7, nationals[1].RateFils)

	// Check expatriate rates
	expats := filterByCustomerType(rates, "expatriate")
	assert.Len(t, expats, 2)
	assert.Equal(t, 6.7, expats[0].RateFils)
	assert.Equal(t, 7.6, expats[1].RateFils)
}

func TestParseWaterRates(t *testing.T) {
	html := `
	<html>
		<body>
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
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	rates, err := parseWaterRates(doc)
	require.NoError(t, err)
	require.NotEmpty(t, rates)

	// Should find 2 rates (one for nationals, one for expats)
	assert.Equal(t, 2, len(rates))

	// Check rates
	var nationalRate, expatRate *WaterRate
	for i := range rates {
		if rates[i].CustomerType == "national" {
			nationalRate = &rates[i]
		} else if rates[i].CustomerType == "expatriate" {
			expatRate = &rates[i]
		}
	}

	require.NotNil(t, nationalRate)
	require.NotNil(t, expatRate)

	assert.Equal(t, 2.09, nationalRate.RateAED)
	assert.Equal(t, 8.55, expatRate.RateAED)
	assert.Equal(t, "1000_ig", nationalRate.Unit)
	assert.Equal(t, "1000_ig", expatRate.Unit)
}

func TestConvertElectricityToDataPoints(t *testing.T) {
	rates := []ElectricityRate{
		{
			ConsumptionMin: 0,
			ConsumptionMax: 400,
			RateFils:       6.7,
			CustomerType:   "expatriate",
		},
		{
			ConsumptionMin:  10001,
			ConsumptionMax:  0,
			RateFils:        28.7,
			CustomerType:    "expatriate",
			IsUnlimitedMax:  true,
		},
	}

	sourceURL := "https://www.aadc.ae/en/pages/maintarrif.aspx"
	dataPoints := convertElectricityToDataPoints(rates, sourceURL)

	require.Len(t, dataPoints, 2)

	// Check first data point
	dp1 := dataPoints[0]
	assert.Equal(t, "Utilities", dp1.Category)
	assert.Equal(t, "Electricity", dp1.SubCategory)
	assert.Contains(t, dp1.ItemName, "AADC Electricity")
	assert.Contains(t, dp1.ItemName, "Expatriate")
	assert.Equal(t, 0.067, dp1.Price) // 6.7 fils = 0.067 AED
	assert.Equal(t, "Abu Dhabi", dp1.Location.Emirate)
	assert.Equal(t, "aadc_official", dp1.Source)
	assert.Equal(t, sourceURL, dp1.SourceURL)
	assert.Equal(t, float32(0.98), dp1.Confidence)
	assert.Equal(t, "AED per kWh", dp1.Unit)
	assert.Contains(t, dp1.Tags, "electricity")
	assert.Contains(t, dp1.Tags, "expatriate")
	assert.Equal(t, "expatriate", dp1.Attributes["customer_type"])
	assert.Equal(t, 6.7, dp1.Attributes["fils_rate"])
	assert.Equal(t, 0, dp1.Attributes["tier_min_kwh"])
	assert.Equal(t, 400, dp1.Attributes["tier_max_kwh"])

	// Check second data point (unlimited)
	dp2 := dataPoints[1]
	assert.Equal(t, 0.287, dp2.Price) // 28.7 fils = 0.287 AED
	assert.Equal(t, "unlimited", dp2.Attributes["tier_max_kwh"])
	assert.Equal(t, 10001, dp2.Attributes["tier_min_kwh"])
}

func TestConvertWaterToDataPoints(t *testing.T) {
	rates := []WaterRate{
		{
			RateAED:      2.09,
			CustomerType: "national",
			Unit:         "1000_ig",
		},
		{
			RateAED:      8.55,
			CustomerType: "expatriate",
			Unit:         "1000_ig",
		},
	}

	sourceURL := "https://www.aadc.ae/en/pages/maintarrif.aspx"
	dataPoints := convertWaterToDataPoints(rates, sourceURL)

	require.Len(t, dataPoints, 2)

	// Check national rate
	dp1 := dataPoints[0]
	assert.Equal(t, "Utilities", dp1.Category)
	assert.Equal(t, "Water", dp1.SubCategory)
	assert.Contains(t, dp1.ItemName, "AADC Water")
	assert.Contains(t, dp1.ItemName, "National")
	assert.Equal(t, 2.09, dp1.Price)
	assert.Equal(t, "Abu Dhabi", dp1.Location.Emirate)
	assert.Equal(t, "aadc_official", dp1.Source)
	assert.Equal(t, float32(0.98), dp1.Confidence)
	assert.Equal(t, "AED per 1000 IG", dp1.Unit)
	assert.Contains(t, dp1.Tags, "water")
	assert.Contains(t, dp1.Tags, "national")
	assert.Equal(t, "national", dp1.Attributes["customer_type"])
	assert.Equal(t, "flat", dp1.Attributes["rate_type"])

	// Check expatriate rate
	dp2 := dataPoints[1]
	assert.Equal(t, 8.55, dp2.Price)
	assert.Contains(t, dp2.ItemName, "Expatriate")
	assert.Equal(t, "expatriate", dp2.Attributes["customer_type"])
}

func TestFormatElectricityTierName(t *testing.T) {
	tests := []struct {
		name     string
		rate     ElectricityRate
		expected string
	}{
		{
			name: "up to tier",
			rate: ElectricityRate{
				ConsumptionMin: 0,
				ConsumptionMax: 400,
			},
			expected: "Tier Up to 400 kWh",
		},
		{
			name: "above tier",
			rate: ElectricityRate{
				ConsumptionMin:  10001,
				ConsumptionMax:  0,
				IsUnlimitedMax:  true,
			},
			expected: "Tier Above 10000 kWh",
		},
		{
			name: "range tier",
			rate: ElectricityRate{
				ConsumptionMin: 401,
				ConsumptionMax: 700,
			},
			expected: "Tier 401-700 kWh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatElectricityTierName(tt.rate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseElectricityRates_EmptyDocument(t *testing.T) {
	html := `<html><body></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	rates, err := parseElectricityRates(doc)
	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "no electricity rates found")
}

func TestParseWaterRates_EmptyDocument(t *testing.T) {
	html := `<html><body></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	rates, err := parseWaterRates(doc)
	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "no water rates found")
}

// Helper function for tests
func filterByCustomerType(rates []ElectricityRate, customerType string) []ElectricityRate {
	filtered := []ElectricityRate{}
	for _, rate := range rates {
		if rate.CustomerType == customerType {
			filtered = append(filtered, rate)
		}
	}
	return filtered
}
