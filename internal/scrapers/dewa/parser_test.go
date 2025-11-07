package dewa

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConsumptionRange(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Range with dash",
			input:       "0 - 2,000",
			expectedMin: 0,
			expectedMax: 2000,
		},
		{
			name:        "Range without spaces",
			input:       "2001-4000",
			expectedMin: 2001,
			expectedMax: 4000,
		},
		{
			name:        "Range with commas",
			input:       "4,001 - 6,000",
			expectedMin: 4001,
			expectedMax: 6000,
		},
		{
			name:        "Above pattern",
			input:       "Above 6,000",
			expectedMin: 6001,
			expectedMax: -1,
		},
		{
			name:        "Over pattern",
			input:       "Over 10,000",
			expectedMin: 10001,
			expectedMax: -1,
		},
		{
			name:        "Water range",
			input:       "0 - 5,000",
			expectedMin: 0,
			expectedMax: 5000,
		},
		{
			name:        "Empty string",
			input:       "",
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := parseConsumptionRange(tt.input)
			assert.Equal(t, tt.expectedMin, min, "Min mismatch for: %s", tt.input)
			assert.Equal(t, tt.expectedMax, max, "Max mismatch for: %s", tt.input)
		})
	}
}

func TestExtractRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Rate with fils suffix",
			input:    "23.0 fils",
			expected: 23.0,
		},
		{
			name:     "Rate without decimal",
			input:    "28 fils",
			expected: 28,
		},
		{
			name:     "Rate with decimal",
			input:    "3.57 fils",
			expected: 3.57,
		},
		{
			name:     "Rate in parentheses",
			input:    "(currently 6.5 fils/kWh)",
			expected: 6.5,
		},
		{
			name:     "Just number",
			input:    "38.0",
			expected: 38.0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "No number",
			input:    "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRate(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestParseRateSlab(t *testing.T) {
	tests := []struct {
		name         string
		slabText     string
		rateText     string
		utilityType  string
		expectedName string
		expectedMin  int
		expectedMax  int
		expectedRate float64
		expectedUnit string
	}{
		{
			name:         "Electricity first slab",
			slabText:     "0 - 2,000",
			rateText:     "23.0 fils",
			utilityType:  "electricity",
			expectedName: "Slab 0-2000 kWh",
			expectedMin:  0,
			expectedMax:  2000,
			expectedRate: 23.0,
			expectedUnit: "fils_per_kwh",
		},
		{
			name:         "Electricity highest slab",
			slabText:     "Above 6,000",
			rateText:     "38.0 fils",
			utilityType:  "electricity",
			expectedName: "Slab 6001+ kWh",
			expectedMin:  6001,
			expectedMax:  -1,
			expectedRate: 38.0,
			expectedUnit: "fils_per_kwh",
		},
		{
			name:         "Water first slab",
			slabText:     "0 - 5,000",
			rateText:     "3.57 fils",
			utilityType:  "water",
			expectedName: "Slab 0-5000 IG",
			expectedMin:  0,
			expectedMax:  5000,
			expectedRate: 3.57,
			expectedUnit: "fils_per_ig",
		},
		{
			name:         "Water highest slab",
			slabText:     "Above 10,000",
			rateText:     "10.52 fils",
			utilityType:  "water",
			expectedName: "Slab 10001+ IG",
			expectedMin:  10001,
			expectedMax:  -1,
			expectedRate: 10.52,
			expectedUnit: "fils_per_ig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRateSlab(tt.slabText, tt.rateText, tt.utilityType)
			assert.Equal(t, tt.expectedName, result.SlabName)
			assert.Equal(t, tt.expectedMin, result.MinRange)
			assert.Equal(t, tt.expectedMax, result.MaxRange)
			assert.Equal(t, tt.expectedRate, result.Rate)
			assert.Equal(t, tt.expectedUnit, result.Unit)
		})
	}
}

func TestParseElectricitySlabs(t *testing.T) {
	html := `
	<html>
		<body>
			<section class="electricity-tariff">
				<h2>Electricity Tariff (Residential)</h2>
				<table class="tariff-table">
					<thead>
						<tr>
							<th>Consumption Slab (kWh)</th>
							<th>Rate (Fils/kWh)</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td>0 - 2,000</td>
							<td>23.0 fils</td>
						</tr>
						<tr>
							<td>2,001 - 4,000</td>
							<td>28.0 fils</td>
						</tr>
						<tr>
							<td>4,001 - 6,000</td>
							<td>32.0 fils</td>
						</tr>
						<tr>
							<td>Above 6,000</td>
							<td>38.0 fils</td>
						</tr>
					</tbody>
				</table>
			</section>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	slabs, err := parseElectricitySlabs(doc)
	require.NoError(t, err)
	assert.Len(t, slabs, 4)

	// Check first slab
	assert.Equal(t, 0, slabs[0].MinRange)
	assert.Equal(t, 2000, slabs[0].MaxRange)
	assert.Equal(t, 23.0, slabs[0].Rate)
	assert.Equal(t, "fils_per_kwh", slabs[0].Unit)

	// Check last slab (Above 6000)
	assert.Equal(t, 6001, slabs[3].MinRange)
	assert.Equal(t, -1, slabs[3].MaxRange)
	assert.Equal(t, 38.0, slabs[3].Rate)
}

func TestParseWaterSlabs(t *testing.T) {
	html := `
	<html>
		<body>
			<section class="water-tariff">
				<h2>Water Tariff (Residential)</h2>
				<table class="tariff-table">
					<thead>
						<tr>
							<th>Consumption Slab (IG)</th>
							<th>Rate (Fils/IG)</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td>0 - 5,000</td>
							<td>3.57 fils</td>
						</tr>
						<tr>
							<td>5,001 - 10,000</td>
							<td>5.24 fils</td>
						</tr>
						<tr>
							<td>Above 10,000</td>
							<td>10.52 fils</td>
						</tr>
					</tbody>
				</table>
			</section>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	slabs, err := parseWaterSlabs(doc)
	require.NoError(t, err)
	assert.Len(t, slabs, 3)

	// Check first slab
	assert.Equal(t, 0, slabs[0].MinRange)
	assert.Equal(t, 5000, slabs[0].MaxRange)
	assert.Equal(t, 3.57, slabs[0].Rate)
	assert.Equal(t, "fils_per_ig", slabs[0].Unit)

	// Check last slab
	assert.Equal(t, 10001, slabs[2].MinRange)
	assert.Equal(t, -1, slabs[2].MaxRange)
	assert.Equal(t, 10.52, slabs[2].Rate)
}

func TestParseFuelSurcharge(t *testing.T) {
	html := `
	<html>
		<body>
			<section class="electricity-tariff">
				<p class="note">Fuel Surcharge: Variable (currently 6.5 fils/kWh)</p>
			</section>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	slab := parseFuelSurcharge(doc)
	require.NotNil(t, slab)

	assert.Equal(t, "Fuel Surcharge", slab.SlabName)
	assert.Equal(t, 6.5, slab.Rate)
	assert.Equal(t, "fils_per_kwh", slab.Unit)
	assert.Equal(t, 0, slab.MinRange)
	assert.Equal(t, -1, slab.MaxRange)
}

func TestParseFuelSurcharge_NotFound(t *testing.T) {
	html := `
	<html>
		<body>
			<section class="electricity-tariff">
				<p>No fuel surcharge information</p>
			</section>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	slab := parseFuelSurcharge(doc)
	assert.Nil(t, slab)
}

func TestSlabToDataPoint(t *testing.T) {
	slab := RateSlab{
		SlabName: "Slab 0-2000 kWh",
		MinRange: 0,
		MaxRange: 2000,
		Rate:     23.0,
		Unit:     "fils_per_kwh",
	}

	dp := slabToDataPoint(slab, "electricity", "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff")

	assert.Equal(t, "Utilities", dp.Category)
	assert.Equal(t, "Electricity", dp.SubCategory)
	assert.Equal(t, "DEWA Electricity Slab 0-2000 kWh", dp.ItemName)
	assert.Equal(t, 0.23, dp.Price) // 23 fils = 0.23 AED
	assert.Equal(t, "Dubai", dp.Location.Emirate)
	assert.Equal(t, "dewa_official", dp.Source)
	assert.Equal(t, float32(0.98), dp.Confidence)
	assert.Equal(t, "AED", dp.Unit)
	assert.Contains(t, dp.Tags, "utility")
	assert.Contains(t, dp.Tags, "electricity")
	assert.Equal(t, "fils_per_kwh", dp.Attributes["unit"])
	assert.Equal(t, 0, dp.Attributes["consumption_range_min"])
	assert.Equal(t, 2000, dp.Attributes["consumption_range_max"])
}

func TestSlabToDataPoint_WaterSlab(t *testing.T) {
	slab := RateSlab{
		SlabName: "Slab 0-5000 IG",
		MinRange: 0,
		MaxRange: 5000,
		Rate:     3.57,
		Unit:     "fils_per_ig",
	}

	dp := slabToDataPoint(slab, "water", "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff")

	assert.Equal(t, "Utilities", dp.Category)
	assert.Equal(t, "Water", dp.SubCategory)
	assert.Equal(t, "DEWA Water Slab 0-5000 IG", dp.ItemName)
	assert.InDelta(t, 0.0357, dp.Price, 0.0001) // 3.57 fils = 0.0357 AED
	assert.Contains(t, dp.Tags, "water")
}

func TestSlabToDataPoint_FuelSurcharge(t *testing.T) {
	slab := RateSlab{
		SlabName: "Fuel Surcharge",
		MinRange: 0,
		MaxRange: -1,
		Rate:     6.5,
		Unit:     "fils_per_kwh",
	}

	dp := slabToDataPoint(slab, "fuel_surcharge", "https://www.dewa.gov.ae/en/consumer/billing/slab-tariff")

	assert.Equal(t, "Utilities", dp.Category)
	assert.Equal(t, "Fuel Surcharge", dp.SubCategory)
	assert.Equal(t, "DEWA Fuel Surcharge", dp.ItemName)
	assert.Equal(t, 0.065, dp.Price) // 6.5 fils = 0.065 AED
}

func TestParseElectricitySlabs_EmptyHTML(t *testing.T) {
	html := `<html><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	_, err = parseElectricitySlabs(doc)
	assert.Error(t, err)
}

func TestParseWaterSlabs_EmptyHTML(t *testing.T) {
	html := `<html><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	_, err = parseWaterSlabs(doc)
	assert.Error(t, err)
}
