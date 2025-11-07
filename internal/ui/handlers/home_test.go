package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository/mock"
	"github.com/adonese/cost-of-living/internal/services/estimator"
)

func TestHomeHandlerIndex(t *testing.T) {
	h := newTestHomeHandler(t)
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Index(c)
	require.NoError(t, err)
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "Launch estimator")
}

func TestHomeHandlerEstimatePartial(t *testing.T) {
	h := newTestHomeHandler(t)
	e := echo.New()

	payload := `{"adults":1,"children":0,"bedrooms":1,"housing_type":"shared","lifestyle":"budget","emirate":"Dubai","transport_mode":"public"}`

	req := httptest.NewRequest(echo.POST, "/ui/estimate", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EstimatePartial(c)
	require.NoError(t, err)
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "Breakdown")
}

func newTestHomeHandler(t *testing.T) *HomeHandler {
	repo := mock.NewCostDataPointRepository()
	now := time.Now()
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "housing",
		Category:    "Housing",
		SubCategory: "Rent",
		Price:       90000,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "test",
		Confidence:  0.8,
	}))
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "util",
		Category:    "Utilities",
		SubCategory: "Electricity",
		Price:       0.38,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "test",
		Confidence:  0.9,
	}))
	require.NoError(t, repo.Create(context.Background(), &models.CostDataPoint{
		ID:          "transport",
		Category:    "Transportation",
		SubCategory: "Public Transport",
		Price:       4.0,
		Location:    models.Location{Emirate: "Dubai"},
		RecordedAt:  now,
		ValidFrom:   now,
		Source:      "test",
		Confidence:  0.85,
	}))

	svc := estimator.NewService(repo, &estimator.Config{LookbackDays: 120})
	return NewHomeHandler(svc)
}
