package estimator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/models"
	mockrepo "github.com/adonese/cost-of-living/internal/repository/mock"
)

func TestServiceEstimateWithLiveData(t *testing.T) {
	repo := mockrepo.NewCostDataPointRepository()
	now := time.Now()

	// Housing (yearly rent)
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "housing-1",
		Category:    "Housing",
		SubCategory: "Rent",
		Price:       120000,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "test_housing",
		Unit:        "AED",
		Confidence:  0.9,
	}))

	// Utilities
	require.NoError(t, repo.Create(context.Background(), newUtilityPoint("Dubai", "Electricity", 0.38, now)))
	require.NoError(t, repo.Create(context.Background(), newUtilityPoint("Dubai", "Water", 3.1, now)))
	require.NoError(t, repo.Create(context.Background(), newUtilityPoint("Dubai", "Fuel Surcharge", 0.05, now)))

	// Transport - public, taxi, ride sharing components
	require.NoError(t, repo.Create(context.Background(), newTransportPoint("Public Transport", 4.0, now)))
	require.NoError(t, repo.Create(context.Background(), newTransportPoint("Taxi", 2.6, now)))
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "ride-base",
		Category:    "Transportation",
		SubCategory: "Ride Sharing",
		Price:       8.0,
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "careem",
		Unit:        "AED",
		Confidence:  0.9,
		Location:    models.Location{Emirate: "Dubai"},
		Attributes: map[string]interface{}{
			"rate_type": "base_fare",
		},
	}))
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "ride-perkm",
		Category:    "Transportation",
		SubCategory: "Ride Sharing",
		Price:       2.6,
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "careem",
		Unit:        "AED",
		Confidence:  0.9,
		Location:    models.Location{Emirate: "Dubai"},
		Attributes: map[string]interface{}{
			"rate_type": "per_km",
		},
	}))
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "ride-min",
		Category:    "Transportation",
		SubCategory: "Ride Sharing",
		Price:       14.0,
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "careem",
		Unit:        "AED",
		Confidence:  0.9,
		Location:    models.Location{Emirate: "Dubai"},
		Attributes: map[string]interface{}{
			"rate_type": "minimum_fare",
		},
	}))

	svc := NewService(repo, nil)

	persona := PersonaInput{
		Adults:        2,
		Children:      1,
		Bedrooms:      2,
		HousingType:   HousingApartment,
		Lifestyle:     LifestyleModerate,
		Emirate:       "Dubai",
		TransportMode: TransportMixed,
	}

	res, err := svc.Estimate(context.Background(), persona)
	require.NoError(t, err)

	assert.Equal(t, "AED", res.Currency)
	assert.Len(t, res.Breakdown, 5)
	assert.Greater(t, res.MonthlyTotalAED, 0.0)

	housing := findCategory(res.Breakdown, "Housing")
	require.NotNil(t, housing)
	assert.InDelta(t, 10000, housing.MonthlyAED, 3000)
	assert.True(t, housing.SampleSize > 0)

	transport := findCategory(res.Breakdown, "Transportation")
	require.NotNil(t, transport)
	assert.Greater(t, transport.MonthlyAED, 0.0)

	assert.Contains(t, res.Dataset.Coverage, "Housing")
	assert.Contains(t, res.Dataset.Coverage, "Transportation")
}

func TestServiceEstimateFallsBackWhenNoData(t *testing.T) {
	repo := mockrepo.NewCostDataPointRepository()
	svc := NewService(repo, nil)

	persona := PersonaInput{
		Adults:        1,
		Children:      0,
		Bedrooms:      1,
		HousingType:   HousingShared,
		Lifestyle:     LifestyleBudget,
		Emirate:       "Ajman",
		TransportMode: TransportPublic,
	}

	res, err := svc.Estimate(context.Background(), persona)
	require.NoError(t, err)

	for _, cat := range res.Breakdown {
		assert.NotZero(t, cat.MonthlyAED)
	}
	assert.NotEmpty(t, res.Dataset.Warnings)
}

func TestServiceSummary(t *testing.T) {
	repo := mockrepo.NewCostDataPointRepository()
	now := time.Now()
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "housing-1",
		Category:    "Housing",
		SubCategory: "Rent",
		Price:       90000,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "test",
		Confidence:  0.8,
	}))

	svc := NewService(repo, nil)
	snap, err := svc.Summary(context.Background(), "Dubai")
	require.NoError(t, err)
	assert.Greater(t, snap.TotalSamples, 0)
	assert.Contains(t, snap.Coverage, "Housing")
}

func newUtilityPoint(emirate, sub string, price float64, ts time.Time) *models.CostDataPoint {
	return &models.CostDataPoint{
		ID:          "util-" + sub,
		Category:    "Utilities",
		SubCategory: sub,
		Price:       price,
		Location:    models.Location{Emirate: emirate},
		RecordedAt:  ts,
		ValidFrom:   ts,
		Source:      "utility",
		Unit:        "AED",
		Confidence:  0.9,
	}
}

func newTransportPoint(sub string, price float64, ts time.Time) *models.CostDataPoint {
	return &models.CostDataPoint{
		ID:          "trans-" + sub,
		Category:    "Transportation",
		SubCategory: sub,
		Price:       price,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  ts,
		ValidFrom:   ts,
		Source:      "transport",
		Unit:        "AED",
		Confidence:  0.85,
	}
}

func findCategory(items []CategoryEstimate, name string) *CategoryEstimate {
	for i := range items {
		if items[i].Category == name {
			return &items[i]
		}
	}
	return nil
}
