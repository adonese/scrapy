package validation

import (
	"testing"
	"time"
)

func TestNewFreshnessChecker(t *testing.T) {
	fc := NewFreshnessChecker()
	if fc == nil {
		t.Fatal("Expected freshness checker to be created")
	}

	if fc.defaultMaxAge != 7*24*time.Hour {
		t.Error("Expected default max age to be 7 days")
	}

	if len(fc.maxAgeBySource) == 0 {
		t.Error("Expected source-specific max ages to be configured")
	}
}

func TestCheckFreshness(t *testing.T) {
	fc := NewFreshnessChecker()
	now := time.Now()

	tests := []struct {
		name     string
		source   string
		recorded time.Time
		expected FreshnessStatus
	}{
		{
			name:     "fresh bayut data",
			source:   "Bayut",
			recorded: now.Add(-2 * 24 * time.Hour), // 2 days ago
			expected: FreshnessFresh,
		},
		{
			name:     "stale bayut data",
			source:   "Bayut",
			recorded: now.Add(-12 * 24 * time.Hour), // 12 days ago (1.5x max age)
			expected: FreshnessStale,
		},
		{
			name:     "expired bayut data",
			source:   "Bayut",
			recorded: now.Add(-25 * 24 * time.Hour), // 25 days ago (3x max age)
			expected: FreshnessExpired,
		},
		{
			name:     "fresh rta data",
			source:   "RTA",
			recorded: now.Add(-12 * time.Hour), // 12 hours ago
			expected: FreshnessFresh,
		},
		{
			name:     "stale rta data",
			source:   "RTA",
			recorded: now.Add(-2 * 24 * time.Hour), // 2 days ago
			expected: FreshnessStale,
		},
		{
			name:     "unknown source",
			source:   "UnknownSource",
			recorded: now.Add(-3 * 24 * time.Hour), // 3 days ago
			expected: FreshnessFresh, // Uses default 7 days
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := fc.CheckFreshness(tt.source, tt.recorded)
			if status != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, status)
			}
		})
	}
}

func TestFreshnessStatus_String(t *testing.T) {
	tests := []struct {
		status   FreshnessStatus
		expected string
	}{
		{FreshnessFresh, "FRESH"},
		{FreshnessStale, "STALE"},
		{FreshnessExpired, "EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.status.String())
			}
		})
	}
}

func TestGetMaxAge(t *testing.T) {
	fc := NewFreshnessChecker()

	tests := []struct {
		source   string
		expected time.Duration
	}{
		{"Bayut", 7 * 24 * time.Hour},
		{"DEWA", 30 * 24 * time.Hour},
		{"RTA", 24 * time.Hour},
		{"UnknownSource", 7 * 24 * time.Hour}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			maxAge := fc.GetMaxAge(tt.source)
			if maxAge != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, maxAge)
			}
		})
	}
}

func TestSetMaxAge(t *testing.T) {
	fc := NewFreshnessChecker()

	customDuration := 15 * 24 * time.Hour
	fc.SetMaxAge("CustomSource", customDuration)

	maxAge := fc.GetMaxAge("CustomSource")
	if maxAge != customDuration {
		t.Errorf("Expected %v, got %v", customDuration, maxAge)
	}
}

func TestSetDefaultMaxAge(t *testing.T) {
	fc := NewFreshnessChecker()

	customDefault := 14 * 24 * time.Hour
	fc.SetDefaultMaxAge(customDefault)

	if fc.defaultMaxAge != customDefault {
		t.Errorf("Expected default max age %v, got %v", customDefault, fc.defaultMaxAge)
	}
}

func TestGenerateFreshnessReport(t *testing.T) {
	fc := NewFreshnessChecker()
	now := time.Now()

	recordedTimes := []time.Time{
		now.Add(-1 * 24 * time.Hour),  // Fresh
		now.Add(-2 * 24 * time.Hour),  // Fresh
		now.Add(-12 * 24 * time.Hour), // Stale (for 7-day max age)
		now.Add(-25 * 24 * time.Hour), // Expired
	}

	report := fc.GenerateFreshnessReport("Bayut", recordedTimes)

	if report.Source != "Bayut" {
		t.Errorf("Expected source Bayut, got %s", report.Source)
	}

	if report.TotalPoints != 4 {
		t.Errorf("Expected 4 total points, got %d", report.TotalPoints)
	}

	if report.FreshCount != 2 {
		t.Errorf("Expected 2 fresh points, got %d", report.FreshCount)
	}

	if report.StaleCount != 1 {
		t.Errorf("Expected 1 stale point, got %d", report.StaleCount)
	}

	if report.ExpiredCount != 1 {
		t.Errorf("Expected 1 expired point, got %d", report.ExpiredCount)
	}

	expectedRate := 2.0 / 4.0
	if report.FreshnessRate != expectedRate {
		t.Errorf("Expected freshness rate %f, got %f", expectedRate, report.FreshnessRate)
	}
}

func TestGenerateFreshnessReport_Empty(t *testing.T) {
	fc := NewFreshnessChecker()

	report := fc.GenerateFreshnessReport("Bayut", []time.Time{})

	if report.TotalPoints != 0 {
		t.Errorf("Expected 0 total points, got %d", report.TotalPoints)
	}

	if report.FreshnessRate != 0 {
		t.Errorf("Expected 0 freshness rate, got %f", report.FreshnessRate)
	}
}

func TestGetRecommendedUpdateFrequency(t *testing.T) {
	fc := NewFreshnessChecker()

	tests := []struct {
		source   string
		expected time.Duration
	}{
		{"Bayut", 3*24*time.Hour + 12*time.Hour},    // 7 days / 2 = 3.5 days
		{"RTA", 12 * time.Hour},                     // 1 day / 2 = 12 hours
		{"DEWA", 15 * 24 * time.Hour},               // 30 days / 2 = 15 days
		{"UnknownSource", 3*24*time.Hour + 12*time.Hour}, // 7 days / 2 = 3.5 days (default)
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			freq := fc.GetRecommendedUpdateFrequency(tt.source)
			if freq != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, freq)
			}
		})
	}
}

func TestNeedsUpdate(t *testing.T) {
	fc := NewFreshnessChecker()
	now := time.Now()

	tests := []struct {
		name        string
		source      string
		lastUpdate  time.Time
		needsUpdate bool
	}{
		{
			name:        "bayut needs update",
			source:      "Bayut",
			lastUpdate:  now.Add(-5 * 24 * time.Hour), // More than 3.5 days ago
			needsUpdate: true,
		},
		{
			name:        "bayut doesn't need update",
			source:      "Bayut",
			lastUpdate:  now.Add(-2 * 24 * time.Hour), // Less than 3.5 days ago
			needsUpdate: false,
		},
		{
			name:        "rta needs update",
			source:      "RTA",
			lastUpdate:  now.Add(-18 * time.Hour), // More than 12 hours ago
			needsUpdate: true,
		},
		{
			name:        "rta doesn't need update",
			source:      "RTA",
			lastUpdate:  now.Add(-6 * time.Hour), // Less than 12 hours ago
			needsUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fc.NeedsUpdate(tt.source, tt.lastUpdate)
			if result != tt.needsUpdate {
				t.Errorf("Expected needs update: %v, got: %v", tt.needsUpdate, result)
			}
		})
	}
}

func TestGenerateFreshnessMap(t *testing.T) {
	fc := NewFreshnessChecker()
	now := time.Now()

	sourceData := map[string]time.Time{
		"Bayut": now.Add(-2 * 24 * time.Hour),  // Fresh
		"DEWA":  now.Add(-50 * 24 * time.Hour), // Stale
		"RTA":   now.Add(-5 * 24 * time.Hour),  // Expired
	}

	freshnessMap := fc.GenerateFreshnessMap(sourceData)

	if len(freshnessMap) != 3 {
		t.Errorf("Expected 3 entries in freshness map, got %d", len(freshnessMap))
	}

	if freshnessMap["Bayut"] != FreshnessFresh {
		t.Error("Expected Bayut to be fresh")
	}

	if freshnessMap["DEWA"] != FreshnessStale {
		t.Error("Expected DEWA to be stale")
	}

	if freshnessMap["RTA"] != FreshnessExpired {
		t.Error("Expected RTA to be expired")
	}
}

func TestFreshnessMap_GetStaleSources(t *testing.T) {
	fm := FreshnessMap{
		"Bayut": FreshnessFresh,
		"DEWA":  FreshnessStale,
		"RTA":   FreshnessExpired,
		"SEWA":  FreshnessFresh,
	}

	staleSources := fm.GetStaleSources()

	if len(staleSources) != 2 {
		t.Errorf("Expected 2 stale sources, got %d", len(staleSources))
	}

	// Check that DEWA and RTA are in the list
	hasDEWA := false
	hasRTA := false
	for _, source := range staleSources {
		if source == "DEWA" {
			hasDEWA = true
		}
		if source == "RTA" {
			hasRTA = true
		}
	}

	if !hasDEWA || !hasRTA {
		t.Error("Expected DEWA and RTA in stale sources list")
	}
}

func TestFreshnessMap_GetFreshSources(t *testing.T) {
	fm := FreshnessMap{
		"Bayut": FreshnessFresh,
		"DEWA":  FreshnessStale,
		"RTA":   FreshnessExpired,
		"SEWA":  FreshnessFresh,
	}

	freshSources := fm.GetFreshSources()

	if len(freshSources) != 2 {
		t.Errorf("Expected 2 fresh sources, got %d", len(freshSources))
	}

	// Check that Bayut and SEWA are in the list
	hasBayut := false
	hasSEWA := false
	for _, source := range freshSources {
		if source == "Bayut" {
			hasBayut = true
		}
		if source == "SEWA" {
			hasSEWA = true
		}
	}

	if !hasBayut || !hasSEWA {
		t.Error("Expected Bayut and SEWA in fresh sources list")
	}
}

func TestConfigureSourceMaxAges(t *testing.T) {
	fc := NewFreshnessChecker()

	// Check that key sources are configured
	expectedAges := map[string]time.Duration{
		"Bayut":    7 * 24 * time.Hour,
		"Dubizzle": 7 * 24 * time.Hour,
		"DEWA":     30 * 24 * time.Hour,
		"SEWA":     30 * 24 * time.Hour,
		"AADC":     30 * 24 * time.Hour,
		"RTA":      24 * time.Hour,
		"Careem":   24 * time.Hour,
		"KHDA":     365 * 24 * time.Hour,
	}

	for source, expectedAge := range expectedAges {
		t.Run(source, func(t *testing.T) {
			maxAge := fc.GetMaxAge(source)
			if maxAge == 0 {
				t.Errorf("Expected %s to have configured max age", source)
			}
			if maxAge != expectedAge {
				t.Errorf("Expected %s to have max age %v, got %v", source, expectedAge, maxAge)
			}
		})
	}
}
