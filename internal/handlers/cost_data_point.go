package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adonese/cost-of-living/internal/handlers/dto"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CostDataPointHandler handles HTTP requests for cost data points
type CostDataPointHandler struct {
	repo     repository.CostDataPointRepository
	validate *validator.Validate
}

// NewCostDataPointHandler creates a new cost data point handler
func NewCostDataPointHandler(repo repository.CostDataPointRepository) *CostDataPointHandler {
	return &CostDataPointHandler{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create handles POST /api/v1/cost-data-points
func (h *CostDataPointHandler) Create(c echo.Context) error {
	var req dto.CreateCostDataPointRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
	}

	// Additional validation
	if req.Price <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Price must be greater than 0")
	}
	if req.Confidence > 0 && (req.Confidence < 0 || req.Confidence > 1) {
		return echo.NewHTTPError(http.StatusBadRequest, "Confidence must be between 0 and 1")
	}

	// Convert DTO to model
	cdp := req.ToModel()

	// Create in database
	if err := h.repo.Create(c.Request().Context(), cdp); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create cost data point")
	}

	// Return created resource
	return c.JSON(http.StatusCreated, dto.FromModel(cdp))
}

// GetByID handles GET /api/v1/cost-data-points/:id
func (h *CostDataPointHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	// Check for recorded_at query parameter
	recordedAtStr := c.QueryParam("recorded_at")
	var recordedAt time.Time
	var err error

	if recordedAtStr != "" {
		// Parse the provided recorded_at timestamp
		recordedAt, err = time.Parse(time.RFC3339, recordedAtStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid recorded_at format, use RFC3339")
		}
	} else {
		// If no recorded_at provided, get the latest record
		// We'll use a query to find the most recent recorded_at for this ID
		filter := repository.ListFilter{
			Limit:  1,
			Offset: 0,
		}
		results, err := h.repo.List(c.Request().Context(), filter)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cost data point")
		}

		// Find the record with matching ID and get its recorded_at
		var found bool
		for _, result := range results {
			if result.ID == id {
				recordedAt = result.RecordedAt
				found = true
				break
			}
		}

		if !found {
			// Try to get more results
			filter.Limit = 100
			results, err = h.repo.List(c.Request().Context(), filter)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cost data point")
			}

			for _, result := range results {
				if result.ID == id {
					recordedAt = result.RecordedAt
					found = true
					break
				}
			}

			if !found {
				return echo.NewHTTPError(http.StatusNotFound, "Cost data point not found")
			}
		}
	}

	// Get the specific record
	cdp, err := h.repo.GetByID(c.Request().Context(), id, recordedAt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Cost data point not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cost data point")
	}

	return c.JSON(http.StatusOK, dto.FromModel(cdp))
}

// List handles GET /api/v1/cost-data-points
func (h *CostDataPointHandler) List(c echo.Context) error {
	filter := repository.ListFilter{}

	// Parse query parameters
	if category := c.QueryParam("category"); category != "" {
		filter.Category = category
	}

	if emirate := c.QueryParam("emirate"); emirate != "" {
		filter.Emirate = emirate
	}

	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format, use RFC3339")
		}
		filter.StartDate = &startDate
	}

	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format, use RFC3339")
		}
		filter.EndDate = &endDate
	}

	// Parse pagination parameters
	limit := 10 // default
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter")
		}
		if l > 100 {
			l = 100 // max limit
		}
		limit = l
	}
	filter.Limit = limit

	offset := 0 // default
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset parameter")
		}
		offset = o
	}
	filter.Offset = offset

	// Get data from repository
	results, err := h.repo.List(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list cost data points")
	}

	// Convert to response DTOs
	responseData := make([]dto.CostDataPointResponse, len(results))
	for i, cdp := range results {
		responseData[i] = dto.FromModel(cdp)
	}

	// Create paginated response
	response := dto.ListResponse{
		Data:       responseData,
		TotalCount: len(responseData), // Note: This is not the true total count, but count of returned items
		Limit:      limit,
		Offset:     offset,
	}

	return c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/v1/cost-data-points/:id
func (h *CostDataPointHandler) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	// Check for recorded_at query parameter (required for update due to composite key)
	recordedAtStr := c.QueryParam("recorded_at")
	if recordedAtStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "recorded_at query parameter is required for update")
	}

	recordedAt, err := time.Parse(time.RFC3339, recordedAtStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid recorded_at format, use RFC3339")
	}

	// Get existing record
	cdp, err := h.repo.GetByID(c.Request().Context(), id, recordedAt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Cost data point not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cost data point")
	}

	// Parse update request
	var req dto.UpdateCostDataPointRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Additional validation for price if provided
	if req.Price > 0 && req.Price <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Price must be greater than 0")
	}
	if req.Confidence > 0 && (req.Confidence < 0 || req.Confidence > 1) {
		return echo.NewHTTPError(http.StatusBadRequest, "Confidence must be between 0 and 1")
	}

	// Apply updates
	req.ApplyUpdate(cdp)

	// Update in database
	if err := h.repo.Update(c.Request().Context(), cdp); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Cost data point not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update cost data point")
	}

	return c.JSON(http.StatusOK, dto.FromModel(cdp))
}

// Delete handles DELETE /api/v1/cost-data-points/:id
func (h *CostDataPointHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	// Check for recorded_at query parameter (required for delete due to composite key)
	recordedAtStr := c.QueryParam("recorded_at")
	if recordedAtStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "recorded_at query parameter is required for delete")
	}

	recordedAt, err := time.Parse(time.RFC3339, recordedAtStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid recorded_at format, use RFC3339")
	}

	// Delete from database
	if err := h.repo.Delete(c.Request().Context(), id, recordedAt); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Cost data point not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete cost data point")
	}

	return c.NoContent(http.StatusNoContent)
}
