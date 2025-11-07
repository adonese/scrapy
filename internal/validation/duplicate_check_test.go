package validation

import (
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func TestNewDuplicateChecker(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)
	if dc == nil {
		t.Fatal("Expected duplicate checker to be created")
	}

	if dc.timeWindow != 24*time.Hour {
		t.Error("Expected time window to be 24 hours")
	}
}

func TestDetectDuplicates_ExactDuplicates(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	points := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now.Add(1 * time.Hour),
		},
		{
			Category:   "Housing",
			ItemName:   "1BR Apartment",
			Price:      80000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
	}

	groups := dc.DetectDuplicates(points)

	if len(groups) == 0 {
		t.Error("Expected duplicate groups to be found")
	}

	// Should find the two studio apartments as duplicates
	foundDuplicate := false
	for _, group := range groups {
		if len(group.Indices) == 2 {
			foundDuplicate = true
			break
		}
	}

	if !foundDuplicate {
		t.Error("Expected to find duplicate group with 2 items")
	}
}

func TestDetectDuplicates_NoDuplicates(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	points := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
		{
			Category:   "Housing",
			ItemName:   "1BR Apartment",
			Price:      80000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
		{
			Category:   "Housing",
			ItemName:   "2BR Apartment",
			Price:      120000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
	}

	groups := dc.DetectDuplicates(points)

	if len(groups) != 0 {
		t.Errorf("Expected no duplicate groups, got %d", len(groups))
	}
}

func TestDetectDuplicates_SmallDataset(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	points := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: time.Now(),
		},
	}

	groups := dc.DetectDuplicates(points)

	if len(groups) != 0 {
		t.Error("Expected no duplicates for single data point")
	}
}

func TestAreSimilar(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	basePoint := &models.CostDataPoint{
		Category:   "Housing",
		ItemName:   "Studio Apartment",
		Price:      50000,
		Location:   models.Location{Emirate: "Dubai"},
		Source:     "Bayut",
		RecordedAt: now,
	}

	tests := []struct {
		name       string
		point      *models.CostDataPoint
		shouldMatch bool
	}{
		{
			name: "similar price",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50500, // Within 5% threshold
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: true,
		},
		{
			name: "different category",
			point: &models.CostDataPoint{
				Category:   "Utilities",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: false,
		},
		{
			name: "different item name",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "1BR Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: false,
		},
		{
			name: "different emirate",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Abu Dhabi"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: false,
		},
		{
			name: "different source",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Dubizzle",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: false,
		},
		{
			name: "price too different",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      60000, // More than 5% difference
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			shouldMatch: false,
		},
		{
			name: "outside time window",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(25 * time.Hour), // Outside 24-hour window
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dc.areSimilar(basePoint, tt.point)
			if result != tt.shouldMatch {
				t.Errorf("Expected similar: %v, got: %v", tt.shouldMatch, result)
			}
		})
	}
}

func TestCalculateSimilarity(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	dp1 := &models.CostDataPoint{
		Category:   "Housing",
		ItemName:   "Studio Apartment",
		Price:      50000,
		Location:   models.Location{Emirate: "Dubai"},
		Source:     "Bayut",
		RecordedAt: now,
	}

	tests := []struct {
		name     string
		dp2      *models.CostDataPoint
		minScore float64
	}{
		{
			name: "identical",
			dp2: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now,
			},
			minScore: 0.99,
		},
		{
			name: "different category only",
			dp2: &models.CostDataPoint{
				Category:   "Utilities",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now,
			},
			minScore: 0.7,
		},
		{
			name: "completely different",
			dp2: &models.CostDataPoint{
				Category:   "Food",
				ItemName:   "Bread",
				Price:      5,
				Location:   models.Location{Emirate: "Abu Dhabi"},
				Source:     "Carrefour",
				RecordedAt: now,
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := dc.calculateSimilarity(dp1, tt.dp2)
			if score < tt.minScore {
				t.Errorf("Expected similarity >= %f, got %f", tt.minScore, score)
			}
		})
	}
}

func TestIsDuplicate(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	existing := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
	}

	tests := []struct {
		name        string
		point       *models.CostDataPoint
		isDuplicate bool
	}{
		{
			name: "exact duplicate",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "Studio Apartment",
				Price:      50000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now.Add(1 * time.Hour),
			},
			isDuplicate: true,
		},
		{
			name: "not duplicate",
			point: &models.CostDataPoint{
				Category:   "Housing",
				ItemName:   "1BR Apartment",
				Price:      80000,
				Location:   models.Location{Emirate: "Dubai"},
				Source:     "Bayut",
				RecordedAt: now,
			},
			isDuplicate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dc.IsDuplicate(tt.point, existing)
			if result != tt.isDuplicate {
				t.Errorf("Expected IsDuplicate: %v, got: %v", tt.isDuplicate, result)
			}
		})
	}
}

func TestDeduplicateDataPoints(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	points := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now.Add(1 * time.Hour),
		},
		{
			Category:   "Housing",
			ItemName:   "1BR Apartment",
			Price:      80000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
	}

	deduplicated := dc.DeduplicateDataPoints(points)

	if len(deduplicated) != 2 {
		t.Errorf("Expected 2 unique points, got %d", len(deduplicated))
	}
}

func TestDeduplicateDataPoints_EmptySlice(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	deduplicated := dc.DeduplicateDataPoints([]*models.CostDataPoint{})

	if len(deduplicated) != 0 {
		t.Error("Expected empty result for empty input")
	}
}

func TestGenerateDuplicateReport(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	now := time.Now()
	points := []*models.CostDataPoint{
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
		{
			Category:   "Housing",
			ItemName:   "Studio Apartment",
			Price:      50000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now.Add(1 * time.Hour),
		},
		{
			Category:   "Housing",
			ItemName:   "1BR Apartment",
			Price:      80000,
			Location:   models.Location{Emirate: "Dubai"},
			Source:     "Bayut",
			RecordedAt: now,
		},
	}

	report := dc.GenerateDuplicateReport(points)

	if report.TotalPoints != 3 {
		t.Errorf("Expected 3 total points, got %d", report.TotalPoints)
	}

	if report.DuplicateGroups == 0 {
		t.Error("Expected at least one duplicate group")
	}

	if report.DuplicateRate == 0 {
		t.Error("Expected non-zero duplicate rate")
	}

	if len(report.Groups) == 0 {
		t.Error("Expected groups to be included in report")
	}
}

func TestRoundPrice(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{100.123, 100.12},
		{100.126, 100.13},
		{100.0, 100.0},
		{99.999, 100.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := roundPrice(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGenerateSignature(t *testing.T) {
	dc := NewDuplicateChecker(24 * time.Hour)

	dp1 := &models.CostDataPoint{
		Category: "Housing",
		ItemName: "Studio Apartment",
		Price:    50000,
		Location: models.Location{Emirate: "Dubai"},
		Source:   "Bayut",
	}

	dp2 := &models.CostDataPoint{
		Category: "Housing",
		ItemName: "Studio Apartment",
		Price:    50000,
		Location: models.Location{Emirate: "Dubai"},
		Source:   "Bayut",
	}

	sig1 := dc.generateSignature(dp1)
	sig2 := dc.generateSignature(dp2)

	if sig1 != sig2 {
		t.Error("Expected identical signatures for identical data points")
	}

	// Different data point should have different signature
	dp3 := &models.CostDataPoint{
		Category: "Housing",
		ItemName: "1BR Apartment",
		Price:    80000,
		Location: models.Location{Emirate: "Dubai"},
		Source:   "Bayut",
	}

	sig3 := dc.generateSignature(dp3)
	if sig1 == sig3 {
		t.Error("Expected different signatures for different data points")
	}
}
