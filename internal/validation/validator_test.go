package validation

import (
	"context"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func createTestDataPoint(category, itemName, emirate string, price float64) *models.CostDataPoint {
	return &models.CostDataPoint{
		ID:         "test-1",
		Category:   category,
		ItemName:   itemName,
		Price:      price,
		Location:   models.Location{Emirate: emirate},
		RecordedAt: time.Now(),
		Source:     "TestSource",
		Confidence: 0.8,
		Unit:       "AED/year",
		SampleSize: 1,
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	if validator.config == nil {
		t.Error("Expected config to be initialized")
	}

	if len(validator.rules) == 0 {
		t.Error("Expected rules to be loaded")
	}
}

func TestValidateDataPoint_Valid(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	dp := createTestDataPoint("Housing", "Studio Apartment", "Dubai", 50000)
	dp.Attributes = map[string]interface{}{
		"bedrooms":  1,
		"area_sqft": 500,
	}

	result, err := validator.ValidateDataPoint(ctx, dp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.IsValid {
		t.Errorf("Expected data point to be valid, got errors: %v", result.Errors)
	}

	if result.Score < 0.8 {
		t.Errorf("Expected high quality score, got: %f", result.Score)
	}
}

func TestValidateDataPoint_Invalid(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name          string
		dp            *models.CostDataPoint
		expectedError bool
	}{
		{
			name:          "negative price",
			dp:            createTestDataPoint("Housing", "Studio", "Dubai", -1000),
			expectedError: true,
		},
		{
			name:          "empty item name",
			dp:            createTestDataPoint("Housing", "", "Dubai", 50000),
			expectedError: true,
		},
		{
			name:          "invalid emirate",
			dp:            createTestDataPoint("Housing", "Studio", "Invalid", 50000),
			expectedError: true,
		},
		{
			name:          "invalid category",
			dp:            createTestDataPoint("InvalidCategory", "Item", "Dubai", 100),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateDataPoint(ctx, tt.dp)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectedError && result.IsValid {
				t.Error("Expected validation to fail")
			}

			if tt.expectedError && len(result.Errors) == 0 {
				t.Error("Expected validation errors")
			}
		})
	}
}

func TestValidateDataPoint_NilDataPoint(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	_, err := validator.ValidateDataPoint(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil data point")
	}
}

func TestValidateBatch(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	points := []*models.CostDataPoint{
		createTestDataPoint("Housing", "Studio", "Dubai", 50000),
		createTestDataPoint("Housing", "1BR", "Dubai", 80000),
		createTestDataPoint("Housing", "2BR", "Dubai", 120000),
	}

	// Add required attributes for housing
	for _, p := range points {
		p.Attributes = map[string]interface{}{
			"bedrooms":  1,
			"area_sqft": 500,
		}
	}

	results, err := validator.ValidateBatch(ctx, points)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != len(points) {
		t.Errorf("Expected %d results, got %d", len(points), len(results))
	}

	for i, result := range results {
		if !result.IsValid {
			t.Errorf("Point %d validation failed: %v", i, result.Errors)
		}
	}
}

func TestValidateBatch_EmptyBatch(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	_, err := validator.ValidateBatch(ctx, []*models.CostDataPoint{})
	if err == nil {
		t.Error("Expected error for empty batch")
	}
}

func TestValidateBatch_ExceedsMaxSize(t *testing.T) {
	validator := NewValidator()
	validator.config.MaxBatchSize = 2
	ctx := context.Background()

	points := []*models.CostDataPoint{
		createTestDataPoint("Housing", "Studio", "Dubai", 50000),
		createTestDataPoint("Housing", "1BR", "Dubai", 80000),
		createTestDataPoint("Housing", "2BR", "Dubai", 120000),
	}

	_, err := validator.ValidateBatch(ctx, points)
	if err == nil {
		t.Error("Expected error for batch exceeding max size")
	}
}

func TestGetRulesForCategory(t *testing.T) {
	validator := NewValidator()

	housingRules := validator.GetRulesForCategory("Housing")
	if len(housingRules) == 0 {
		t.Error("Expected housing rules to be returned")
	}

	// Should include common rules and housing-specific rules
	hasCommonRule := false
	hasHousingRule := false

	for _, rule := range housingRules {
		if rule.Category == "all" {
			hasCommonRule = true
		}
		if rule.Category == "Housing" {
			hasHousingRule = true
		}
	}

	if !hasCommonRule {
		t.Error("Expected common rules to be included")
	}

	if !hasHousingRule {
		t.Error("Expected housing-specific rules to be included")
	}
}

func TestAddRule(t *testing.T) {
	validator := NewValidator()
	initialCount := len(validator.rules)

	customRule := Rule{
		Name:     "custom_rule",
		Category: "Housing",
		Field:    "Price",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			return nil
		},
	}

	validator.AddRule(customRule)

	if len(validator.rules) != initialCount+1 {
		t.Errorf("Expected %d rules, got %d", initialCount+1, len(validator.rules))
	}
}

func TestRemoveRule(t *testing.T) {
	validator := NewValidator()

	// Add a custom rule
	customRule := Rule{
		Name:     "test_rule_to_remove",
		Category: "Housing",
		Field:    "Price",
		Severity: SeverityWarning,
		Validator: func(dp *models.CostDataPoint) error {
			return nil
		},
	}

	validator.AddRule(customRule)
	initialCount := len(validator.rules)

	validator.RemoveRule("test_rule_to_remove")

	if len(validator.rules) != initialCount-1 {
		t.Errorf("Expected %d rules after removal, got %d", initialCount-1, len(validator.rules))
	}
}

func TestValidationResult_QualityScore(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ValidationError
		expected float64
	}{
		{
			name:     "no errors",
			errors:   []ValidationError{},
			expected: 1.0,
		},
		{
			name: "one error",
			errors: []ValidationError{
				{Severity: SeverityError},
			},
			expected: 0.7,
		},
		{
			name: "one warning",
			errors: []ValidationError{
				{Severity: SeverityWarning},
			},
			expected: 0.9,
		},
		{
			name: "mixed errors",
			errors: []ValidationError{
				{Severity: SeverityError},
				{Severity: SeverityWarning},
			},
			expected: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors: tt.errors,
			}

			score := result.QualityScore()
			if score != tt.expected {
				t.Errorf("Expected quality score %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestValidatorConfig(t *testing.T) {
	config := DefaultValidatorConfig()

	if !config.EnableOutlierDetection {
		t.Error("Expected outlier detection to be enabled by default")
	}

	if !config.EnableDuplicateCheck {
		t.Error("Expected duplicate check to be enabled by default")
	}

	if !config.EnableFreshnessCheck {
		t.Error("Expected freshness check to be enabled by default")
	}

	if config.MaxBatchSize != 10000 {
		t.Errorf("Expected max batch size 10000, got %d", config.MaxBatchSize)
	}
}

func TestValidateDataPoint_WithOutlierDetectionDisabled(t *testing.T) {
	config := DefaultValidatorConfig()
	config.EnableOutlierDetection = false
	validator := NewValidatorWithConfig(config)

	ctx := context.Background()
	dp := createTestDataPoint("Housing", "Studio", "Dubai", 50000)
	dp.Attributes = map[string]interface{}{
		"bedrooms":  1,
		"area_sqft": 500,
	}

	result, err := validator.ValidateDataPoint(ctx, dp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result")
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.severity.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.severity.String())
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := ValidationError{
		Field:    "Price",
		Message:  "invalid price",
		Severity: SeverityError,
		Value:    -100,
	}

	errorStr := ve.Error()
	if errorStr == "" {
		t.Error("Expected non-empty error string")
	}

	// Check that error string contains key information
	expectedParts := []string{"ERROR", "Price", "invalid price"}
	for _, part := range expectedParts {
		if !contains(errorStr, part) {
			t.Errorf("Expected error string to contain %s, got: %s", part, errorStr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
