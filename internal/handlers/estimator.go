package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/adonese/cost-of-living/internal/handlers/dto"
	"github.com/adonese/cost-of-living/internal/services/estimator"
)

// EstimatorHandler wires estimator service to Echo.
type EstimatorHandler struct {
	service  *estimator.Service
	validate *validator.Validate
}

// NewEstimatorHandler builds the handler.
func NewEstimatorHandler(service *estimator.Service) *EstimatorHandler {
	return &EstimatorHandler{
		service:  service,
		validate: validator.New(),
	}
}

// Estimate aggregates cost breakdown based on the persona payload.
func (h *EstimatorHandler) Estimate(c echo.Context) error {
	var req dto.EstimateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.service.Estimate(c.Request().Context(), req.ToPersona())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// Summary exposes coverage/freshness metadata to UI cards.
func (h *EstimatorHandler) Summary(c echo.Context) error {
	emirate := c.QueryParam("emirate")
	if emirate == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "emirate query parameter is required")
	}

	snap, err := h.service.Summary(c.Request().Context(), emirate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, snap)
}
