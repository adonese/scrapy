package validation

import (
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func TestGetAllRules(t *testing.T) {
	rules := GetAllRules()

	if len(rules) == 0 {
		t.Error("Expected rules to be returned")
	}

	// Check that we have rules from different categories
	hasCommonRules := false
	hasHousingRules := false
	hasUtilitiesRules := false

	for _, rule := range rules {
		if rule.Category == "all" {
			hasCommonRules = true
		}
		if rule.Category == "Housing" {
			hasHousingRules = true
		}
		if rule.Category == "Utilities" {
			hasUtilitiesRules = true
		}
	}

	if !hasCommonRules {
		t.Error("Expected common rules")
	}
	if !hasHousingRules {
		t.Error("Expected housing rules")
	}
	if !hasUtilitiesRules {
		t.Error("Expected utilities rules")
	}
}

func TestCommonRules_RequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		dp        *models.CostDataPoint
		shouldErr bool
	}{
		{
			name: "all fields present",
			dp: &models.CostDataPoint{
				ItemName: "Test Item",
				Category: "Housing",
				Source:   "TestSource",
				Location: models.Location{Emirate: "Dubai"},
			},
			shouldErr: false,
		},
		{
			name: "missing item name",
			dp: &models.CostDataPoint{
				ItemName: "",
				Category: "Housing",
				Source:   "TestSource",
				Location: models.Location{Emirate: "Dubai"},
			},
			shouldErr: true,
		},
		{
			name: "missing category",
			dp: &models.CostDataPoint{
				ItemName: "Test Item",
				Category: "",
				Source:   "TestSource",
				Location: models.Location{Emirate: "Dubai"},
			},
			shouldErr: true,
		},
		{
			name: "missing source",
			dp: &models.CostDataPoint{
				ItemName: "Test Item",
				Category: "Housing",
				Source:   "",
				Location: models.Location{Emirate: "Dubai"},
			},
			shouldErr: true,
		},
		{
			name: "missing emirate",
			dp: &models.CostDataPoint{
				ItemName: "Test Item",
				Category: "Housing",
				Source:   "TestSource",
				Location: models.Location{Emirate: ""},
			},
			shouldErr: true,
		},
	}

	rule := CommonRules[0] // required_fields rule

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validator(tt.dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_ValidCategory(t *testing.T) {
	tests := []struct {
		category  string
		shouldErr bool
	}{
		{"Housing", false},
		{"Utilities", false},
		{"Transportation", false},
		{"Food", false},
		{"Education", false},
		{"InvalidCategory", true},
		{"", true},
	}

	// Find the valid_category rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "valid_category" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			dp := &models.CostDataPoint{
				Category: tt.category,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Category %s: expected error: %v, got: %v", tt.category, tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_ValidEmirate(t *testing.T) {
	tests := []struct {
		emirate   string
		shouldErr bool
	}{
		{"Dubai", false},
		{"Abu Dhabi", false},
		{"Sharjah", false},
		{"Ajman", false},
		{"InvalidEmirate", true},
		{"", true},
	}

	// Find the valid_emirate rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "valid_emirate" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run(tt.emirate, func(t *testing.T) {
			dp := &models.CostDataPoint{
				Location: models.Location{Emirate: tt.emirate},
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Emirate %s: expected error: %v, got: %v", tt.emirate, tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_PositivePrice(t *testing.T) {
	tests := []struct {
		price     float64
		shouldErr bool
	}{
		{100.0, false},
		{0.01, false},
		{1000000.0, false},
		{0, true},
		{-1, true},
		{-100, true},
	}

	// Find the positive_price rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "positive_price" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dp := &models.CostDataPoint{
				Price: tt.price,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Price %f: expected error: %v, got: %v", tt.price, tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_ValidConfidence(t *testing.T) {
	tests := []struct {
		confidence float32
		shouldErr  bool
	}{
		{0.8, false},
		{1.0, false},
		{0.5, false},
		{0.4, true},  // Low confidence
		{-0.1, true}, // Invalid range
		{1.5, true},  // Invalid range
	}

	// Find the valid_confidence rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "valid_confidence" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dp := &models.CostDataPoint{
				Confidence: tt.confidence,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Confidence %f: expected error: %v, got: %v", tt.confidence, tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_ValidTimestamp(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		recordedAt time.Time
		shouldErr  bool
	}{
		{"recent", now.Add(-1 * time.Hour), false},
		{"one month ago", now.Add(-30 * 24 * time.Hour), false},
		{"future", now.Add(1 * time.Hour), true},
		{"too old", now.Add(-400 * 24 * time.Hour), true},
		{"zero time", time.Time{}, true},
	}

	// Find the valid_timestamp rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "valid_timestamp" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &models.CostDataPoint{
				RecordedAt: tt.recordedAt,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

func TestCommonRules_MinMaxPriceConsistency(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		minPrice  float64
		maxPrice  float64
		shouldErr bool
	}{
		{"valid range", 100, 80, 120, false},
		{"min > max", 100, 120, 80, true},
		{"price below min", 70, 80, 120, true},
		{"price above max", 130, 80, 120, true},
		{"no min/max", 100, 0, 0, false},
	}

	// Find the min_max_price_consistency rule
	var rule Rule
	for _, r := range CommonRules {
		if r.Name == "min_max_price_consistency" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &models.CostDataPoint{
				Price:    tt.price,
				MinPrice: tt.minPrice,
				MaxPrice: tt.maxPrice,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

func TestHousingRules_PriceRange(t *testing.T) {
	tests := []struct {
		price     float64
		shouldErr bool
	}{
		{50000, false},
		{100000, false},
		{5000, true},   // Below min
		{6000000, true}, // Above max
	}

	// Find the housing_price_range rule
	var rule Rule
	for _, r := range HousingRules {
		if r.Name == "housing_price_range" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dp := &models.CostDataPoint{
				Category: "Housing",
				Price:    tt.price,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Price %f: expected error: %v, got: %v", tt.price, tt.shouldErr, err)
			}
		})
	}
}

func TestHousingRules_Attributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
		shouldErr  bool
	}{
		{
			name: "all attributes",
			attributes: map[string]interface{}{
				"bedrooms":  2,
				"area_sqft": 1000,
			},
			shouldErr: false,
		},
		{
			name: "missing bedrooms",
			attributes: map[string]interface{}{
				"area_sqft": 1000,
			},
			shouldErr: true,
		},
		{
			name: "missing area_sqft",
			attributes: map[string]interface{}{
				"bedrooms": 2,
			},
			shouldErr: true,
		},
	}

	// Find the housing_attributes rule
	var rule Rule
	for _, r := range HousingRules {
		if r.Name == "housing_attributes" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &models.CostDataPoint{
				Category:   "Housing",
				Attributes: tt.attributes,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

func TestUtilitiesRules_PriceRange(t *testing.T) {
	tests := []struct {
		price     float64
		shouldErr bool
	}{
		{100, false},
		{500, false},
		{30, true},   // Below min
		{3000, true}, // Above max
	}

	// Find the utilities_price_range rule
	var rule Rule
	for _, r := range UtilitiesRules {
		if r.Name == "utilities_price_range" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dp := &models.CostDataPoint{
				Category: "Utilities",
				Price:    tt.price,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Price %f: expected error: %v, got: %v", tt.price, tt.shouldErr, err)
			}
		})
	}
}

func TestTransportationRules_PriceRange(t *testing.T) {
	tests := []struct {
		price     float64
		shouldErr bool
	}{
		{5, false},
		{50, false},
		{0.5, true},  // Below min
		{150, true}, // Above max
	}

	// Find the transportation_price_range rule
	var rule Rule
	for _, r := range TransportationRules {
		if r.Name == "transportation_price_range" {
			rule = r
			break
		}
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			dp := &models.CostDataPoint{
				Category: "Transportation",
				Price:    tt.price,
			}
			err := rule.Validator(dp)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Price %f: expected error: %v, got: %v", tt.price, tt.shouldErr, err)
			}
		})
	}
}

func TestPriceRanges(t *testing.T) {
	// Verify that all categories have price ranges defined
	categories := []string{
		"Housing", "Utilities", "Transportation", "Food",
		"Education", "Entertainment", "Healthcare", "Shopping",
		"Communications", "Personal Care",
	}

	for _, category := range categories {
		t.Run(category, func(t *testing.T) {
			priceRange, ok := PriceRanges[category]
			if !ok {
				t.Errorf("Price range not defined for category: %s", category)
				return
			}

			if priceRange.Min >= priceRange.Max {
				t.Errorf("Invalid price range for %s: min (%f) >= max (%f)",
					category, priceRange.Min, priceRange.Max)
			}

			if priceRange.Min < 0 {
				t.Errorf("Invalid minimum price for %s: %f", category, priceRange.Min)
			}
		})
	}
}

func TestValidEmirates(t *testing.T) {
	expectedEmirates := []string{
		"Dubai", "Abu Dhabi", "Sharjah", "Ajman",
		"Umm Al Quwain", "Ras Al Khaimah", "Fujairah",
	}

	for _, emirate := range expectedEmirates {
		t.Run(emirate, func(t *testing.T) {
			if !ValidEmirates[emirate] {
				t.Errorf("Expected %s to be a valid emirate", emirate)
			}
		})
	}

	if len(ValidEmirates) != 7 {
		t.Errorf("Expected 7 valid emirates, got %d", len(ValidEmirates))
	}
}

func TestValidCategories(t *testing.T) {
	expectedCategories := []string{
		"Housing", "Utilities", "Transportation", "Food",
		"Education", "Entertainment", "Healthcare", "Shopping",
		"Communications", "Personal Care",
	}

	for _, category := range expectedCategories {
		t.Run(category, func(t *testing.T) {
			if !ValidCategories[category] {
				t.Errorf("Expected %s to be a valid category", category)
			}
		})
	}

	if len(ValidCategories) != 10 {
		t.Errorf("Expected 10 valid categories, got %d", len(ValidCategories))
	}
}
