package dubizzle

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adonese/cost-of-living/internal/models"
)

// parsePrice extracts price from text like "AED 85,000" or "85000 Dhs" or "85,000"
func parsePrice(text string) float64 {
	if text == "" {
		return 0
	}

	// Clean up the text
	text = strings.TrimSpace(text)

	// Remove common suffixes and currency markers
	replacements := []string{
		"/year", "/month", "/yearly", "/monthly",
		"per year", "per month",
		"AED", "aed", "Dhs", "DHS", "dhs",
		"Dirhams", "dirhams",
	}
	for _, r := range replacements {
		text = strings.ReplaceAll(text, r, "")
	}

	// Extract numbers (including commas)
	re := regexp.MustCompile(`[\d,]+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return 0
	}

	// Take first number found and remove commas
	numStr := strings.ReplaceAll(matches[0], ",", "")
	price, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return price
}

// parseLocation extracts location from text like "Dubai Marina, Dubai" or "Business Bay - Dubai"
func parseLocation(text string) models.Location {
	if text == "" {
		return models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
		}
	}

	// Clean up the text
	text = strings.TrimSpace(text)

	// Split by comma, hyphen, or pipe
	var parts []string
	if strings.Contains(text, ",") {
		parts = strings.Split(text, ",")
	} else if strings.Contains(text, "-") {
		parts = strings.Split(text, "-")
	} else if strings.Contains(text, "|") {
		parts = strings.Split(text, "|")
	} else {
		parts = []string{text}
	}

	location := models.Location{
		Emirate: "Dubai", // Default
	}

	// Parse based on number of parts
	if len(parts) >= 2 {
		location.Area = strings.TrimSpace(parts[0])
		cityOrEmirate := strings.TrimSpace(parts[1])

		// Check if it's an emirate
		if isEmirate(cityOrEmirate) {
			location.Emirate = cityOrEmirate
			location.City = cityOrEmirate
		} else {
			location.City = cityOrEmirate
		}
	} else if len(parts) == 1 {
		areaText := strings.TrimSpace(parts[0])

		// Check if the single part is an emirate
		if isEmirate(areaText) {
			location.Emirate = areaText
			location.City = areaText
		} else {
			location.Area = areaText
			location.City = "Dubai"
		}
	}

	return location
}

// isEmirate checks if a string is a UAE emirate name
func isEmirate(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	emirates := []string{
		"dubai",
		"abu dhabi",
		"sharjah",
		"ajman",
		"ras al khaimah",
		"rak", // Common abbreviation
		"fujairah",
		"umm al quwain",
		"uaq", // Common abbreviation
	}

	for _, emirate := range emirates {
		if s == emirate {
			return true
		}
	}

	return false
}

// parseBedrooms extracts the number of bedrooms from text
func parseBedrooms(text string) string {
	if text == "" {
		return ""
	}

	text = strings.TrimSpace(text)
	textLower := strings.ToLower(text)

	// Check for studio
	if strings.Contains(textLower, "studio") {
		return "Studio"
	}

	// Extract numbers
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return ""
	}

	// Handle common patterns like "1BR", "2 Bed", "3 Bedroom"
	if len(matches) > 0 {
		return matches[0]
	}

	return ""
}

// parseBathrooms extracts the number of bathrooms from text
func parseBathrooms(text string) string {
	if text == "" {
		return ""
	}

	text = strings.TrimSpace(text)

	// Extract numbers
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return ""
	}

	return matches[0]
}

// parseArea extracts square footage from text like "1200 sqft" or "1,200 sq.ft."
func parseArea(text string) float64 {
	if text == "" {
		return 0
	}

	// Clean up the text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	// Remove common suffixes
	replacements := []string{
		"sqft", "sq.ft", "sq ft", "sq. ft.",
		"sqm", "sq.m", "sq m", "sq. m.",
	}
	for _, r := range replacements {
		text = strings.ReplaceAll(text, r, "")
	}

	// Extract numbers (including commas)
	re := regexp.MustCompile(`[\d,]+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return 0
	}

	// Take first number found and remove commas
	numStr := strings.ReplaceAll(matches[0], ",", "")
	area, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return area
}

// normalizeTitle cleans up property titles
func normalizeTitle(title string) string {
	title = strings.TrimSpace(title)

	// Remove excessive whitespace
	re := regexp.MustCompile(`\s+`)
	title = re.ReplaceAllString(title, " ")

	return title
}
