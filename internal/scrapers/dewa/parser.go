package dewa

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/internal/models"
)

// RateSlab represents a consumption slab with its rate
type RateSlab struct {
	SlabName string
	MinRange int
	MaxRange int // -1 for "Above X" slabs
	Rate     float64
	Unit     string
}

// parseElectricitySlabs extracts electricity rate slabs from HTML
func parseElectricitySlabs(doc *goquery.Document) ([]RateSlab, error) {
	var slabs []RateSlab

	// Find electricity tariff section
	electricitySection := doc.Find("section.electricity-tariff, div.electricity-tariff")
	if electricitySection.Length() == 0 {
		// Try alternative selector
		electricitySection = doc.Find("h2:contains('Electricity')").Parent()
	}

	// Parse table rows
	electricitySection.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		slabText := row.Find("td").First().Text()
		rateText := row.Find("td").Last().Text()

		slab := parseRateSlab(slabText, rateText, "electricity")
		if slab.Rate > 0 {
			slabs = append(slabs, slab)
		}
	})

	if len(slabs) == 0 {
		return nil, fmt.Errorf("no electricity slabs found")
	}

	return slabs, nil
}

// parseWaterSlabs extracts water rate slabs from HTML
func parseWaterSlabs(doc *goquery.Document) ([]RateSlab, error) {
	var slabs []RateSlab

	// Find water tariff section
	waterSection := doc.Find("section.water-tariff, div.water-tariff")
	if waterSection.Length() == 0 {
		// Try alternative selector
		waterSection = doc.Find("h2:contains('Water')").Parent()
	}

	// Parse table rows
	waterSection.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		slabText := row.Find("td").First().Text()
		rateText := row.Find("td").Last().Text()

		slab := parseRateSlab(slabText, rateText, "water")
		if slab.Rate > 0 {
			slabs = append(slabs, slab)
		}
	})

	if len(slabs) == 0 {
		return nil, fmt.Errorf("no water slabs found")
	}

	return slabs, nil
}

// parseFuelSurcharge extracts fuel surcharge information
func parseFuelSurcharge(doc *goquery.Document) *RateSlab {
	// Look for fuel surcharge note
	fuelText := ""
	doc.Find("p.note, div.note, p:contains('Fuel'), p:contains('fuel')").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(strings.ToLower(text), "fuel") && strings.Contains(strings.ToLower(text), "surcharge") {
			fuelText = text
		}
	})

	if fuelText == "" {
		return nil
	}

	// Extract rate from text like "Fuel Surcharge: Variable (currently 6.5 fils/kWh)"
	rate := extractRate(fuelText)
	if rate == 0 {
		return nil
	}

	return &RateSlab{
		SlabName: "Fuel Surcharge",
		MinRange: 0,
		MaxRange: -1, // Applies to all consumption
		Rate:     rate,
		Unit:     "fils_per_kwh",
	}
}

// parseRateSlab parses a single rate slab from text
func parseRateSlab(slabText, rateText, utilityType string) RateSlab {
	slab := RateSlab{}

	// Parse consumption range from text like "0 - 2,000" or "Above 6,000"
	min, max := parseConsumptionRange(slabText)
	slab.MinRange = min
	slab.MaxRange = max

	// Parse rate from text like "23.0 fils" or "3.57 fils"
	rate := extractRate(rateText)
	slab.Rate = rate

	// Set unit based on utility type
	if utilityType == "electricity" {
		slab.Unit = "fils_per_kwh"
		if max == -1 {
			slab.SlabName = fmt.Sprintf("Slab %d+ kWh", min)
		} else {
			slab.SlabName = fmt.Sprintf("Slab %d-%d kWh", min, max)
		}
	} else {
		slab.Unit = "fils_per_ig"
		if max == -1 {
			slab.SlabName = fmt.Sprintf("Slab %d+ IG", min)
		} else {
			slab.SlabName = fmt.Sprintf("Slab %d-%d IG", min, max)
		}
	}

	return slab
}

// parseConsumptionRange extracts min and max consumption from text
func parseConsumptionRange(text string) (int, int) {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "")

	// Check for "Above X" or "Over X" pattern
	if strings.Contains(strings.ToLower(text), "above") || strings.Contains(strings.ToLower(text), "over") {
		re := regexp.MustCompile(`(\d+)`)
		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			min, _ := strconv.Atoi(matches[0])
			return min + 1, -1 // -1 indicates no upper limit
		}
	}

	// Parse range like "0 - 2000" or "2001-4000"
	re := regexp.MustCompile(`(\d+)\s*[-–]\s*(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 3 {
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		return min, max
	}

	return 0, 0
}

// extractRate extracts the rate value from text
func extractRate(text string) float64 {
	if text == "" {
		return 0
	}

	text = strings.TrimSpace(text)

	// Extract numbers with decimal points
	re := regexp.MustCompile(`(\d+\.?\d*)`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return 0
	}

	// Take first number found
	rate, err := strconv.ParseFloat(matches[0], 64)
	if err != nil {
		return 0
	}

	return rate
}

// slabToDataPoint converts a RateSlab to a CostDataPoint
func slabToDataPoint(slab RateSlab, utilityType string, sourceURL string) *models.CostDataPoint {
	now := time.Now()

	// Convert fils to AED (100 fils = 1 AED)
	priceInAED := slab.Rate / 100.0

	category := "Utilities"
	subCategory := ""
	if utilityType == "electricity" {
		subCategory = "Electricity"
	} else if utilityType == "water" {
		subCategory = "Water"
	} else {
		subCategory = "Fuel Surcharge"
	}

	// Build item name
	itemName := fmt.Sprintf("DEWA %s %s", subCategory, slab.SlabName)
	if slab.SlabName == "Fuel Surcharge" {
		itemName = "DEWA Fuel Surcharge"
	}

	attributes := map[string]interface{}{
		"rate_type":              "slab",
		"unit":                   slab.Unit,
		"consumption_range_min":  slab.MinRange,
	}

	// Only add max range if it's not -1 (unlimited)
	if slab.MaxRange > 0 {
		attributes["consumption_range_max"] = slab.MaxRange
	}

	return &models.CostDataPoint{
		Category:    category,
		SubCategory: subCategory,
		ItemName:    itemName,
		Price:       priceInAED,
		Location: models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
		},
		Source:      "dewa_official",
		SourceURL:   sourceURL,
		Confidence:  0.98, // Official source
		Unit:        "AED",
		RecordedAt:  now,
		ValidFrom:   now,
		SampleSize:  1,
		Tags:        []string{"utility", "official", utilityType},
		Attributes:  attributes,
	}
}
