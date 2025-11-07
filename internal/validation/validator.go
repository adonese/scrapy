package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// Severity represents the severity level of a validation error
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// String returns the string representation of severity
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field    string
	Message  string
	Severity Severity
	Value    interface{}
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s (value: %v)", ve.Severity, ve.Field, ve.Message, ve.Value)
}

// ValidationResult represents the result of validating a data point
type ValidationResult struct {
	DataPoint  *models.CostDataPoint
	Errors     []ValidationError
	Warnings   []string
	Score      float64 // 0-1 quality score
	IsValid    bool
	ValidatedAt time.Time
}

// QualityScore calculates the quality score based on errors and warnings
func (vr *ValidationResult) QualityScore() float64 {
	if len(vr.Errors) > 0 {
		// Start with base score and deduct for each error/warning
		score := 1.0
		for _, err := range vr.Errors {
			switch err.Severity {
			case SeverityError:
				score -= 0.3
			case SeverityWarning:
				score -= 0.1
			case SeverityInfo:
				score -= 0.05
			}
		}
		if score < 0 {
			score = 0
		}
		return score
	}
	return 1.0
}

// Validator defines the interface for data validation
type Validator interface {
	// ValidateDataPoint validates a single data point
	ValidateDataPoint(ctx context.Context, dp *models.CostDataPoint) (*ValidationResult, error)

	// ValidateBatch validates multiple data points
	ValidateBatch(ctx context.Context, points []*models.CostDataPoint) ([]*ValidationResult, error)

	// GetRulesForCategory returns applicable rules for a category
	GetRulesForCategory(category string) []Rule

	// AddRule adds a custom validation rule
	AddRule(rule Rule)

	// RemoveRule removes a validation rule by name
	RemoveRule(name string)
}

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	rules            []Rule
	outlierDetector  *OutlierDetector
	duplicateChecker *DuplicateChecker
	freshnessChecker *FreshnessChecker
	config           *ValidatorConfig
}

// ValidatorConfig holds configuration for the validator
type ValidatorConfig struct {
	EnableOutlierDetection bool
	EnableDuplicateCheck   bool
	EnableFreshnessCheck   bool
	StrictMode             bool // Fail on warnings in strict mode
	MaxBatchSize           int
}

// DefaultValidatorConfig returns the default configuration
func DefaultValidatorConfig() *ValidatorConfig {
	return &ValidatorConfig{
		EnableOutlierDetection: true,
		EnableDuplicateCheck:   true,
		EnableFreshnessCheck:   true,
		StrictMode:             false,
		MaxBatchSize:           10000,
	}
}

// NewValidator creates a new validator with default configuration
func NewValidator() *DefaultValidator {
	return NewValidatorWithConfig(DefaultValidatorConfig())
}

// NewValidatorWithConfig creates a new validator with custom configuration
func NewValidatorWithConfig(config *ValidatorConfig) *DefaultValidator {
	v := &DefaultValidator{
		rules:            GetAllRules(),
		outlierDetector:  NewOutlierDetector(DetectionMethodIQR, 1.5),
		duplicateChecker: NewDuplicateChecker(24 * time.Hour),
		freshnessChecker: NewFreshnessChecker(),
		config:           config,
	}
	return v
}

// ValidateDataPoint validates a single data point
func (v *DefaultValidator) ValidateDataPoint(ctx context.Context, dp *models.CostDataPoint) (*ValidationResult, error) {
	if dp == nil {
		return nil, fmt.Errorf("data point is nil")
	}

	result := &ValidationResult{
		DataPoint:   dp,
		Errors:      make([]ValidationError, 0),
		Warnings:    make([]string, 0),
		ValidatedAt: time.Now(),
		IsValid:     true,
	}

	// Apply all rules for this category
	categoryRules := v.GetRulesForCategory(dp.Category)
	for _, rule := range categoryRules {
		if err := rule.Validator(dp); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:    rule.Field,
				Message:  err.Error(),
				Severity: rule.Severity,
				Value:    getFieldValue(dp, rule.Field),
			})
			if rule.Severity == SeverityError {
				result.IsValid = false
			}
		}
	}

	// Check data freshness
	if v.config.EnableFreshnessCheck {
		status := v.freshnessChecker.CheckFreshness(dp.Source, dp.RecordedAt)
		if status == FreshnessStale {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Data is stale for source %s", dp.Source))
		}
	}

	// Calculate quality score
	result.Score = result.QualityScore()

	return result, nil
}

// ValidateBatch validates multiple data points
func (v *DefaultValidator) ValidateBatch(ctx context.Context, points []*models.CostDataPoint) ([]*ValidationResult, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	if len(points) > v.config.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(points), v.config.MaxBatchSize)
	}

	results := make([]*ValidationResult, 0, len(points))

	// Validate each point individually
	for _, dp := range points {
		result, err := v.ValidateDataPoint(ctx, dp)
		if err != nil {
			return nil, fmt.Errorf("error validating data point %s: %w", dp.ID, err)
		}
		results = append(results, result)
	}

	// Detect outliers across the batch
	if v.config.EnableOutlierDetection {
		outlierIndices := v.outlierDetector.DetectOutliers(points)
		for _, idx := range outlierIndices {
			if idx < len(results) {
				results[idx].Warnings = append(results[idx].Warnings, "Detected as statistical outlier")
				results[idx].Score *= 0.9 // Reduce score for outliers
			}
		}
	}

	// Detect duplicates
	if v.config.EnableDuplicateCheck {
		duplicateGroups := v.duplicateChecker.DetectDuplicates(points)
		for _, group := range duplicateGroups {
			for _, idx := range group.Indices {
				if idx < len(results) {
					results[idx].Warnings = append(results[idx].Warnings,
						fmt.Sprintf("Potential duplicate (group size: %d)", len(group.Indices)))
					results[idx].Score *= 0.95 // Slightly reduce score for duplicates
				}
			}
		}
	}

	return results, nil
}

// GetRulesForCategory returns applicable rules for a category
func (v *DefaultValidator) GetRulesForCategory(category string) []Rule {
	categoryRules := make([]Rule, 0)
	for _, rule := range v.rules {
		if rule.Category == category || rule.Category == "all" {
			categoryRules = append(categoryRules, rule)
		}
	}
	return categoryRules
}

// AddRule adds a custom validation rule
func (v *DefaultValidator) AddRule(rule Rule) {
	v.rules = append(v.rules, rule)
}

// RemoveRule removes a validation rule by name
func (v *DefaultValidator) RemoveRule(name string) {
	for i, rule := range v.rules {
		if rule.Name == name {
			v.rules = append(v.rules[:i], v.rules[i+1:]...)
			return
		}
	}
}

// getFieldValue extracts field value from data point using reflection-like logic
func getFieldValue(dp *models.CostDataPoint, field string) interface{} {
	switch field {
	case "Price":
		return dp.Price
	case "MinPrice":
		return dp.MinPrice
	case "MaxPrice":
		return dp.MaxPrice
	case "Category":
		return dp.Category
	case "ItemName":
		return dp.ItemName
	case "Location":
		return dp.Location
	case "Source":
		return dp.Source
	case "Confidence":
		return dp.Confidence
	default:
		return nil
	}
}
