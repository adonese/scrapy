package rta

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Simple AED", "AED 5", 5.0},
		{"AED with decimal", "AED 7.50", 7.5},
		{"Without AED", "3.00", 3.0},
		{"With Dhs", "Dhs 10", 10.0},
		{"With commas", "AED 1,234.56", 1234.56},
		{"Empty string", "", 0.0},
		{"Dash", "-", 0.0},
		{"N/A", "N/A", 0.0},
		{"Only number", "12", 12.0},
		{"Number with spaces", "  20  ", 20.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrice(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Single digit", "1 Zone", 1},
		{"Two digits", "12 Zones", 12},
		{"Number only", "5", 5},
		{"Mixed text", "Zone 3 to 5", 3}, // Gets first number
		{"No number", "All Zones", 0},
		{"Empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNumber(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestNormalizeFareTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Day Flag Down", "day_flag_down"},
		{"Per Kilometer", "per_kilometer"},
		{"Waiting Time (per minute)", "waiting_time_per_minute"},
		{"Simple", "simple"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeFareTypeName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseMetroFares(t *testing.T) {
	html := `
	<html>
	<body>
		<section class="metro-fares">
			<h2>Dubai Metro Fares</h2>
			<table class="fare-table">
				<thead>
					<tr>
						<th>Zone</th>
						<th>Silver (Standard)</th>
						<th>Gold (Premium)</th>
						<th>Day Pass</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Regular</td>
						<td>AED 4</td>
						<td>AED 8</td>
						<td>AED 20</td>
					</tr>
					<tr>
						<td>1 Zone</td>
						<td>AED 3</td>
						<td>AED 6</td>
						<td>-</td>
					</tr>
					<tr>
						<td>2 Zones</td>
						<td>AED 5</td>
						<td>AED 10</td>
						<td>-</td>
					</tr>
				</tbody>
			</table>
		</section>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	dataPoints := parseMetroFares(doc, "https://test.com", time.Now())

	if len(dataPoints) == 0 {
		t.Fatal("Expected metro fares but got none")
	}

	// Should have Regular (2), 1 Zone (2), 2 Zones (2), and Day Pass (1) = 7 total
	// Regular: Silver + Gold + Day Pass
	// 1 Zone: Silver + Gold
	// 2 Zones: Silver + Gold
	if len(dataPoints) < 5 {
		t.Errorf("Expected at least 5 data points, got %d", len(dataPoints))
	}

	// Check that we have both card types
	hasSilver := false
	hasGold := false

	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}

		if dp.SubCategory != "Public Transport" {
			t.Errorf("Expected subcategory Public Transport, got %s", dp.SubCategory)
		}

		if dp.Location.Emirate != "Dubai" {
			t.Errorf("Expected emirate Dubai, got %s", dp.Location.Emirate)
		}

		attrs, ok := dp.Attributes["card_type"]
		if ok {
			cardType := attrs.(string)
			if cardType == "silver" {
				hasSilver = true
			} else if cardType == "gold" {
				hasGold = true
			}
		}
	}

	if !hasSilver {
		t.Error("Expected at least one silver card fare")
	}

	if !hasGold {
		t.Error("Expected at least one gold card fare")
	}
}

func TestParseBusFares(t *testing.T) {
	html := `
	<html>
	<body>
		<section class="bus-fares">
			<h2>Dubai Bus Fares</h2>
			<table class="fare-table">
				<thead>
					<tr>
						<th>Journey Type</th>
						<th>Fare (AED)</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Single Journey (1 Zone)</td>
						<td>AED 3</td>
					</tr>
					<tr>
						<td>Single Journey (2 Zones)</td>
						<td>AED 5</td>
					</tr>
					<tr>
						<td>Day Pass</td>
						<td>AED 20</td>
					</tr>
				</tbody>
			</table>
		</section>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	dataPoints := parseBusFares(doc, "https://test.com", time.Now())

	if len(dataPoints) == 0 {
		t.Fatal("Expected bus fares but got none")
	}

	// Should have 3 data points
	if len(dataPoints) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(dataPoints))
	}

	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}

		if dp.SubCategory != "Public Transport" {
			t.Errorf("Expected subcategory Public Transport, got %s", dp.SubCategory)
		}

		mode, ok := dp.Attributes["transport_mode"]
		if !ok || mode != "bus" {
			t.Errorf("Expected transport_mode bus, got %v", mode)
		}
	}
}

func TestParseTramFares(t *testing.T) {
	html := `
	<html>
	<body>
		<section class="tram-fares">
			<h2>Dubai Tram Fares</h2>
			<table class="fare-table">
				<thead>
					<tr>
						<th>Ticket Type</th>
						<th>Silver</th>
						<th>Gold</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Single Journey</td>
						<td>AED 3</td>
						<td>AED 6</td>
					</tr>
					<tr>
						<td>Day Pass</td>
						<td>AED 15</td>
						<td>AED 30</td>
					</tr>
				</tbody>
			</table>
		</section>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	dataPoints := parseTramFares(doc, "https://test.com", time.Now())

	if len(dataPoints) == 0 {
		t.Fatal("Expected tram fares but got none")
	}

	// Should have 4 data points (2 rows x 2 card types)
	if len(dataPoints) != 4 {
		t.Errorf("Expected 4 data points, got %d", len(dataPoints))
	}

	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}

		mode, ok := dp.Attributes["transport_mode"]
		if !ok || mode != "tram" {
			t.Errorf("Expected transport_mode tram, got %v", mode)
		}
	}
}

func TestParseTaxiFares(t *testing.T) {
	html := `
	<html>
	<body>
		<section class="taxi-info">
			<h2>Dubai Taxi Base Fares</h2>
			<table class="taxi-fares">
				<tbody>
					<tr>
						<td>Day Flag Down (6 AM - 10 PM)</td>
						<td>AED 5</td>
					</tr>
					<tr>
						<td>Night Flag Down (10 PM - 6 AM)</td>
						<td>AED 5.50</td>
					</tr>
					<tr>
						<td>Per Kilometer</td>
						<td>AED 1.96</td>
					</tr>
					<tr>
						<td>Minimum Fare</td>
						<td>AED 12</td>
					</tr>
				</tbody>
			</table>
		</section>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	dataPoints := parseTaxiFares(doc, "https://test.com", time.Now())

	if len(dataPoints) == 0 {
		t.Fatal("Expected taxi fares but got none")
	}

	// Should have 4 data points
	if len(dataPoints) != 4 {
		t.Errorf("Expected 4 data points, got %d", len(dataPoints))
	}

	hasPerKm := false
	for _, dp := range dataPoints {
		if dp.Category != "Transportation" {
			t.Errorf("Expected category Transportation, got %s", dp.Category)
		}

		if dp.SubCategory != "Taxi" {
			t.Errorf("Expected subcategory Taxi, got %s", dp.SubCategory)
		}

		// Check for per kilometer fare with correct unit
		if strings.Contains(dp.ItemName, "Kilometer") || strings.Contains(dp.ItemName, "kilometer") {
			hasPerKm = true
			if dp.Unit != "AED/km" {
				t.Errorf("Expected unit AED/km for per kilometer fare, got %s", dp.Unit)
			}
		}
	}

	if !hasPerKm {
		t.Error("Expected to find per kilometer fare")
	}
}

func TestParseFares(t *testing.T) {
	// Use the actual fixture HTML
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>RTA Public Transport Fares</title>
</head>
<body>
    <div class="fare-information">
        <h1>Dubai Public Transport Fares</h1>
        <section class="metro-fares">
            <h2>Dubai Metro Fares</h2>
            <table class="fare-table">
                <thead>
                    <tr>
                        <th>Zone</th>
                        <th>Silver (Standard)</th>
                        <th>Gold (Premium)</th>
                        <th>Day Pass</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Regular</td>
                        <td>AED 4</td>
                        <td>AED 8</td>
                        <td>AED 20</td>
                    </tr>
                    <tr>
                        <td>1 Zone</td>
                        <td>AED 3</td>
                        <td>AED 6</td>
                        <td>-</td>
                    </tr>
                    <tr>
                        <td>2 Zones</td>
                        <td>AED 5</td>
                        <td>AED 10</td>
                        <td>-</td>
                    </tr>
                    <tr>
                        <td>All Zones</td>
                        <td>AED 7.50</td>
                        <td>AED 15</td>
                        <td>-</td>
                    </tr>
                </tbody>
            </table>
        </section>
        <section class="bus-fares">
            <h2>Dubai Bus Fares</h2>
            <table class="fare-table">
                <tbody>
                    <tr>
                        <td>Single Journey (1 Zone)</td>
                        <td>AED 3</td>
                    </tr>
                    <tr>
                        <td>Single Journey (2 Zones)</td>
                        <td>AED 5</td>
                    </tr>
                    <tr>
                        <td>Single Journey (All Zones)</td>
                        <td>AED 7.50</td>
                    </tr>
                    <tr>
                        <td>Day Pass</td>
                        <td>AED 20</td>
                    </tr>
                </tbody>
            </table>
        </section>
        <section class="tram-fares">
            <h2>Dubai Tram Fares</h2>
            <table class="fare-table">
                <tbody>
                    <tr>
                        <td>Single Journey</td>
                        <td>AED 3</td>
                        <td>AED 6</td>
                    </tr>
                    <tr>
                        <td>Day Pass</td>
                        <td>AED 15</td>
                        <td>AED 30</td>
                    </tr>
                </tbody>
            </table>
        </section>
        <section class="taxi-info">
            <h2>Dubai Taxi Base Fares</h2>
            <table class="taxi-fares">
                <tbody>
                    <tr>
                        <td>Day Flag Down (6 AM - 10 PM)</td>
                        <td>AED 5</td>
                    </tr>
                    <tr>
                        <td>Per Kilometer</td>
                        <td>AED 1.96</td>
                    </tr>
                </tbody>
            </table>
        </section>
    </div>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	dataPoints, err := ParseFares(doc, "https://test.com")
	if err != nil {
		t.Fatalf("ParseFares failed: %v", err)
	}

	if len(dataPoints) == 0 {
		t.Fatal("Expected data points but got none")
	}

	// Should have extracted data from all sections
	t.Logf("Total data points extracted: %d", len(dataPoints))

	// Count by mode
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

	if metroCount == 0 {
		t.Error("Expected metro fares")
	}

	if busCount == 0 {
		t.Error("Expected bus fares")
	}

	if tramCount == 0 {
		t.Error("Expected tram fares")
	}

	if taxiCount == 0 {
		t.Error("Expected taxi fares")
	}
}

func TestCreateFareDataPoint(t *testing.T) {
	timestamp := time.Now()
	dp := createFareDataPoint(
		ModeMetro,
		CardSilver,
		FareSingleJourney,
		2,
		5.0,
		"https://test.com",
		timestamp,
	)

	if dp == nil {
		t.Fatal("Expected data point but got nil")
	}

	if dp.Category != "Transportation" {
		t.Errorf("Expected category Transportation, got %s", dp.Category)
	}

	if dp.SubCategory != "Public Transport" {
		t.Errorf("Expected subcategory Public Transport, got %s", dp.SubCategory)
	}

	if dp.Price != 5.0 {
		t.Errorf("Expected price 5.0, got %f", dp.Price)
	}

	if dp.Unit != "AED" {
		t.Errorf("Expected unit AED, got %s", dp.Unit)
	}

	if dp.Source != "rta_official" {
		t.Errorf("Expected source rta_official, got %s", dp.Source)
	}

	if dp.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", dp.Confidence)
	}

	// Check attributes
	if mode, ok := dp.Attributes["transport_mode"]; !ok || mode != "metro" {
		t.Errorf("Expected transport_mode metro, got %v", mode)
	}

	if cardType, ok := dp.Attributes["card_type"]; !ok || cardType != "silver" {
		t.Errorf("Expected card_type silver, got %v", cardType)
	}

	if zones, ok := dp.Attributes["zones_crossed"]; !ok || zones != 2 {
		t.Errorf("Expected zones_crossed 2, got %v", zones)
	}

	// Check tags
	expectedTags := []string{"transport", "public_transport", "rta", "dubai", "metro", "silver"}
	if len(dp.Tags) < 5 {
		t.Errorf("Expected at least 5 tags, got %d", len(dp.Tags))
	}

	for _, expectedTag := range expectedTags {
		found := false
		for _, tag := range dp.Tags {
			if tag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag %s not found in %v", expectedTag, dp.Tags)
		}
	}
}
