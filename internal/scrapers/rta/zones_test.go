package rta

import (
	"testing"
)

func TestGetMetroFares(t *testing.T) {
	fares := GetMetroFares()

	if len(fares) == 0 {
		t.Fatal("Expected metro fares but got none")
	}

	// Should have Regular, 1 Zone, 2 Zones, All Zones, and Day Pass
	if len(fares) < 5 {
		t.Errorf("Expected at least 5 metro fares, got %d", len(fares))
	}

	// Validate each fare
	for i, fare := range fares {
		if fare.Mode != ModeMetro {
			t.Errorf("Fare %d: expected mode %s, got %s", i, ModeMetro, fare.Mode)
		}

		// Check prices are positive (except for optional fields)
		if fare.FareType != FareDayPass && fare.SilverCard <= 0 {
			t.Errorf("Fare %d: invalid silver card price %f", i, fare.SilverCard)
		}

		// Gold card should be approximately double silver (for metro)
		if fare.FareType != FareDayPass && fare.SilverCard > 0 && fare.GoldCard > 0 {
			ratio := fare.GoldCard / fare.SilverCard
			if ratio < 1.8 || ratio > 2.2 {
				t.Errorf("Fare %d: unexpected gold/silver ratio %.2f (gold=%f, silver=%f)",
					i, ratio, fare.GoldCard, fare.SilverCard)
			}
		}
	}
}

func TestGetBusFares(t *testing.T) {
	fares := GetBusFares()

	if len(fares) == 0 {
		t.Fatal("Expected bus fares but got none")
	}

	// Should have 1 Zone, 2 Zones, All Zones, and Day Pass
	if len(fares) < 4 {
		t.Errorf("Expected at least 4 bus fares, got %d", len(fares))
	}

	// Validate each fare
	for i, fare := range fares {
		if fare.Mode != ModeBus {
			t.Errorf("Fare %d: expected mode %s, got %s", i, ModeBus, fare.Mode)
		}

		if fare.SilverCard <= 0 {
			t.Errorf("Fare %d: invalid silver card price %f", i, fare.SilverCard)
		}

		// Bus typically only has silver card pricing
		if fare.GoldCard > 0 && fare.FareType != FareDayPass {
			t.Logf("Note: Fare %d has gold card pricing which is unusual for buses", i)
		}
	}
}

func TestGetTramFares(t *testing.T) {
	fares := GetTramFares()

	if len(fares) == 0 {
		t.Fatal("Expected tram fares but got none")
	}

	// Should have at least single journey and day pass options
	if len(fares) < 2 {
		t.Errorf("Expected at least 2 tram fares, got %d", len(fares))
	}

	// Validate each fare
	for i, fare := range fares {
		if fare.Mode != ModeTram {
			t.Errorf("Fare %d: expected mode %s, got %s", i, ModeTram, fare.Mode)
		}

		// Either silver or gold should be set
		if fare.SilverCard <= 0 && fare.GoldCard <= 0 {
			t.Errorf("Fare %d: neither silver nor gold card price is set", i)
		}
	}
}

func TestGetTaxiFares(t *testing.T) {
	fares := GetTaxiFares()

	if len(fares) == 0 {
		t.Fatal("Expected taxi fares but got none")
	}

	// Should have flag down, per km, waiting time, minimum fare
	if len(fares) < 4 {
		t.Errorf("Expected at least 4 taxi fares, got %d", len(fares))
	}

	expectedTypes := []string{"day_flag_down", "night_flag_down", "per_km", "minimum_fare"}
	foundTypes := make(map[string]bool)

	for _, fare := range fares {
		foundTypes[fare.FareType] = true

		if fare.Amount <= 0 {
			t.Errorf("Fare %s: invalid amount %f", fare.FareType, fare.Amount)
		}

		if fare.Unit == "" {
			t.Errorf("Fare %s: missing unit", fare.FareType)
		}

		if fare.Description == "" {
			t.Errorf("Fare %s: missing description", fare.FareType)
		}
	}

	for _, expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Missing expected taxi fare type: %s", expectedType)
		}
	}
}

func TestGetCardTypeName(t *testing.T) {
	tests := []struct {
		cardType CardType
		expected string
	}{
		{CardSilver, "Silver Card (Standard)"},
		{CardGold, "Gold Card (Premium)"},
		{CardBlue, "Blue Card (Concession)"},
		{CardRed, "Red Ticket (Single Use)"},
		{CardType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.cardType), func(t *testing.T) {
			result := GetCardTypeName(tt.cardType)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetTransportModeName(t *testing.T) {
	tests := []struct {
		mode     TransportMode
		expected string
	}{
		{ModeMetro, "Dubai Metro"},
		{ModeBus, "Dubai Bus"},
		{ModeTram, "Dubai Tram"},
		{ModeTaxi, "Dubai Taxi"},
		{TransportMode("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := GetTransportModeName(tt.mode)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatItemName(t *testing.T) {
	tests := []struct {
		name     string
		mode     TransportMode
		fareType FareType
		zones    int
		cardType CardType
		expected string
	}{
		{
			name:     "Metro 1 Zone Silver",
			mode:     ModeMetro,
			fareType: FareSingleJourney,
			zones:    1,
			cardType: CardSilver,
			expected: "Dubai Metro 1 Zone - Silver Card (Standard)",
		},
		{
			name:     "Metro All Zones Gold",
			mode:     ModeMetro,
			fareType: FareSingleJourney,
			zones:    7,
			cardType: CardGold,
			expected: "Dubai Metro All Zones - Gold Card (Premium)",
		},
		{
			name:     "Metro Day Pass",
			mode:     ModeMetro,
			fareType: FareDayPass,
			zones:    0,
			cardType: "",
			expected: "Dubai Metro Day Pass",
		},
		{
			name:     "Bus 2 Zones",
			mode:     ModeBus,
			fareType: FareSingleJourney,
			zones:    2,
			cardType: CardSilver,
			expected: "Dubai Bus 2 Zone - Silver Card (Standard)",
		},
		{
			name:     "Tram Single Journey Gold",
			mode:     ModeTram,
			fareType: FareSingleJourney,
			zones:    0,
			cardType: CardGold,
			expected: "Dubai Tram Single Journey - Gold Card (Premium)",
		},
		{
			name:     "Metro Regular Fare",
			mode:     ModeMetro,
			fareType: FareRegular,
			zones:    0,
			cardType: CardSilver,
			expected: "Dubai Metro Regular Fare - Silver Card (Standard)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatItemName(tt.mode, tt.fareType, tt.zones, tt.cardType)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestZoneFareStructure(t *testing.T) {
	// Test that zone fare structure is properly defined
	fare := ZoneFare{
		Zones:      2,
		SilverCard: 5.0,
		GoldCard:   10.0,
		BlueCard:   2.5,
		RedTicket:  7.0,
		FareType:   FareSingleJourney,
		Mode:       ModeMetro,
	}

	if fare.Zones != 2 {
		t.Errorf("Expected 2 zones, got %d", fare.Zones)
	}

	if fare.SilverCard != 5.0 {
		t.Errorf("Expected silver card 5.0, got %f", fare.SilverCard)
	}

	if fare.Mode != ModeMetro {
		t.Errorf("Expected mode metro, got %s", fare.Mode)
	}
}

func TestTaxiFareStructure(t *testing.T) {
	// Test that taxi fare structure is properly defined
	fare := TaxiFare{
		FareType:    "day_flag_down",
		Amount:      5.0,
		TimeRange:   "6 AM - 10 PM",
		Unit:        "AED",
		Description: "Day time starting fare",
	}

	if fare.FareType != "day_flag_down" {
		t.Errorf("Expected fare type 'day_flag_down', got '%s'", fare.FareType)
	}

	if fare.Amount != 5.0 {
		t.Errorf("Expected amount 5.0, got %f", fare.Amount)
	}

	if fare.Unit != "AED" {
		t.Errorf("Expected unit 'AED', got '%s'", fare.Unit)
	}
}

func TestCalculateZones(t *testing.T) {
	// Test the zone calculation function
	// Note: This is currently a placeholder implementation
	zones := CalculateZones("Dubai Mall", "Burj Khalifa")

	if zones < 0 {
		t.Errorf("Expected non-negative zones, got %d", zones)
	}

	// For now, it should return 1 as default
	if zones != 1 {
		t.Logf("Note: CalculateZones returned %d, expected 1 (placeholder implementation)", zones)
	}
}

func TestFareConsistency(t *testing.T) {
	// Verify that metro fares increase with zones
	metroFares := GetMetroFares()

	var zone1Silver, zone2Silver, allZonesSilver float64

	for _, fare := range metroFares {
		if fare.FareType == FareSingleJourney {
			switch fare.Zones {
			case 1:
				zone1Silver = fare.SilverCard
			case 2:
				zone2Silver = fare.SilverCard
			case 7:
				allZonesSilver = fare.SilverCard
			}
		}
	}

	if zone1Silver >= zone2Silver {
		t.Errorf("Expected zone1 (%f) < zone2 (%f)", zone1Silver, zone2Silver)
	}

	if zone2Silver >= allZonesSilver {
		t.Errorf("Expected zone2 (%f) < all zones (%f)", zone2Silver, allZonesSilver)
	}
}
