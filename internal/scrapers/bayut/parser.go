package bayut

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adonese/cost-of-living/internal/models"
)

// parsePrice extracts price from text like "AED 85,000/year" or "85000 AED"
func parsePrice(text string) float64 {
	if text == "" {
		return 0
	}

	// Clean up the text
	text = strings.TrimSpace(text)

	// Remove common suffixes
	text = strings.ReplaceAll(text, "/year", "")
	text = strings.ReplaceAll(text, "/month", "")
	text = strings.ReplaceAll(text, "per year", "")
	text = strings.ReplaceAll(text, "per month", "")
	text = strings.ReplaceAll(text, "AED", "")
	text = strings.ReplaceAll(text, "aed", "")

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

// parseLocation extracts location from text like "Dubai Marina, Dubai" or "Business Bay"
func parseLocation(text string) models.Location {
	if text == "" {
		return models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
		}
	}

	// Clean up the text
	text = strings.TrimSpace(text)

	// Split by comma or hyphen
	var parts []string
	if strings.Contains(text, ",") {
		parts = strings.Split(text, ",")
	} else if strings.Contains(text, "-") {
		parts = strings.Split(text, "-")
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
		location.Area = strings.TrimSpace(parts[0])
		location.City = "Dubai"
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
		"fujairah",
		"umm al quwain",
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

	// Extract numbers
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		// Check for "Studio"
		if strings.Contains(strings.ToLower(text), "studio") {
			return "Studio"
		}
		return ""
	}

	return matches[0]
}
