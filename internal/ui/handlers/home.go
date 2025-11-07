package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/adonese/cost-of-living/internal/handlers/dto"
	"github.com/adonese/cost-of-living/internal/services/estimator"
	"github.com/adonese/cost-of-living/internal/ui/render"
	ui "github.com/adonese/cost-of-living/web/ui"
)

// HomeHandler renders the landing page and HTMX fragments.
type HomeHandler struct {
	estimator *estimator.Service
	validate  *validator.Validate
}

// NewHomeHandler builds a HomeHandler instance.
func NewHomeHandler(estimatorService *estimator.Service) *HomeHandler {
	return &HomeHandler{
		estimator: estimatorService,
		validate:  validator.New(),
	}
}

// Index renders the home/landing page populated with a default persona estimate.
func (h *HomeHandler) Index(c echo.Context) error {
	result, err := h.estimator.Estimate(c.Request().Context(), defaultPersona())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return render.Component(c, http.StatusOK, ui.HomePage(result))
}

// EstimatePartial recomputes the estimator panel for HTMX interactions.
func (h *HomeHandler) EstimatePartial(c echo.Context) error {
	var req dto.EstimateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid form body")
	}
	if err := h.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.estimator.Estimate(c.Request().Context(), req.ToPersona())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return render.Component(c, http.StatusOK, ui.EstimatePanel(result))
}

func defaultPersona() estimator.PersonaInput {
	return estimator.PersonaInput{
		Adults:        2,
		Children:      0,
		Bedrooms:      2,
		HousingType:   estimator.HousingApartment,
		Lifestyle:     estimator.LifestyleModerate,
		Emirate:       "Dubai",
		TransportMode: estimator.TransportMixed,
	}
}
