package rta

import "fmt"

// TransportMode represents different public transport modes in Dubai
type TransportMode string

const (
	ModeMetro TransportMode = "metro"
	ModeBus   TransportMode = "bus"
	ModeTram  TransportMode = "tram"
	ModeTaxi  TransportMode = "taxi"
)

// CardType represents different Nol card types
type CardType string

const (
	CardSilver CardType = "silver" // Standard card
	CardGold   CardType = "gold"   // Premium/First class
	CardBlue   CardType = "blue"   // Student/Senior/People of determination
	CardRed    CardType = "red"    // Single-use ticket
)

// FareType represents different fare categories
type FareType string

const (
	FareSingleJourney FareType = "single_journey"
	FareDayPass       FareType = "day_pass"
	FareRegular       FareType = "regular"
)

// ZoneFare represents fare information for a specific number of zones
type ZoneFare struct {
	Zones      int             // Number of zones crossed (0 for regular/fixed fares)
	SilverCard float64         // Silver card fare (AED)
	GoldCard   float64         // Gold card fare (AED)
	BlueCard   float64         // Blue card fare (AED) - optional
	RedTicket  float64         // Red ticket fare (AED) - optional
	FareType   FareType        // Type of fare
	Mode       TransportMode   // Transport mode
}

// TaxiFare represents taxi fare information
type TaxiFare struct {
	FareType    string  // day_flag_down, night_flag_down, airport_pickup, per_km, waiting_time, minimum
	Amount      float64 // Fare amount in AED
	TimeRange   string  // e.g., "6 AM - 10 PM", "10 PM - 6 AM"
	Unit        string  // e.g., "AED", "AED/km", "AED/minute"
	Description string  // Human-readable description
}

// GetMetroFares returns all metro fare configurations
func GetMetroFares() []ZoneFare {
	return []ZoneFare{
		// Regular fare (default single journey)
		{
			Zones:      0,
			SilverCard: 4.0,
			GoldCard:   8.0,
			FareType:   FareRegular,
			Mode:       ModeMetro,
		},
		// 1 Zone
		{
			Zones:      1,
			SilverCard: 3.0,
			GoldCard:   6.0,
			FareType:   FareSingleJourney,
			Mode:       ModeMetro,
		},
		// 2 Zones
		{
			Zones:      2,
			SilverCard: 5.0,
			GoldCard:   10.0,
			FareType:   FareSingleJourney,
			Mode:       ModeMetro,
		},
		// All Zones
		{
			Zones:      7, // Dubai Metro has 7 zones
			SilverCard: 7.5,
			GoldCard:   15.0,
			FareType:   FareSingleJourney,
			Mode:       ModeMetro,
		},
		// Day Pass
		{
			Zones:      0,
			SilverCard: 20.0,
			GoldCard:   0.0, // Day pass fare is same for all classes
			FareType:   FareDayPass,
			Mode:       ModeMetro,
		},
	}
}

// GetBusFares returns all bus fare configurations
func GetBusFares() []ZoneFare {
	return []ZoneFare{
		// 1 Zone
		{
			Zones:      1,
			SilverCard: 3.0,
			FareType:   FareSingleJourney,
			Mode:       ModeBus,
		},
		// 2 Zones
		{
			Zones:      2,
			SilverCard: 5.0,
			FareType:   FareSingleJourney,
			Mode:       ModeBus,
		},
		// All Zones
		{
			Zones:      7,
			SilverCard: 7.5,
			FareType:   FareSingleJourney,
			Mode:       ModeBus,
		},
		// Day Pass
		{
			Zones:      0,
			SilverCard: 20.0,
			FareType:   FareDayPass,
			Mode:       ModeBus,
		},
	}
}

// GetTramFares returns all tram fare configurations
func GetTramFares() []ZoneFare {
	return []ZoneFare{
		// Single Journey
		{
			Zones:      0,
			SilverCard: 3.0,
			GoldCard:   6.0,
			FareType:   FareSingleJourney,
			Mode:       ModeTram,
		},
		// Day Pass Silver
		{
			Zones:      0,
			SilverCard: 15.0,
			FareType:   FareDayPass,
			Mode:       ModeTram,
		},
		// Day Pass Gold
		{
			Zones:      0,
			GoldCard:   30.0,
			FareType:   FareDayPass,
			Mode:       ModeTram,
		},
	}
}

// GetTaxiFares returns all taxi fare configurations
func GetTaxiFares() []TaxiFare {
	return []TaxiFare{
		{
			FareType:    "day_flag_down",
			Amount:      5.0,
			TimeRange:   "6 AM - 10 PM",
			Unit:        "AED",
			Description: "Day time flag down fare",
		},
		{
			FareType:    "night_flag_down",
			Amount:      5.5,
			TimeRange:   "10 PM - 6 AM",
			Unit:        "AED",
			Description: "Night time flag down fare",
		},
		{
			FareType:    "airport_pickup",
			Amount:      25.0,
			TimeRange:   "24 hours",
			Unit:        "AED",
			Description: "Airport pickup fare",
		},
		{
			FareType:    "per_km",
			Amount:      1.96,
			TimeRange:   "24 hours",
			Unit:        "AED/km",
			Description: "Per kilometer fare",
		},
		{
			FareType:    "waiting_time",
			Amount:      0.5,
			TimeRange:   "24 hours",
			Unit:        "AED/minute",
			Description: "Waiting time per minute",
		},
		{
			FareType:    "minimum_fare",
			Amount:      12.0,
			TimeRange:   "24 hours",
			Unit:        "AED",
			Description: "Minimum fare",
		},
	}
}

// GetCardTypeName returns human-readable name for card type
func GetCardTypeName(cardType CardType) string {
	switch cardType {
	case CardSilver:
		return "Silver Card (Standard)"
	case CardGold:
		return "Gold Card (Premium)"
	case CardBlue:
		return "Blue Card (Concession)"
	case CardRed:
		return "Red Ticket (Single Use)"
	default:
		return string(cardType)
	}
}

// GetTransportModeName returns human-readable name for transport mode
func GetTransportModeName(mode TransportMode) string {
	switch mode {
	case ModeMetro:
		return "Dubai Metro"
	case ModeBus:
		return "Dubai Bus"
	case ModeTram:
		return "Dubai Tram"
	case ModeTaxi:
		return "Dubai Taxi"
	default:
		return string(mode)
	}
}

// FormatItemName generates a standardized item name for RTA fares
func FormatItemName(mode TransportMode, fareType FareType, zones int, cardType CardType) string {
	modeName := GetTransportModeName(mode)
	cardName := GetCardTypeName(cardType)

	switch fareType {
	case FareDayPass:
		if cardType != "" {
			return fmt.Sprintf("%s Day Pass - %s", modeName, cardName)
		}
		return fmt.Sprintf("%s Day Pass", modeName)
	case FareRegular:
		return fmt.Sprintf("%s Regular Fare - %s", modeName, cardName)
	case FareSingleJourney:
		if zones == 7 {
			return fmt.Sprintf("%s All Zones - %s", modeName, cardName)
		} else if zones > 0 {
			return fmt.Sprintf("%s %d Zone - %s", modeName, zones, cardName)
		}
		return fmt.Sprintf("%s Single Journey - %s", modeName, cardName)
	default:
		return fmt.Sprintf("%s - %s", modeName, cardName)
	}
}

// CalculateZones calculates number of zones between two stations
// Note: This is a placeholder - actual zone calculation would require
// station-to-zone mapping data
func CalculateZones(fromStation, toStation string) int {
	// In a real implementation, this would lookup station zones
	// and calculate the difference
	// For now, return a default value
	return 1
}
