package validation

import (
	"fmt"
	"strings"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// Rule represents a validation rule
type Rule struct {
	Name      string
	Category  string // "all" for rules that apply to all categories
	Field     string
	Severity  Severity
	Validator func(*models.CostDataPoint) error
}

// PriceRange defines acceptable price ranges
type PriceRange struct {
	Min float64
	Max float64
}

// Price ranges by category (in AED)
var PriceRanges = map[string]PriceRange{
	"Housing":        {Min: 10000, Max: 5000000},    // Annual rent
	"Utilities":      {Min: 50, Max: 2000},          // Monthly bill
	"Transportation": {Min: 1, Max: 100},            // Per trip/km
	"Food":           {Min: 0.5, Max: 500},          // Per item
	"Education":      {Min: 5000, Max: 200000},      // Annual fees
	"Entertainment":  {Min: 10, Max: 5000},          // Per activity
	"Healthcare":     {Min: 50, Max: 50000},         // Per visit/procedure
	"Shopping":       {Min: 1, Max: 10000},          // Per item
	"Communications": {Min: 50, Max: 1000},          // Monthly
	"Personal Care":  {Min: 10, Max: 2000},          // Per service
}

// Valid emirates in the UAE
var ValidEmirates = map[string]bool{
	"Dubai":       true,
	"Abu Dhabi":   true,
	"Sharjah":     true,
	"Ajman":       true,
	"Umm Al Quwain": true,
	"Ras Al Khaimah": true,
	"Fujairah":    true,
}

// Valid categories
var ValidCategories = map[string]bool{
	"Housing":        true,
	"Utilities":      true,
	"Transportation": true,
	"Food":           true,
	"Education":      true,
	"Entertainment":  true,
	"Healthcare":     true,
	"Shopping":       true,
	"Communications": true,
	"Personal Care":  true,
}

// GetAllRules returns all validation rules
func GetAllRules() []Rule {
	return append(
		CommonRules,
		append(HousingRules,
			append(UtilitiesRules,
				append(TransportationRules,
					append(FoodRules, EducationRules...)...)...)...)...,
	)
}

// CommonRules apply to all data points regardless of category
var CommonRules = []Rule{
	{
		Name:     "required_fields",
		Category: "all",
		Field:    "ItemName",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.ItemName == "" {
				return fmt.Errorf("item name is required")
			}
			if dp.Category == "" {
				return fmt.Errorf("category is required")
			}
			if dp.Source == "" {
				return fmt.Errorf("source is required")
			}
			if dp.Location.Emirate == "" {
				return fmt.Errorf("emirate is required")
			}
			return nil
		},
	},
	{
		Name:     "valid_category",
		Category: "all",
		Field:    "Category",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if !ValidCategories[dp.Category] {
				return fmt.Errorf("invalid category: %s", dp.Category)
			}
			return nil
		},
	},
	{
		Name:     "valid_emirate",
		Category: "all",
		Field:    "Location",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if !ValidEmirates[dp.Location.Emirate] {
				return fmt.Errorf("invalid emirate: %s", dp.Location.Emirate)
			}
			return nil
		},
	},
	{
		Name:     "positive_price",
		Category: "all",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.Price <= 0 {
				return fmt.Errorf("price must be positive: %f", dp.Price)
			}
			return nil
		},
	},
	{
		Name:     "valid_confidence",
		Category: "all",
		Field:    "Confidence",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.Confidence < 0 || dp.Confidence > 1 {
				return fmt.Errorf("confidence must be between 0 and 1: %f", dp.Confidence)
			}
			if dp.Confidence < 0.5 {
				return fmt.Errorf("low confidence score: %f", dp.Confidence)
			}
			return nil
		},
	},
	{
		Name:     "valid_timestamp",
		Category: "all",
		Field:    "RecordedAt",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.RecordedAt.IsZero() {
				return fmt.Errorf("recorded_at timestamp is required")
			}
			if dp.RecordedAt.After(time.Now()) {
				return fmt.Errorf("recorded_at cannot be in the future")
			}
			// Check if data is too old (more than 1 year)
			if time.Since(dp.RecordedAt) > 365*24*time.Hour {
				return fmt.Errorf("data is older than 1 year")
			}
			return nil
		},
	},
	{
		Name:     "min_max_price_consistency",
		Category: "all",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.MinPrice > 0 && dp.MaxPrice > 0 {
				if dp.MinPrice > dp.MaxPrice {
					return fmt.Errorf("min_price (%f) cannot be greater than max_price (%f)", dp.MinPrice, dp.MaxPrice)
				}
				if dp.Price < dp.MinPrice || dp.Price > dp.MaxPrice {
					return fmt.Errorf("price (%f) must be between min_price (%f) and max_price (%f)", dp.Price, dp.MinPrice, dp.MaxPrice)
				}
			}
			return nil
		},
	},
	{
		Name:     "sample_size_validation",
		Category: "all",
		Field:    "SampleSize",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			if dp.SampleSize < 1 {
				return fmt.Errorf("sample size should be at least 1")
			}
			if dp.SampleSize == 1 && dp.Confidence > 0.7 {
				return fmt.Errorf("high confidence (%f) with sample size of 1 is suspicious", dp.Confidence)
			}
			return nil
		},
	},
}

// HousingRules apply to housing data
var HousingRules = []Rule{
	{
		Name:     "housing_price_range",
		Category: "Housing",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			priceRange := PriceRanges["Housing"]
			if dp.Price < priceRange.Min || dp.Price > priceRange.Max {
				return fmt.Errorf("housing price out of range: %f (expected %f-%f)", dp.Price, priceRange.Min, priceRange.Max)
			}
			return nil
		},
	},
	{
		Name:     "housing_attributes",
		Category: "Housing",
		Field:    "Attributes",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			// Check for expected attributes
			if _, ok := dp.Attributes["bedrooms"]; !ok {
				return fmt.Errorf("missing 'bedrooms' attribute for housing data")
			}
			if _, ok := dp.Attributes["area_sqft"]; !ok {
				return fmt.Errorf("missing 'area_sqft' attribute for housing data")
			}
			return nil
		},
	},
	{
		Name:     "housing_unit",
		Category: "Housing",
		Field:    "Unit",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			validUnits := map[string]bool{
				"AED/year":  true,
				"AED/month": true,
			}
			if !validUnits[dp.Unit] {
				return fmt.Errorf("unexpected unit for housing: %s (expected AED/year or AED/month)", dp.Unit)
			}
			return nil
		},
	},
}

// UtilitiesRules apply to utilities data
var UtilitiesRules = []Rule{
	{
		Name:     "utilities_price_range",
		Category: "Utilities",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			priceRange := PriceRanges["Utilities"]
			if dp.Price < priceRange.Min || dp.Price > priceRange.Max {
				return fmt.Errorf("utilities price out of range: %f (expected %f-%f)", dp.Price, priceRange.Min, priceRange.Max)
			}
			return nil
		},
	},
	{
		Name:     "utilities_provider",
		Category: "Utilities",
		Field:    "Source",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			validProviders := []string{"DEWA", "SEWA", "AADC", "ADDC", "FEWA"}
			isValid := false
			for _, provider := range validProviders {
				if strings.Contains(strings.ToUpper(dp.Source), provider) {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("unexpected utility provider: %s", dp.Source)
			}
			return nil
		},
	},
	{
		Name:     "utilities_unit",
		Category: "Utilities",
		Field:    "Unit",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			validUnits := map[string]bool{
				"AED/month": true,
				"AED/kWh":   true,
				"AED/unit":  true,
			}
			if !validUnits[dp.Unit] {
				return fmt.Errorf("unexpected unit for utilities: %s", dp.Unit)
			}
			return nil
		},
	},
}

// TransportationRules apply to transportation data
var TransportationRules = []Rule{
	{
		Name:     "transportation_price_range",
		Category: "Transportation",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			priceRange := PriceRanges["Transportation"]
			if dp.Price < priceRange.Min || dp.Price > priceRange.Max {
				return fmt.Errorf("transportation price out of range: %f (expected %f-%f)", dp.Price, priceRange.Min, priceRange.Max)
			}
			return nil
		},
	},
	{
		Name:     "transportation_source",
		Category: "Transportation",
		Field:    "Source",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			validSources := []string{"RTA", "Careem", "Uber", "Emirates Transport"}
			isValid := false
			for _, source := range validSources {
				if strings.Contains(dp.Source, source) {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("unexpected transportation source: %s", dp.Source)
			}
			return nil
		},
	},
}

// FoodRules apply to food data
var FoodRules = []Rule{
	{
		Name:     "food_price_range",
		Category: "Food",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			priceRange := PriceRanges["Food"]
			if dp.Price < priceRange.Min || dp.Price > priceRange.Max {
				return fmt.Errorf("food price out of range: %f (expected %f-%f)", dp.Price, priceRange.Min, priceRange.Max)
			}
			return nil
		},
	},
}

// EducationRules apply to education data
var EducationRules = []Rule{
	{
		Name:     "education_price_range",
		Category: "Education",
		Field:    "Price",
		Severity: SeverityError,
		Validator: func(dp *models.CostDataPoint) error {
			priceRange := PriceRanges["Education"]
			if dp.Price < priceRange.Min || dp.Price > priceRange.Max {
				return fmt.Errorf("education price out of range: %f (expected %f-%f)", dp.Price, priceRange.Min, priceRange.Max)
			}
			return nil
		},
	},
	{
		Name:     "education_unit",
		Category: "Education",
		Field:    "Unit",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			validUnits := map[string]bool{
				"AED/year":     true,
				"AED/semester": true,
				"AED/term":     true,
			}
			if !validUnits[dp.Unit] {
				return fmt.Errorf("unexpected unit for education: %s", dp.Unit)
			}
			return nil
		},
	},
}
