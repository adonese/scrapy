package sewa

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/internal/models"
)

// TariffType represents the type of customer
type TariffType string

const (
	TariffEmirati   TariffType = "emirati"
	TariffExpatriate TariffType = "expatriate"
)

// RateType represents the type of utility rate
type RateType string

const (
	RateElectricity RateType = "electricity"
	RateWater       RateType = "water"
	RateSewerage    RateType = "sewerage"
)

// parseSEWATariffs extracts all tariff data from SEWA HTML page
func parseSEWATariffs(doc *goquery.Document, sourceURL string) ([]*models.CostDataPoint, error) {
	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	// Parse electricity tariffs for Emirati customers
	emiratiElectricityRates := parseElectricityTariff(doc, TariffEmirati)
	for _, rate := range emiratiElectricityRates {
		dataPoints = append(dataPoints, createElectricityDataPoint(rate, TariffEmirati, sourceURL, now))
	}

	// Parse electricity tariffs for Expatriate customers
	expatElectricityRates := parseElectricityTariff(doc, TariffExpatriate)
	for _, rate := range expatElectricityRates {
		dataPoints = append(dataPoints, createElectricityDataPoint(rate, TariffExpatriate, sourceURL, now))
	}

	// Parse water tariffs
	waterRates := parseWaterTariff(doc)
	for _, rate := range waterRates {
		dataPoints = append(dataPoints, rate)
	}

	// Parse sewerage information from additional info
	sewerageRate := parseSewerageInfo(doc, sourceURL, now)
	if sewerageRate != nil {
		dataPoints = append(dataPoints, sewerageRate)
	}

	return dataPoints, nil
}

// ElectricityRate represents a single electricity tier
type ElectricityRate struct {
	MinConsumption int
	MaxConsumption int
	Rate           float64 // Rate in fils per kWh
}

// parseElectricityTariff extracts electricity tariff data for a specific customer type
func parseElectricityTariff(doc *goquery.Document, tariffType TariffType) []ElectricityRate {
	rates := []ElectricityRate{}

	// Find the correct section based on customer type
	var section *goquery.Selection
	if tariffType == TariffEmirati {
		section = doc.Find("section.electricity-tariff")
	} else {
		section = doc.Find("section.electricity-tariff-expat")
	}

	// Parse each row in the tariff table
	section.Find("table.tariff-table tbody tr").Each(func(i int, row *goquery.Selection) {
		consumptionText := strings.TrimSpace(row.Find("td").First().Text())
		rateText := strings.TrimSpace(row.Find("td").Last().Text())

		min, max := parseConsumptionRange(consumptionText)
		rate := parseRateValue(rateText)

		if rate > 0 {
			rates = append(rates, ElectricityRate{
				MinConsumption: min,
				MaxConsumption: max,
				Rate:           rate,
			})
		}
	})

	return rates
}

// parseConsumptionRange extracts min and max consumption from text like "1 - 3,000" or "Above 10,000"
func parseConsumptionRange(text string) (int, int) {
	text = strings.ReplaceAll(text, ",", "")
	text = strings.TrimSpace(text)

	// Handle "Above X" format
	if strings.HasPrefix(strings.ToLower(text), "above") {
		re := regexp.MustCompile(`\d+`)
		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			min, _ := strconv.Atoi(matches[0])
			return min, -1 // -1 indicates unlimited
		}
		return 0, 0
	}

	// Handle "X - Y" format
	re := regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		return min, max
	}

	return 0, 0
}

// parseRateValue extracts numeric rate from text like "14 fils" or "27.5 fils"
func parseRateValue(text string) float64 {
	// Remove "fils" and any other text
	text = strings.ReplaceAll(strings.ToLower(text), "fils", "")
	text = strings.TrimSpace(text)

	// Extract numeric value
	re := regexp.MustCompile(`\d+\.?\d*`)
	match := re.FindString(text)
	if match != "" {
		value, err := strconv.ParseFloat(match, 64)
		if err == nil {
			return value
		}
	}

	return 0
}

// createElectricityDataPoint creates a CostDataPoint for an electricity tier
func createElectricityDataPoint(rate ElectricityRate, tariffType TariffType, sourceURL string, recordedAt time.Time) *models.CostDataPoint {
	// Create item name based on consumption range
	var itemName string
	var tierDesc string

	if rate.MaxConsumption == -1 {
		tierDesc = fmt.Sprintf("Above %d kWh", rate.MinConsumption)
	} else {
		tierDesc = fmt.Sprintf("%d-%d kWh", rate.MinConsumption, rate.MaxConsumption)
	}

	customerType := "Emirati"
	if tariffType == TariffExpatriate {
		customerType = "Expatriate"
	}

	itemName = fmt.Sprintf("SEWA Electricity (%s) - %s", tierDesc, customerType)

	// Convert fils to AED (100 fils = 1 AED)
	priceInAED := rate.Rate / 100.0

	attributes := map[string]interface{}{
		"consumption_range_min": rate.MinConsumption,
		"customer_type":         strings.ToLower(customerType),
		"rate_type":             "tier",
		"unit":                  "fils_per_kwh",
		"rate_fils":             rate.Rate,
	}

	if rate.MaxConsumption != -1 {
		attributes["consumption_range_max"] = rate.MaxConsumption
	}

	return &models.CostDataPoint{
		Category:    "Utilities",
		SubCategory: "Electricity",
		ItemName:    itemName,
		Price:       priceInAED,
		Location: models.Location{
			Emirate: "Sharjah",
			City:    "Sharjah",
		},
		Source:      "sewa_official",
		SourceURL:   sourceURL,
		Confidence:  0.98, // Official source
		Unit:        "AED per kWh",
		RecordedAt:  recordedAt,
		ValidFrom:   recordedAt,
		SampleSize:  1,
		Tags:        []string{"utilities", "electricity", "sewa", "sharjah"},
		Attributes:  attributes,
	}
}

// parseWaterTariff extracts water tariff data
func parseWaterTariff(doc *goquery.Document) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	section := doc.Find("section.water-tariff")

	section.Find("table.tariff-table tbody tr").Each(func(i int, row *goquery.Selection) {
		categoryText := strings.TrimSpace(row.Find("td").First().Text())
		rateText := strings.TrimSpace(row.Find("td").Last().Text())

		// Determine customer type
		var customerType string
		if strings.Contains(strings.ToLower(categoryText), "emirati") {
			customerType = "emirati"
		} else if strings.Contains(strings.ToLower(categoryText), "expatriate") {
			customerType = "expatriate"
		} else {
			return // Skip if customer type not recognized
		}

		// Parse rate (e.g., "AED 8.00")
		rate := parseAEDValue(rateText)
		if rate == 0 {
			return
		}

		customerTypeDisplay := "Emirati"
		if customerType == "expatriate" {
			customerTypeDisplay = "Expatriate"
		}

		itemName := fmt.Sprintf("SEWA Water (%s) - per 1000 Gallons", customerTypeDisplay)

		dataPoints = append(dataPoints, &models.CostDataPoint{
			Category:    "Utilities",
			SubCategory: "Water",
			ItemName:    itemName,
			Price:       rate,
			Location: models.Location{
				Emirate: "Sharjah",
				City:    "Sharjah",
			},
			Source:      "sewa_official",
			SourceURL:   "", // Will be set by caller
			Confidence:  0.98,
			Unit:        "AED per 1000 gallons",
			RecordedAt:  now,
			ValidFrom:   now,
			SampleSize:  1,
			Tags:        []string{"utilities", "water", "sewa", "sharjah"},
			Attributes: map[string]interface{}{
				"customer_type": customerType,
				"unit_type":     "per_1000_gallons",
			},
		})
	})

	return dataPoints
}

// parseAEDValue extracts numeric AED value from text like "AED 8.00"
func parseAEDValue(text string) float64 {
	// Remove "AED" and any other text
	text = strings.ReplaceAll(strings.ToLower(text), "aed", "")
	text = strings.TrimSpace(text)

	// Extract numeric value
	re := regexp.MustCompile(`\d+\.?\d*`)
	match := re.FindString(text)
	if match != "" {
		value, err := strconv.ParseFloat(match, 64)
		if err == nil {
			return value
		}
	}

	return 0
}

// parseSewerageInfo extracts sewerage charge information from additional info section
func parseSewerageInfo(doc *goquery.Document, sourceURL string, recordedAt time.Time) *models.CostDataPoint {
	section := doc.Find("section.additional-info")

	// Look for sewerage charge information
	var seweragePercent float64
	section.Find("ul li").Each(func(i int, li *goquery.Selection) {
		text := strings.ToLower(li.Text())
		if strings.Contains(text, "sewerage") {
			// Extract percentage (e.g., "Sewerage charge: 50% of water consumption charge")
			re := regexp.MustCompile(`(\d+)%`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				seweragePercent = percent
			}
		}
	})

	if seweragePercent == 0 {
		return nil
	}

	// Create a data point representing the sewerage calculation method
	return &models.CostDataPoint{
		Category:    "Utilities",
		SubCategory: "Sewerage",
		ItemName:    "SEWA Sewerage Charge",
		Price:       seweragePercent / 100.0, // Store as decimal (0.50 for 50%)
		Location: models.Location{
			Emirate: "Sharjah",
			City:    "Sharjah",
		},
		Source:      "sewa_official",
		SourceURL:   sourceURL,
		Confidence:  0.98,
		Unit:        "percentage of water charge",
		RecordedAt:  recordedAt,
		ValidFrom:   recordedAt,
		SampleSize:  1,
		Tags:        []string{"utilities", "sewerage", "sewa", "sharjah"},
		Attributes: map[string]interface{}{
			"calculation_method": fmt.Sprintf("%.0f%% of water consumption charge", seweragePercent),
			"rate_type":          "percentage",
		},
	}
}
