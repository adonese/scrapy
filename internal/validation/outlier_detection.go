package validation

import (
	"math"
	"sort"

	"github.com/adonese/cost-of-living/internal/models"
)

// DetectionMethod defines the outlier detection method
type DetectionMethod int

const (
	// DetectionMethodIQR uses Interquartile Range method
	DetectionMethodIQR DetectionMethod = iota
	// DetectionMethodZScore uses Z-Score method
	DetectionMethodZScore
	// DetectionMethodModifiedZScore uses Modified Z-Score method
	DetectionMethodModifiedZScore
)

// OutlierDetector detects statistical outliers in data
type OutlierDetector struct {
	method    DetectionMethod
	threshold float64
}

// NewOutlierDetector creates a new outlier detector
func NewOutlierDetector(method DetectionMethod, threshold float64) *OutlierDetector {
	return &OutlierDetector{
		method:    method,
		threshold: threshold,
	}
}

// DetectOutliers returns the indices of outlier data points
func (od *OutlierDetector) DetectOutliers(points []*models.CostDataPoint) []int {
	if len(points) < 3 {
		return []int{} // Need at least 3 points for outlier detection
	}

	// Group by category for more accurate outlier detection
	categoryGroups := make(map[string][]*indexedPoint)
	for i, point := range points {
		categoryGroups[point.Category] = append(categoryGroups[point.Category], &indexedPoint{
			index: i,
			point: point,
		})
	}

	outliers := make([]int, 0)
	for _, group := range categoryGroups {
		if len(group) < 3 {
			continue // Skip small groups
		}

		var groupOutliers []int
		switch od.method {
		case DetectionMethodIQR:
			groupOutliers = od.detectIQROutliers(group)
		case DetectionMethodZScore:
			groupOutliers = od.detectZScoreOutliers(group)
		case DetectionMethodModifiedZScore:
			groupOutliers = od.detectModifiedZScoreOutliers(group)
		}
		outliers = append(outliers, groupOutliers...)
	}

	return outliers
}

// indexedPoint pairs a data point with its original index
type indexedPoint struct {
	index int
	point *models.CostDataPoint
}

// detectIQROutliers uses the IQR method to detect outliers
func (od *OutlierDetector) detectIQROutliers(points []*indexedPoint) []int {
	prices := make([]float64, len(points))
	for i, p := range points {
		prices[i] = p.point.Price
	}

	q1, q3 := quartiles(prices)
	iqr := q3 - q1
	lowerBound := q1 - od.threshold*iqr
	upperBound := q3 + od.threshold*iqr

	outliers := make([]int, 0)
	for _, p := range points {
		if p.point.Price < lowerBound || p.point.Price > upperBound {
			outliers = append(outliers, p.index)
		}
	}

	return outliers
}

// detectZScoreOutliers uses the Z-Score method to detect outliers
func (od *OutlierDetector) detectZScoreOutliers(points []*indexedPoint) []int {
	prices := make([]float64, len(points))
	for i, p := range points {
		prices[i] = p.point.Price
	}

	mean := mean(prices)
	stdDev := standardDeviation(prices, mean)

	if stdDev == 0 {
		return []int{} // All values are the same
	}

	outliers := make([]int, 0)
	for _, p := range points {
		zScore := math.Abs((p.point.Price - mean) / stdDev)
		if zScore > od.threshold {
			outliers = append(outliers, p.index)
		}
	}

	return outliers
}

// detectModifiedZScoreOutliers uses the Modified Z-Score method (more robust)
func (od *OutlierDetector) detectModifiedZScoreOutliers(points []*indexedPoint) []int {
	prices := make([]float64, len(points))
	for i, p := range points {
		prices[i] = p.point.Price
	}

	median := median(prices)
	mad := medianAbsoluteDeviation(prices, median)

	if mad == 0 {
		return []int{} // All values are the same
	}

	outliers := make([]int, 0)
	for _, p := range points {
		modifiedZScore := math.Abs(0.6745 * (p.point.Price - median) / mad)
		if modifiedZScore > od.threshold {
			outliers = append(outliers, p.index)
		}
	}

	return outliers
}

// Statistical helper functions

// quartiles calculates Q1 and Q3
func quartiles(data []float64) (float64, float64) {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	n := len(sorted)
	q1 := percentile(sorted, 0.25)
	q3 := percentile(sorted, 0.75)

	return q1, q3
}

// percentile calculates the given percentile
func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}

	index := p * float64(n-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// mean calculates the arithmetic mean
func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// median calculates the median value
func median(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// standardDeviation calculates the standard deviation
func standardDeviation(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sumSquares := 0.0
	for _, v := range data {
		diff := v - mean
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares / float64(len(data)))
}

// medianAbsoluteDeviation calculates the MAD
func medianAbsoluteDeviation(data []float64, median float64) float64 {
	deviations := make([]float64, len(data))
	for i, v := range data {
		deviations[i] = math.Abs(v - median)
	}
	return median(deviations)
}

// OutlierInfo contains information about an outlier
type OutlierInfo struct {
	Index      int
	DataPoint  *models.CostDataPoint
	Score      float64
	Method     DetectionMethod
	Reason     string
}

// DetectOutliersWithInfo returns detailed information about outliers
func (od *OutlierDetector) DetectOutliersWithInfo(points []*models.CostDataPoint) []OutlierInfo {
	indices := od.DetectOutliers(points)
	infos := make([]OutlierInfo, 0, len(indices))

	for _, idx := range indices {
		if idx < len(points) {
			infos = append(infos, OutlierInfo{
				Index:     idx,
				DataPoint: points[idx],
				Method:    od.method,
				Reason:    "Statistical outlier detected",
			})
		}
	}

	return infos
}
