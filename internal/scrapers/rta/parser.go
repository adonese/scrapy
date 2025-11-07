package rta

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/internal/models"
)

// ParseFares extracts all fare information from RTA HTML document
func ParseFares(doc *goquery.Document, sourceURL string) ([]*models.CostDataPoint, error) {
	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	// Parse Metro fares
	metroFares := parseMetroFares(doc, sourceURL, now)
	dataPoints = append(dataPoints, metroFares...)

	// Parse Bus fares
	busFares := parseBusFares(doc, sourceURL, now)
	dataPoints = append(dataPoints, busFares...)

	// Parse Tram fares
	tramFares := parseTramFares(doc, sourceURL, now)
	dataPoints = append(dataPoints, tramFares...)

	// Parse Taxi fares
	taxiFares := parseTaxiFares(doc, sourceURL, now)
	dataPoints = append(dataPoints, taxiFares...)

	return dataPoints, nil
}

// parseMetroFares extracts metro fare information
func parseMetroFares(doc *goquery.Document, sourceURL string, timestamp time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Find the metro fares section
	metroSection := doc.Find("section.metro-fares, .metro-fares, #metro-fares")
	if metroSection.Length() == 0 {
		// Try to find by heading
		doc.Find("h1, h2, h3").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(strings.ToLower(s.Text()), "metro") {
				metroSection = s.Parent()
			}
		})
	}

	// Parse metro fare table
	metroSection.Find("table").Each(func(i int, table *goquery.Selection) {
		// Find table rows
		table.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
			fareData := parseMetroTableRow(row, sourceURL, timestamp)
			if fareData != nil {
				dataPoints = append(dataPoints, fareData...)
			}
		})
	})

	return dataPoints
}

// parseMetroTableRow parses a single row from the metro fare table
func parseMetroTableRow(row *goquery.Selection, sourceURL string, timestamp time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	cells := row.Find("td")
	if cells.Length() < 2 {
		return nil
	}

	// First cell is zone information
	zoneText := strings.TrimSpace(cells.Eq(0).Text())

	// Parse zone number or type
	zones := 0
	fareType := FareSingleJourney

	if strings.Contains(strings.ToLower(zoneText), "regular") {
		fareType = FareRegular
		zones = 0
	} else if strings.Contains(strings.ToLower(zoneText), "day pass") {
		fareType = FareDayPass
		zones = 0
	} else if strings.Contains(strings.ToLower(zoneText), "all") {
		zones = 7 // All zones
	} else {
		// Extract number from zone text (e.g., "1 Zone" -> 1)
		zones = extractNumber(zoneText)
	}

	// Parse Silver card fare (always present)
	if cells.Length() >= 2 {
		silverPrice := parsePrice(cells.Eq(1).Text())
		if silverPrice > 0 {
			dataPoints = append(dataPoints, createFareDataPoint(
				ModeMetro, CardSilver, fareType, zones, silverPrice, sourceURL, timestamp,
			))
		}
	}

	// Parse Gold card fare (if present)
	if cells.Length() >= 3 {
		goldPrice := parsePrice(cells.Eq(2).Text())
		if goldPrice > 0 {
			dataPoints = append(dataPoints, createFareDataPoint(
				ModeMetro, CardGold, fareType, zones, goldPrice, sourceURL, timestamp,
			))
		}
	}

	// Parse Day Pass (if present and different from card-specific day passes)
	if cells.Length() >= 4 && fareType != FareDayPass {
		dayPassPrice := parsePrice(cells.Eq(3).Text())
		if dayPassPrice > 0 {
			dataPoints = append(dataPoints, createFareDataPoint(
				ModeMetro, "", FareDayPass, 0, dayPassPrice, sourceURL, timestamp,
			))
		}
	}

	return dataPoints
}

// parseBusFares extracts bus fare information
func parseBusFares(doc *goquery.Document, sourceURL string, timestamp time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Find the bus fares section
	busSection := doc.Find("section.bus-fares, .bus-fares, #bus-fares")
	if busSection.Length() == 0 {
		doc.Find("h1, h2, h3").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(strings.ToLower(s.Text()), "bus") {
				busSection = s.Parent()
			}
		})
	}

	// Parse bus fare table
	busSection.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}

		journeyType := strings.TrimSpace(cells.Eq(0).Text())
		priceText := strings.TrimSpace(cells.Eq(1).Text())
		price := parsePrice(priceText)

		if price <= 0 {
			return
		}

		// Determine fare type and zones
		fareType := FareSingleJourney
		zones := 0

		if strings.Contains(strings.ToLower(journeyType), "day pass") {
			fareType = FareDayPass
		} else if strings.Contains(strings.ToLower(journeyType), "all") {
			zones = 7
		} else {
			zones = extractNumber(journeyType)
		}

		dataPoints = append(dataPoints, createFareDataPoint(
			ModeBus, CardSilver, fareType, zones, price, sourceURL, timestamp,
		))
	})

	return dataPoints
}

// parseTramFares extracts tram fare information
func parseTramFares(doc *goquery.Document, sourceURL string, timestamp time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Find the tram fares section
	tramSection := doc.Find("section.tram-fares, .tram-fares, #tram-fares")
	if tramSection.Length() == 0 {
		doc.Find("h1, h2, h3").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(strings.ToLower(s.Text()), "tram") {
				tramSection = s.Parent()
			}
		})
	}

	// Parse tram fare table
	tramSection.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}

		ticketType := strings.TrimSpace(cells.Eq(0).Text())

		fareType := FareSingleJourney
		if strings.Contains(strings.ToLower(ticketType), "day pass") {
			fareType = FareDayPass
		}

		// Parse Silver fare
		if cells.Length() >= 2 {
			silverPrice := parsePrice(cells.Eq(1).Text())
			if silverPrice > 0 {
				dataPoints = append(dataPoints, createFareDataPoint(
					ModeTram, CardSilver, fareType, 0, silverPrice, sourceURL, timestamp,
				))
			}
		}

		// Parse Gold fare
		if cells.Length() >= 3 {
			goldPrice := parsePrice(cells.Eq(2).Text())
			if goldPrice > 0 {
				dataPoints = append(dataPoints, createFareDataPoint(
					ModeTram, CardGold, fareType, 0, goldPrice, sourceURL, timestamp,
				))
			}
		}
	})

	return dataPoints
}

// parseTaxiFares extracts taxi fare information
func parseTaxiFares(doc *goquery.Document, sourceURL string, timestamp time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Find the taxi info section
	taxiSection := doc.Find("section.taxi-info, .taxi-info, #taxi-info, section.taxi-fares, .taxi-fares")
	if taxiSection.Length() == 0 {
		doc.Find("h1, h2, h3").Each(func(i int, s *goquery.Selection) {
			text := strings.ToLower(s.Text())
			if strings.Contains(text, "taxi") {
				taxiSection = s.Parent()
			}
		})
	}

	// Parse taxi fare table
	taxiSection.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}

		fareType := strings.TrimSpace(cells.Eq(0).Text())
		priceText := strings.TrimSpace(cells.Eq(1).Text())
		price := parsePrice(priceText)

		if price <= 0 {
			return
		}

		// Create taxi fare data point
		itemName := fmt.Sprintf("Dubai Taxi - %s", fareType)

		// Determine unit based on fare type
		unit := "AED"
		if strings.Contains(strings.ToLower(fareType), "kilometer") ||
		   strings.Contains(strings.ToLower(fareType), "km") {
			unit = "AED/km"
		} else if strings.Contains(strings.ToLower(fareType), "minute") ||
		          strings.Contains(strings.ToLower(fareType), "waiting") {
			unit = "AED/minute"
		}

		dataPoint := &models.CostDataPoint{
			Category:    "Transportation",
			SubCategory: "Taxi",
			ItemName:    itemName,
			Price:       price,
			Location: models.Location{
				Emirate: "Dubai",
				City:    "Dubai",
			},
			Source:     "rta_official",
			SourceURL:  sourceURL,
			Confidence: 0.95,
			Unit:       unit,
			RecordedAt: timestamp,
			ValidFrom:  timestamp,
			SampleSize: 1,
			Tags:       []string{"transport", "taxi", "rta", "dubai"},
			Attributes: map[string]interface{}{
				"transport_mode": "taxi",
				"fare_type":      normalizeFareTypeName(fareType),
			},
		}

		dataPoints = append(dataPoints, dataPoint)
	})

	return dataPoints
}

// createFareDataPoint creates a standardized CostDataPoint for RTA fares
func createFareDataPoint(mode TransportMode, cardType CardType, fareType FareType,
	zones int, price float64, sourceURL string, timestamp time.Time) *models.CostDataPoint {

	itemName := FormatItemName(mode, fareType, zones, cardType)

	// Build tags
	tags := []string{"transport", "public_transport", "rta", "dubai", string(mode)}
	if cardType != "" {
		tags = append(tags, string(cardType))
	}
	if fareType == FareDayPass {
		tags = append(tags, "day_pass")
	}

	// Build attributes
	attributes := map[string]interface{}{
		"transport_mode": string(mode),
		"fare_type":      string(fareType),
	}

	if cardType != "" {
		attributes["card_type"] = string(cardType)
	}

	if zones > 0 {
		attributes["zones_crossed"] = zones
	}

	// Determine subcategory
	subCategory := "Public Transport"
	if mode == ModeTaxi {
		subCategory = "Taxi"
	}

	return &models.CostDataPoint{
		Category:    "Transportation",
		SubCategory: subCategory,
		ItemName:    itemName,
		Price:       price,
		Location: models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
		},
		Source:     "rta_official",
		SourceURL:  sourceURL,
		Confidence: 0.95,
		Unit:       "AED",
		RecordedAt: timestamp,
		ValidFrom:  timestamp,
		SampleSize: 1,
		Tags:       tags,
		Attributes: attributes,
	}
}

// parsePrice extracts price from text (e.g., "AED 3.50" -> 3.50)
func parsePrice(text string) float64 {
	// Remove common currency symbols and text
	text = strings.ReplaceAll(text, "AED", "")
	text = strings.ReplaceAll(text, "Dhs", "")
	text = strings.ReplaceAll(text, "Dh", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.TrimSpace(text)

	// Handle dashes or empty values
	if text == "" || text == "-" || text == "N/A" {
		return 0
	}

	// Extract first number found
	re := regexp.MustCompile(`\d+\.?\d*`)
	matches := re.FindString(text)
	if matches == "" {
		return 0
	}

	price, err := strconv.ParseFloat(matches, 64)
	if err != nil {
		return 0
	}

	return price
}

// extractNumber extracts the first number from a string
func extractNumber(text string) int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(text)
	if matches == "" {
		return 0
	}

	num, err := strconv.Atoi(matches)
	if err != nil {
		return 0
	}

	return num
}

// normalizeFareTypeName normalizes fare type names for consistency
func normalizeFareTypeName(fareType string) string {
	fareType = strings.ToLower(fareType)
	fareType = strings.ReplaceAll(fareType, " ", "_")
	fareType = strings.ReplaceAll(fareType, "-", "_")
	fareType = strings.ReplaceAll(fareType, "(", "")
	fareType = strings.ReplaceAll(fareType, ")", "")
	return fareType
}
