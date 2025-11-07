package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetEnabledScraperNamesByCategory(t *testing.T) {
	housing := getEnabledScraperNames("housing", 24*time.Hour, 0)
	expectedHousing := []string{
		"bayut",
		"bayut_sharjah",
		"bayut_ajman",
		"bayut_abu_dhabi",
		"dubizzle",
		"dubizzle_sharjah",
		"dubizzle_ajman",
		"dubizzle_abu_dhabi",
		"dubizzle_dubai_bedspace",
		"dubizzle_dubai_roomspace",
	}
	require.Equal(t, expectedHousing, housing)

	utilities := getEnabledScraperNames("utilities", 7*24*time.Hour, 24*time.Hour)
	require.Equal(t, []string{"dewa", "sewa", "aadc"}, utilities)

	rideshare := getEnabledScraperNames("rideshare", 0, 0)
	require.Equal(t, []string{"careem"}, rideshare)
}

func TestResolveScraperNames(t *testing.T) {
	input := &BatchScraperWorkflowInput{Category: "housing"}
	resolved := resolveScraperNames(input)
	require.Equal(t, dailyScraperNames(), resolved)

	input = &BatchScraperWorkflowInput{ScraperNames: []string{"  Bayut ", "DUBIZZLE", "bayut"}}
	resolved = resolveScraperNames(input)
	require.Equal(t, []string{"bayut", "dubizzle"}, resolved)
}

func TestPeriodicNameHelpers(t *testing.T) {
	require.Equal(t, dailyScraperNames(), resolveScraperNames(&BatchScraperWorkflowInput{Category: "housing"}))
	require.Equal(t, []string{"dewa", "sewa", "aadc", "rta"}, weeklyScraperNames())
	require.Equal(t, []string{"careem"}, monthlyScraperNames())
}
