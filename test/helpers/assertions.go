package helpers

import (
	"fmt"
	"testing"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/stretchr/testify/assert"
)

// AssertCostDataPoint performs comprehensive validation on a CostDataPoint
func AssertCostDataPoint(t *testing.T, cdp *models.CostDataPoint, expectations CostDataPointExpectations) {
	t.Helper()

	if cdp == nil {
		t.Fatal("CostDataPoint is nil")
		return
	}

	// Required fields
	assert.NotEmpty(t, cdp.Category, "Category should not be empty")
	assert.NotEmpty(t, cdp.ItemName, "ItemName should not be empty")
	assert.NotEmpty(t, cdp.Source, "Source should not be empty")
	assert.NotEmpty(t, cdp.Unit, "Unit should not be empty")

	// Price validation
	if expectations.MinPrice > 0 {
		assert.GreaterOrEqual(t, cdp.Price, expectations.MinPrice,
			"Price should be >= %v", expectations.MinPrice)
	}
	if expectations.MaxPrice > 0 {
		assert.LessOrEqual(t, cdp.Price, expectations.MaxPrice,
			"Price should be <= %v", expectations.MaxPrice)
	}

	// Category validation
	if expectations.Category != "" {
		assert.Equal(t, expectations.Category, cdp.Category,
			"Category mismatch")
	}

	// SubCategory validation
	if expectations.SubCategory != "" {
		assert.Equal(t, expectations.SubCategory, cdp.SubCategory,
			"SubCategory mismatch")
	}

	// Source validation
	if expectations.Source != "" {
		assert.Equal(t, expectations.Source, cdp.Source,
			"Source mismatch")
	}

	// Location validation
	if expectations.Emirate != "" {
		assert.Equal(t, expectations.Emirate, cdp.Location.Emirate,
			"Emirate mismatch")
	}

	// Confidence validation
	assert.Greater(t, cdp.Confidence, float32(0), "Confidence should be > 0")
	assert.LessOrEqual(t, cdp.Confidence, float32(1), "Confidence should be <= 1")

	// Timestamps
	assert.False(t, cdp.RecordedAt.IsZero(), "RecordedAt should be set")
	assert.False(t, cdp.ValidFrom.IsZero(), "ValidFrom should be set")
}

// CostDataPointExpectations defines expected values for validation
type CostDataPointExpectations struct {
	MinPrice    float64
	MaxPrice    float64
	Category    string
	SubCategory string
	Source      string
	Emirate     string
}

// AssertHousingDataPoint validates a housing-specific data point
func AssertHousingDataPoint(t *testing.T, cdp *models.CostDataPoint) {
	t.Helper()

	AssertCostDataPoint(t, cdp, CostDataPointExpectations{
		Category: "Housing",
		MinPrice: 0,
	})

	// Housing-specific validations
	assert.NotEmpty(t, cdp.Location.Emirate, "Location emirate should not be empty for housing")

	// Check for common housing attributes
	// Note: Some housing types (bedspace, roomspace) may not have bedroom counts
	if cdp.Attributes != nil {
		if bedrooms, ok := cdp.Attributes["bedrooms"]; ok {
			bedroomsStr, isString := bedrooms.(string)
			// Only assert if the attribute exists AND has a non-empty value
			// For shared accommodation, bedrooms might be empty
			if isString && cdp.SubCategory != "Shared Accommodation" {
				assert.NotEmpty(t, bedroomsStr, "Bedrooms attribute should not be empty for full apartments")
			}
		}
	}
}

// AssertDataPointCount validates the number of scraped data points
func AssertDataPointCount(t *testing.T, dataPoints []*models.CostDataPoint, min, max int) {
	t.Helper()

	count := len(dataPoints)
	if min > 0 {
		assert.GreaterOrEqual(t, count, min,
			"Expected at least %d data points, got %d", min, count)
	}
	if max > 0 {
		assert.LessOrEqual(t, count, max,
			"Expected at most %d data points, got %d", max, count)
	}
}

// AssertAllDataPointsValid validates that all data points in a slice are valid
func AssertAllDataPointsValid(t *testing.T, dataPoints []*models.CostDataPoint) {
	t.Helper()

	for i, cdp := range dataPoints {
		t.Run(fmt.Sprintf("DataPoint_%d", i), func(t *testing.T) {
			AssertCostDataPoint(t, cdp, CostDataPointExpectations{})
		})
	}
}

// AssertNoDuplicates checks that there are no duplicate entries based on ItemName and Price
func AssertNoDuplicates(t *testing.T, dataPoints []*models.CostDataPoint) {
	t.Helper()

	seen := make(map[string]bool)
	duplicates := []string{}

	for _, cdp := range dataPoints {
		key := fmt.Sprintf("%s_%v", cdp.ItemName, cdp.Price)
		if seen[key] {
			duplicates = append(duplicates, key)
		}
		seen[key] = true
	}

	assert.Empty(t, duplicates, "Found duplicate data points: %v", duplicates)
}

// AssertLocationValid validates location data
func AssertLocationValid(t *testing.T, location models.Location) {
	t.Helper()

	assert.NotEmpty(t, location.Emirate, "Emirate should not be empty")

	// Validate emirate is one of the UAE emirates
	validEmirates := []string{
		"Dubai", "Abu Dhabi", "Sharjah", "Ajman",
		"Ras Al Khaimah", "RAK", "Fujairah", "Umm Al Quwain", "UAQ",
	}

	isValid := false
	for _, valid := range validEmirates {
		if location.Emirate == valid {
			isValid = true
			break
		}
	}

	assert.True(t, isValid, "Emirate '%s' is not a valid UAE emirate", location.Emirate)
}

// AssertPriceInRange validates price is within reasonable bounds for housing
func AssertPriceInRange(t *testing.T, price float64, priceType string) {
	t.Helper()

	switch priceType {
	case "yearly_rent":
		// Yearly rent typically between 10k and 1M AED
		assert.GreaterOrEqual(t, price, 10000.0, "Yearly rent too low")
		assert.LessOrEqual(t, price, 1000000.0, "Yearly rent too high")

	case "monthly_rent":
		// Monthly rent typically between 500 and 100k AED
		assert.GreaterOrEqual(t, price, 500.0, "Monthly rent too low")
		assert.LessOrEqual(t, price, 100000.0, "Monthly rent too high")

	case "bedspace":
		// Bedspace typically between 300 and 1500 AED per month
		assert.GreaterOrEqual(t, price, 300.0, "Bedspace rent too low")
		assert.LessOrEqual(t, price, 1500.0, "Bedspace rent too high")

	case "roomspace":
		// Roomspace typically between 600 and 3000 AED per month
		assert.GreaterOrEqual(t, price, 600.0, "Roomspace rent too low")
		assert.LessOrEqual(t, price, 3000.0, "Roomspace rent too high")

	default:
		t.Errorf("Unknown price type: %s", priceType)
	}
}

// AssertSourceURL validates that source URL is properly formatted
func AssertSourceURL(t *testing.T, url, expectedDomain string) {
	t.Helper()

	if url == "" {
		t.Log("Warning: SourceURL is empty")
		return
	}

	assert.Contains(t, url, expectedDomain,
		"URL should contain domain '%s'", expectedDomain)
}

// AssertTagsContain validates that required tags are present
func AssertTagsContain(t *testing.T, tags []string, requiredTags ...string) {
	t.Helper()

	for _, required := range requiredTags {
		found := false
		for _, tag := range tags {
			if tag == required {
				found = true
				break
			}
		}
		assert.True(t, found, "Required tag '%s' not found in tags: %v", required, tags)
	}
}

// PrintDataPointSummary prints a summary of scraped data points (useful for debugging)
func PrintDataPointSummary(t *testing.T, dataPoints []*models.CostDataPoint) {
	t.Helper()

	t.Logf("=== Data Points Summary ===")
	t.Logf("Total count: %d", len(dataPoints))

	if len(dataPoints) > 0 {
		var totalPrice float64
		minPrice := dataPoints[0].Price
		maxPrice := dataPoints[0].Price

		for _, cdp := range dataPoints {
			totalPrice += cdp.Price
			if cdp.Price < minPrice {
				minPrice = cdp.Price
			}
			if cdp.Price > maxPrice {
				maxPrice = cdp.Price
			}
		}

		avgPrice := totalPrice / float64(len(dataPoints))
		t.Logf("Price range: %.2f - %.2f AED", minPrice, maxPrice)
		t.Logf("Average price: %.2f AED", avgPrice)
		t.Logf("First item: %s (%.2f AED)", dataPoints[0].ItemName, dataPoints[0].Price)
	}
}
