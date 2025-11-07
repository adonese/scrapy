package validation

import (
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func createTestPoints(prices []float64) []*models.CostDataPoint {
	points := make([]*models.CostDataPoint, len(prices))
	for i, price := range prices {
		points[i] = &models.CostDataPoint{
			Category:   "Housing",
			ItemName:   "Test Item",
			Price:      price,
			Location:   models.Location{Emirate: "Dubai"},
			RecordedAt: time.Now(),
			Source:     "TestSource",
		}
	}
	return points
}

func TestNewOutlierDetector(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)
	if detector == nil {
		t.Fatal("Expected detector to be created")
	}

	if detector.method != DetectionMethodIQR {
		t.Error("Expected IQR method")
	}

	if detector.threshold != 1.5 {
		t.Error("Expected threshold 1.5")
	}
}

func TestDetectOutliers_IQR(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	// Create dataset with outliers
	prices := []float64{100, 110, 105, 108, 102, 500, 95, 103, 107}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) == 0 {
		t.Error("Expected outliers to be detected")
	}

	// Check if the outlier (500) was detected
	foundOutlier := false
	for _, idx := range outliers {
		if points[idx].Price == 500 {
			foundOutlier = true
			break
		}
	}

	if !foundOutlier {
		t.Error("Expected price 500 to be detected as outlier")
	}
}

func TestDetectOutliers_ZScore(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodZScore, 3.0)

	prices := []float64{100, 110, 105, 108, 102, 600, 95, 103, 107}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) == 0 {
		t.Error("Expected outliers to be detected")
	}
}

func TestDetectOutliers_ModifiedZScore(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodModifiedZScore, 3.5)

	prices := []float64{100, 110, 105, 108, 102, 700, 95, 103, 107}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) == 0 {
		t.Error("Expected outliers to be detected")
	}
}

func TestDetectOutliers_SmallDataset(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	// Dataset with less than 3 points
	prices := []float64{100, 110}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) != 0 {
		t.Error("Expected no outliers for small dataset")
	}
}

func TestDetectOutliers_NoOutliers(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	// Dataset with no outliers
	prices := []float64{100, 102, 105, 108, 103, 107, 104, 106, 101}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) != 0 {
		t.Errorf("Expected no outliers, got %d", len(outliers))
	}
}

func TestDetectOutliers_MultipleCategories(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	points := []*models.CostDataPoint{
		{Category: "Housing", Price: 100, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
		{Category: "Housing", Price: 110, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
		{Category: "Housing", Price: 105, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
		{Category: "Housing", Price: 1000, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"}, // Outlier
		{Category: "Utilities", Price: 50, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
		{Category: "Utilities", Price: 55, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
		{Category: "Utilities", Price: 52, Location: models.Location{Emirate: "Dubai"}, RecordedAt: time.Now(), Source: "Test"},
	}

	outliers := detector.DetectOutliers(points)

	// Should detect the housing outlier
	if len(outliers) == 0 {
		t.Error("Expected outliers to be detected")
	}
}

func TestDetectOutliersWithInfo(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	prices := []float64{100, 110, 105, 108, 102, 500, 95, 103, 107}
	points := createTestPoints(prices)

	infos := detector.DetectOutliersWithInfo(points)

	if len(infos) == 0 {
		t.Error("Expected outlier info to be returned")
	}

	for _, info := range infos {
		if info.DataPoint == nil {
			t.Error("Expected data point in outlier info")
		}
		if info.Method != DetectionMethodIQR {
			t.Error("Expected IQR method in info")
		}
		if info.Reason == "" {
			t.Error("Expected reason in outlier info")
		}
	}
}

func TestQuartiles(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	q1, q3 := quartiles(data)

	// Q1 should be around 3, Q3 around 8
	if q1 < 2 || q1 > 4 {
		t.Errorf("Unexpected Q1 value: %f", q1)
	}
	if q3 < 7 || q3 > 9 {
		t.Errorf("Unexpected Q3 value: %f", q3)
	}
}

func TestPercentile(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	p50 := percentile(data, 0.5) // Median
	if p50 != 5.5 {
		t.Errorf("Expected median 5.5, got %f", p50)
	}

	p100 := percentile(data, 1.0) // Maximum
	if p100 != 10 {
		t.Errorf("Expected max 10, got %f", p100)
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{"simple", []float64{1, 2, 3, 4, 5}, 3},
		{"zeros", []float64{0, 0, 0}, 0},
		{"empty", []float64{}, 0},
		{"single", []float64{10}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.data)
			if result != tt.expected {
				t.Errorf("Expected mean %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{"odd count", []float64{1, 2, 3, 4, 5}, 3},
		{"even count", []float64{1, 2, 3, 4}, 2.5},
		{"single", []float64{5}, 5},
		{"empty", []float64{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := median(tt.data)
			if result != tt.expected {
				t.Errorf("Expected median %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestStandardDeviation(t *testing.T) {
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	m := mean(data)
	stdDev := standardDeviation(data, m)

	// Standard deviation should be approximately 2
	if stdDev < 1.5 || stdDev > 2.5 {
		t.Errorf("Unexpected standard deviation: %f", stdDev)
	}

	// Test with all same values
	sameData := []float64{5, 5, 5, 5}
	m = mean(sameData)
	stdDev = standardDeviation(sameData, m)
	if stdDev != 0 {
		t.Errorf("Expected standard deviation 0 for same values, got %f", stdDev)
	}
}

func TestMedianAbsoluteDeviation(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	m := median(data)
	mad := medianAbsoluteDeviation(data, m)

	if mad <= 0 {
		t.Errorf("Expected positive MAD, got %f", mad)
	}

	// Test with all same values
	sameData := []float64{5, 5, 5, 5}
	m = median(sameData)
	mad = medianAbsoluteDeviation(sameData, m)
	if mad != 0 {
		t.Errorf("Expected MAD 0 for same values, got %f", mad)
	}
}

func TestDetectOutliers_AllSameValues(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	// Dataset with all same values
	prices := []float64{100, 100, 100, 100, 100}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	// Should not detect any outliers when all values are the same
	if len(outliers) != 0 {
		t.Errorf("Expected no outliers for identical values, got %d", len(outliers))
	}
}

func TestDetectOutliers_MultipleOutliers(t *testing.T) {
	detector := NewOutlierDetector(DetectionMethodIQR, 1.5)

	// Dataset with multiple outliers
	prices := []float64{100, 110, 105, 108, 1000, 102, 5, 95, 103, 107}
	points := createTestPoints(prices)

	outliers := detector.DetectOutliers(points)

	if len(outliers) < 2 {
		t.Error("Expected multiple outliers to be detected")
	}
}
