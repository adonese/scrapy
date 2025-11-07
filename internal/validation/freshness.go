package validation

import (
	"time"
)

// FreshnessStatus represents the freshness status of data
type FreshnessStatus int

const (
	// FreshnessFresh indicates data is current
	FreshnessFresh FreshnessStatus = iota
	// FreshnessStale indicates data is outdated
	FreshnessStale
	// FreshnessExpired indicates data is too old to be useful
	FreshnessExpired
)

// String returns the string representation of freshness status
func (fs FreshnessStatus) String() string {
	switch fs {
	case FreshnessFresh:
		return "FRESH"
	case FreshnessStale:
		return "STALE"
	case FreshnessExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

// FreshnessChecker checks data freshness based on source-specific rules
type FreshnessChecker struct {
	maxAgeBySource map[string]time.Duration
	defaultMaxAge  time.Duration
}

// NewFreshnessChecker creates a new freshness checker
func NewFreshnessChecker() *FreshnessChecker {
	fc := &FreshnessChecker{
		maxAgeBySource: make(map[string]time.Duration),
		defaultMaxAge:  7 * 24 * time.Hour, // Default: 7 days
	}

	// Configure source-specific max ages
	fc.configureSourceMaxAges()

	return fc
}

// configureSourceMaxAges sets up max age for different sources
func (fc *FreshnessChecker) configureSourceMaxAges() {
	// Housing data changes less frequently
	fc.maxAgeBySource["Bayut"] = 7 * 24 * time.Hour        // 7 days
	fc.maxAgeBySource["Dubizzle"] = 7 * 24 * time.Hour     // 7 days
	fc.maxAgeBySource["PropertyFinder"] = 7 * 24 * time.Hour // 7 days

	// Utility rates change seasonally or monthly
	fc.maxAgeBySource["DEWA"] = 30 * 24 * time.Hour  // 30 days
	fc.maxAgeBySource["SEWA"] = 30 * 24 * time.Hour  // 30 days
	fc.maxAgeBySource["AADC"] = 30 * 24 * time.Hour  // 30 days
	fc.maxAgeBySource["ADDC"] = 30 * 24 * time.Hour  // 30 days
	fc.maxAgeBySource["FEWA"] = 30 * 24 * time.Hour  // 30 days

	// Transportation prices can change daily
	fc.maxAgeBySource["RTA"] = 24 * time.Hour     // 1 day
	fc.maxAgeBySource["Careem"] = 24 * time.Hour  // 1 day
	fc.maxAgeBySource["Uber"] = 24 * time.Hour    // 1 day

	// Food prices change frequently
	fc.maxAgeBySource["Carrefour"] = 7 * 24 * time.Hour  // 7 days
	fc.maxAgeBySource["Lulu"] = 7 * 24 * time.Hour       // 7 days
	fc.maxAgeBySource["Spinneys"] = 7 * 24 * time.Hour   // 7 days

	// Education fees are typically annual
	fc.maxAgeBySource["KHDA"] = 365 * 24 * time.Hour  // 1 year
	fc.maxAgeBySource["ADEK"] = 365 * 24 * time.Hour  // 1 year
}

// CheckFreshness checks if data from a source is fresh
func (fc *FreshnessChecker) CheckFreshness(source string, recordedAt time.Time) FreshnessStatus {
	age := time.Since(recordedAt)

	maxAge, ok := fc.maxAgeBySource[source]
	if !ok {
		maxAge = fc.defaultMaxAge
	}

	// Define stale threshold as 1.5x max age and expired as 3x max age
	staleThreshold := time.Duration(float64(maxAge) * 1.5)
	expiredThreshold := time.Duration(float64(maxAge) * 3.0)

	if age > expiredThreshold {
		return FreshnessExpired
	} else if age > staleThreshold {
		return FreshnessStale
	}

	return FreshnessFresh
}

// GetMaxAge returns the maximum age for a source
func (fc *FreshnessChecker) GetMaxAge(source string) time.Duration {
	if maxAge, ok := fc.maxAgeBySource[source]; ok {
		return maxAge
	}
	return fc.defaultMaxAge
}

// SetMaxAge sets the maximum age for a source
func (fc *FreshnessChecker) SetMaxAge(source string, maxAge time.Duration) {
	fc.maxAgeBySource[source] = maxAge
}

// SetDefaultMaxAge sets the default maximum age
func (fc *FreshnessChecker) SetDefaultMaxAge(maxAge time.Duration) {
	fc.defaultMaxAge = maxAge
}

// FreshnessReport contains freshness statistics
type FreshnessReport struct {
	Source         string
	TotalPoints    int
	FreshCount     int
	StaleCount     int
	ExpiredCount   int
	FreshnessRate  float64
	OldestDataAge  time.Duration
	NewestDataAge  time.Duration
}

// GenerateFreshnessReport creates a freshness report for data from a specific source
func (fc *FreshnessChecker) GenerateFreshnessReport(source string, recordedTimes []time.Time) FreshnessReport {
	report := FreshnessReport{
		Source:      source,
		TotalPoints: len(recordedTimes),
	}

	if len(recordedTimes) == 0 {
		return report
	}

	now := time.Now()
	var oldestAge, newestAge time.Duration

	for i, recordedAt := range recordedTimes {
		status := fc.CheckFreshness(source, recordedAt)
		age := now.Sub(recordedAt)

		// Track oldest and newest
		if i == 0 {
			oldestAge = age
			newestAge = age
		} else {
			if age > oldestAge {
				oldestAge = age
			}
			if age < newestAge {
				newestAge = age
			}
		}

		switch status {
		case FreshnessFresh:
			report.FreshCount++
		case FreshnessStale:
			report.StaleCount++
		case FreshnessExpired:
			report.ExpiredCount++
		}
	}

	report.OldestDataAge = oldestAge
	report.NewestDataAge = newestAge
	report.FreshnessRate = float64(report.FreshCount) / float64(report.TotalPoints)

	return report
}

// GetRecommendedUpdateFrequency returns the recommended update frequency for a source
func (fc *FreshnessChecker) GetRecommendedUpdateFrequency(source string) time.Duration {
	maxAge := fc.GetMaxAge(source)
	// Recommend updating at half the max age to maintain freshness
	return maxAge / 2
}

// NeedsUpdate checks if a source needs updating based on last update time
func (fc *FreshnessChecker) NeedsUpdate(source string, lastUpdate time.Time) bool {
	recommendedFrequency := fc.GetRecommendedUpdateFrequency(source)
	return time.Since(lastUpdate) > recommendedFrequency
}

// FreshnessMap represents freshness status for multiple sources
type FreshnessMap map[string]FreshnessStatus

// GenerateFreshnessMap creates a map of freshness status for multiple sources
func (fc *FreshnessChecker) GenerateFreshnessMap(sourceData map[string]time.Time) FreshnessMap {
	freshnessMap := make(FreshnessMap)

	for source, lastUpdate := range sourceData {
		freshnessMap[source] = fc.CheckFreshness(source, lastUpdate)
	}

	return freshnessMap
}

// GetStaleSources returns a list of sources with stale or expired data
func (fm FreshnessMap) GetStaleSources() []string {
	staleSources := make([]string, 0)

	for source, status := range fm {
		if status == FreshnessStale || status == FreshnessExpired {
			staleSources = append(staleSources, source)
		}
	}

	return staleSources
}

// GetFreshSources returns a list of sources with fresh data
func (fm FreshnessMap) GetFreshSources() []string {
	freshSources := make([]string, 0)

	for source, status := range fm {
		if status == FreshnessFresh {
			freshSources = append(freshSources, source)
		}
	}

	return freshSources
}
