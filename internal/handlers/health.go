package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/labstack/echo/v4"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

type HealthHandler struct {
	db *database.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health returns the health status of the service
func (h *HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.db.HealthCheck(ctx); err != nil {
		dbStatus = "disconnected"
	}

	response := HealthResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response)
}
