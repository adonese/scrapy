package aadc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/internal/models"
)

// ElectricityRate represents a single electricity rate tier
type ElectricityRate struct {
	ConsumptionMin  int     // kWh per month (minimum, 0 for first tier)
	ConsumptionMax  int     // kWh per month (0 for unlimited/above)
	RateFils        float64 // Rate in fils per kWh
	CustomerType    string  // "national" or "expatriate"
	IsUnlimitedMax  bool    // True if this is an "above X" tier
}

// WaterRate represents water rates
type WaterRate struct {
	RateAED      float64 // Rate in AED per 1000 IG
	CustomerType string  // "national" or "expatriate"
	Unit         string  // Usually "1000_ig" (imperial gallons)
}

// parseElectricityRates extracts electricity rates from the HTML document
func parseElectricityRates(doc *goquery.Document) ([]ElectricityRate, error) {
	rates := []ElectricityRate{}

	// Find all electricity tariff sections
	doc.Find("section.electricity-residential, section[class*='electricity']").Each(func(_ int, section *goquery.Selection) {
		// Look for UAE Nationals section
		section.Find("h3").Each(func(_ int, heading *goquery.Selection) {
			headingText := strings.ToLower(strings.TrimSpace(heading.Text()))
			customerType := ""

			if strings.Contains(headingText, "national") {
				customerType = "national"
			} else if strings.Contains(headingText, "expatriate") || strings.Contains(headingText, "expat") {
				customerType = "expatriate"
			}

			if customerType == "" {
				return
			}

			// Find the table after this heading
			table := heading.NextAll().Filter("table").First()
			if table.Length() == 0 {
				return
			}

			// Parse rows
			table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
				consumption := strings.TrimSpace(row.Find("td").First().Text())
				rateText := strings.TrimSpace(row.Find("td").Last().Text())

				if consumption == "" || rateText == "" {
					return
				}

				rate := parseElectricityTier(consumption, rateText, customerType)
				if rate != nil {
					rates = append(rates, *rate)
				}
			})
		})
	})

	if len(rates) == 0 {
		return nil, fmt.Errorf("no electricity rates found in HTML")
	}

	return rates, nil
}

// parseElectricityTier parses a single electricity rate tier
func parseElectricityTier(consumption, rateText, customerType string) *ElectricityRate {
	// Parse rate (e.g., "6.7 fils", "5.8 fils")
	rateFils := parseFilsRate(rateText)
	if rateFils == 0 {
		return nil
	}

	// Parse consumption range
	// Examples: "Up to 30,000", "Above 30,000", "401 - 700", "Up to 400"
	consumption = strings.ToLower(consumption)
	consumption = strings.ReplaceAll(consumption, ",", "")

	rate := &ElectricityRate{
		RateFils:     rateFils,
		CustomerType: customerType,
	}

	// Pattern: "above X" or "above X kWh"
	abovePattern := regexp.MustCompile(`above\s+(\d+)`)
	if matches := abovePattern.FindStringSubmatch(consumption); len(matches) > 1 {
		min, _ := strconv.Atoi(matches[1])
		rate.ConsumptionMin = min + 1 // Above means starting from X+1
		rate.ConsumptionMax = 0
		rate.IsUnlimitedMax = true
		return rate
	}

	// Pattern: "up to X" or "up to X kWh"
	upToPattern := regexp.MustCompile(`up\s+to\s+(\d+)`)
	if matches := upToPattern.FindStringSubmatch(consumption); len(matches) > 1 {
		max, _ := strconv.Atoi(matches[1])
		rate.ConsumptionMin = 0
		rate.ConsumptionMax = max
		return rate
	}

	// Pattern: "X - Y" or "X-Y"
	rangePattern := regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)
	if matches := rangePattern.FindStringSubmatch(consumption); len(matches) > 2 {
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		rate.ConsumptionMin = min
		rate.ConsumptionMax = max
		return rate
	}

	return nil
}

// parseWaterRates extracts water rates from the HTML document
func parseWaterRates(doc *goquery.Document) ([]WaterRate, error) {
	rates := []WaterRate{}

	// Find water tariff sections
	doc.Find("section.water-residential, section[class*='water']").Each(func(_ int, section *goquery.Selection) {
		// Look for customer type headings
		section.Find("h3").Each(func(_ int, heading *goquery.Selection) {
			headingText := strings.ToLower(strings.TrimSpace(heading.Text()))
			customerType := ""

			if strings.Contains(headingText, "national") {
				customerType = "national"
			} else if strings.Contains(headingText, "expatriate") || strings.Contains(headingText, "expat") {
				customerType = "expatriate"
			}

			if customerType == "" {
				return
			}

			// Find the table after this heading
			table := heading.NextAll().Filter("table").First()
			if table.Length() == 0 {
				return
			}

			// Parse water rate
			table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
				rateText := strings.TrimSpace(row.Find("td").Last().Text())
				if rateText == "" {
					return
				}

				// Parse AED rate (e.g., "AED 2.09", "8.55")
				rateAED := parseAEDRate(rateText)
				if rateAED > 0 {
					rates = append(rates, WaterRate{
						RateAED:      rateAED,
						CustomerType: customerType,
						Unit:         "1000_ig", // 1000 Imperial Gallons
					})
				}
			})
		})
	})

	if len(rates) == 0 {
		return nil, fmt.Errorf("no water rates found in HTML")
	}

	return rates, nil
}

// parseFilsRate extracts fils value from text like "6.7 fils" or "6.7"
func parseFilsRate(text string) float64 {
	// Remove "fils" and clean up
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "fils", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.TrimSpace(text)

	// Extract number
	re := regexp.MustCompile(`(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		rate, _ := strconv.ParseFloat(matches[1], 64)
		return rate
	}

	return 0
}

// parseAEDRate extracts AED value from text like "AED 2.09" or "8.55"
func parseAEDRate(text string) float64 {
	// Remove "AED" and clean up
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "aed", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.TrimSpace(text)

	// Extract number
	re := regexp.MustCompile(`(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		rate, _ := strconv.ParseFloat(matches[1], 64)
		return rate
	}

	return 0
}

// convertToDataPoints converts parsed rates to CostDataPoint models
func convertElectricityToDataPoints(rates []ElectricityRate, sourceURL string) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	for _, rate := range rates {
		// Create descriptive name based on tier
		tierName := formatElectricityTierName(rate)
		itemName := fmt.Sprintf("AADC Electricity %s - %s", tierName, strings.Title(rate.CustomerType))

		// Convert fils to AED (100 fils = 1 AED)
		priceAED := rate.RateFils / 100.0

		// Create attributes map
		attributes := map[string]interface{}{
			"customer_type": rate.CustomerType,
			"rate_type":     "tiered",
			"unit":          "fils_per_kwh",
			"fils_rate":     rate.RateFils,
			"tier_min_kwh":  rate.ConsumptionMin,
		}

		if rate.IsUnlimitedMax {
			attributes["tier_max_kwh"] = "unlimited"
		} else {
			attributes["tier_max_kwh"] = rate.ConsumptionMax
		}

		// Create tags
		tags := []string{"electricity", "utility", "aadc", rate.CustomerType}

		dataPoint := &models.CostDataPoint{
			Category:    "Utilities",
			SubCategory: "Electricity",
			ItemName:    itemName,
			Price:       priceAED,
			Location: models.Location{
				Emirate: "Abu Dhabi",
				City:    "Abu Dhabi",
			},
			Source:      "aadc_official",
			SourceURL:   sourceURL,
			Confidence:  0.98, // Official source
			Unit:        "AED per kWh",
			RecordedAt:  now,
			ValidFrom:   now,
			SampleSize:  1,
			Tags:        tags,
			Attributes:  attributes,
		}

		dataPoints = append(dataPoints, dataPoint)
	}

	return dataPoints
}

// convertWaterToDataPoints converts water rates to CostDataPoint models
func convertWaterToDataPoints(rates []WaterRate, sourceURL string) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	for _, rate := range rates {
		itemName := fmt.Sprintf("AADC Water - %s", strings.Title(rate.CustomerType))

		// Create attributes map
		attributes := map[string]interface{}{
			"customer_type": rate.CustomerType,
			"rate_type":     "flat",
			"unit":          "aed_per_1000_ig",
			"volume_unit":   "imperial_gallons",
		}

		// Create tags
		tags := []string{"water", "utility", "aadc", rate.CustomerType}

		dataPoint := &models.CostDataPoint{
			Category:    "Utilities",
			SubCategory: "Water",
			ItemName:    itemName,
			Price:       rate.RateAED,
			Location: models.Location{
				Emirate: "Abu Dhabi",
				City:    "Abu Dhabi",
			},
			Source:      "aadc_official",
			SourceURL:   sourceURL,
			Confidence:  0.98, // Official source
			Unit:        "AED per 1000 IG",
			RecordedAt:  now,
			ValidFrom:   now,
			SampleSize:  1,
			Tags:        tags,
			Attributes:  attributes,
		}

		dataPoints = append(dataPoints, dataPoint)
	}

	return dataPoints
}

// formatElectricityTierName creates a readable tier name
func formatElectricityTierName(rate ElectricityRate) string {
	if rate.ConsumptionMin == 0 && rate.ConsumptionMax > 0 {
		return fmt.Sprintf("Tier Up to %d kWh", rate.ConsumptionMax)
	}

	if rate.IsUnlimitedMax {
		return fmt.Sprintf("Tier Above %d kWh", rate.ConsumptionMin-1)
	}

	return fmt.Sprintf("Tier %d-%d kWh", rate.ConsumptionMin, rate.ConsumptionMax)
}
