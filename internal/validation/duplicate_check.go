package validation

import (
	"crypto/sha256"
	"fmt"
	"math"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// DuplicateChecker detects duplicate or near-duplicate data points
type DuplicateChecker struct {
	timeWindow      time.Duration // Time window to consider for duplicates
	priceThreshold  float64       // Percentage threshold for price similarity (0.0-1.0)
}

// NewDuplicateChecker creates a new duplicate checker
func NewDuplicateChecker(timeWindow time.Duration) *DuplicateChecker {
	return &DuplicateChecker{
		timeWindow:     timeWindow,
		priceThreshold: 0.05, // 5% price difference threshold
	}
}

// DuplicateGroup represents a group of duplicate data points
type DuplicateGroup struct {
	Indices      []int
	DataPoints   []*models.CostDataPoint
	Signature    string
	SimilarityScore float64
}

// DetectDuplicates finds duplicate groups in the data
func (dc *DuplicateChecker) DetectDuplicates(points []*models.CostDataPoint) []DuplicateGroup {
	if len(points) < 2 {
		return []DuplicateGroup{}
	}

	// Create signature map
	signatureMap := make(map[string][]int)

	for i, point := range points {
		sig := dc.generateSignature(point)
		signatureMap[sig] = append(signatureMap[sig], i)
	}

	// Find duplicate groups (signature match)
	duplicateGroups := make([]DuplicateGroup, 0)
	for sig, indices := range signatureMap {
		if len(indices) > 1 {
			duplicateGroup := DuplicateGroup{
				Indices:         indices,
				DataPoints:      make([]*models.CostDataPoint, len(indices)),
				Signature:       sig,
				SimilarityScore: 1.0, // Exact match
			}
			for i, idx := range indices {
				duplicateGroup.DataPoints[i] = points[idx]
			}
			duplicateGroups = append(duplicateGroups, duplicateGroup)
		}
	}

	// Also check for near-duplicates (fuzzy matching)
	nearDuplicates := dc.findNearDuplicates(points)
	duplicateGroups = append(duplicateGroups, nearDuplicates...)

	return duplicateGroups
}

// generateSignature creates a unique signature for a data point
func (dc *DuplicateChecker) generateSignature(dp *models.CostDataPoint) string {
	// Use key fields to generate signature
	sig := fmt.Sprintf("%s|%s|%s|%.2f|%s",
		dp.Category,
		dp.ItemName,
		dp.Location.Emirate,
		roundPrice(dp.Price),
		dp.Source,
	)

	// Hash the signature
	hash := sha256.Sum256([]byte(sig))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for shorter signature
}

// findNearDuplicates finds data points that are similar but not exact duplicates
func (dc *DuplicateChecker) findNearDuplicates(points []*models.CostDataPoint) []DuplicateGroup {
	nearDuplicates := make([]DuplicateGroup, 0)
	visited := make(map[int]bool)

	for i := 0; i < len(points); i++ {
		if visited[i] {
			continue
		}

		group := []int{i}
		for j := i + 1; j < len(points); j++ {
			if visited[j] {
				continue
			}

			if dc.areSimilar(points[i], points[j]) {
				group = append(group, j)
				visited[j] = true
			}
		}

		if len(group) > 1 {
			similarity := dc.calculateSimilarity(points[group[0]], points[group[1]])
			duplicateGroup := DuplicateGroup{
				Indices:         group,
				DataPoints:      make([]*models.CostDataPoint, len(group)),
				Signature:       "near-duplicate",
				SimilarityScore: similarity,
			}
			for k, idx := range group {
				duplicateGroup.DataPoints[k] = points[idx]
			}
			nearDuplicates = append(nearDuplicates, duplicateGroup)
			visited[i] = true
		}
	}

	return nearDuplicates
}

// areSimilar checks if two data points are similar enough to be considered near-duplicates
func (dc *DuplicateChecker) areSimilar(dp1, dp2 *models.CostDataPoint) bool {
	// Same category
	if dp1.Category != dp2.Category {
		return false
	}

	// Same or very similar item name (exact match for now, could use fuzzy matching)
	if dp1.ItemName != dp2.ItemName {
		return false
	}

	// Same location
	if dp1.Location.Emirate != dp2.Location.Emirate {
		return false
	}

	// Same source
	if dp1.Source != dp2.Source {
		return false
	}

	// Similar price (within threshold)
	priceDiff := math.Abs(dp1.Price-dp2.Price) / math.Max(dp1.Price, dp2.Price)
	if priceDiff > dc.priceThreshold {
		return false
	}

	// Within time window
	timeDiff := dp1.RecordedAt.Sub(dp2.RecordedAt)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > dc.timeWindow {
		return false
	}

	return true
}

// calculateSimilarity calculates similarity score between two data points (0-1)
func (dc *DuplicateChecker) calculateSimilarity(dp1, dp2 *models.CostDataPoint) float64 {
	score := 0.0
	weights := 0.0

	// Category match (weight: 0.2)
	if dp1.Category == dp2.Category {
		score += 0.2
	}
	weights += 0.2

	// Item name match (weight: 0.3)
	if dp1.ItemName == dp2.ItemName {
		score += 0.3
	}
	weights += 0.3

	// Location match (weight: 0.2)
	if dp1.Location.Emirate == dp2.Location.Emirate {
		score += 0.2
	}
	weights += 0.2

	// Price similarity (weight: 0.2)
	priceDiff := math.Abs(dp1.Price-dp2.Price) / math.Max(dp1.Price, dp2.Price)
	priceSimilarity := 1.0 - math.Min(priceDiff, 1.0)
	score += 0.2 * priceSimilarity
	weights += 0.2

	// Source match (weight: 0.1)
	if dp1.Source == dp2.Source {
		score += 0.1
	}
	weights += 0.1

	return score / weights
}

// roundPrice rounds price to 2 decimal places for signature generation
func roundPrice(price float64) float64 {
	return math.Round(price*100) / 100
}

// IsDuplicate checks if a single data point is a duplicate of any in a list
func (dc *DuplicateChecker) IsDuplicate(point *models.CostDataPoint, existing []*models.CostDataPoint) bool {
	pointSig := dc.generateSignature(point)

	for _, existingPoint := range existing {
		existingSig := dc.generateSignature(existingPoint)
		if pointSig == existingSig {
			return true
		}

		if dc.areSimilar(point, existingPoint) {
			return true
		}
	}

	return false
}

// DeduplicateDataPoints removes duplicates from a slice, keeping the first occurrence
func (dc *DuplicateChecker) DeduplicateDataPoints(points []*models.CostDataPoint) []*models.CostDataPoint {
	if len(points) <= 1 {
		return points
	}

	seen := make(map[string]bool)
	deduplicated := make([]*models.CostDataPoint, 0, len(points))

	for _, point := range points {
		sig := dc.generateSignature(point)
		if !seen[sig] {
			seen[sig] = true
			deduplicated = append(deduplicated, point)
		}
	}

	return deduplicated
}

// DuplicateReport contains summary information about duplicates
type DuplicateReport struct {
	TotalPoints      int
	DuplicateGroups  int
	TotalDuplicates  int
	DuplicateRate    float64
	Groups           []DuplicateGroup
}

// GenerateDuplicateReport creates a comprehensive duplicate report
func (dc *DuplicateChecker) GenerateDuplicateReport(points []*models.CostDataPoint) DuplicateReport {
	groups := dc.DetectDuplicates(points)

	totalDuplicates := 0
	for _, group := range groups {
		totalDuplicates += len(group.Indices) - 1 // Exclude one as the original
	}

	rate := 0.0
	if len(points) > 0 {
		rate = float64(totalDuplicates) / float64(len(points))
	}

	return DuplicateReport{
		TotalPoints:     len(points),
		DuplicateGroups: len(groups),
		TotalDuplicates: totalDuplicates,
		DuplicateRate:   rate,
		Groups:          groups,
	}
}
