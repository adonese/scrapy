package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Health returns the health status of the service
func Health(c echo.Context) error {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	return c.JSON(http.StatusOK, response)
}
