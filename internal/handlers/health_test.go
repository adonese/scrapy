package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// Note: This test uses a mock/nil database since we're testing the handler structure,
// not the actual database connection. For database integration tests, see pkg/database tests.
func TestHealth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create handler with nil DB (will cause "disconnected" status but handler should not crash)
	handler := &HealthHandler{db: nil}

	if err := handler.Health(c); err != nil {
		t.Fatalf("Health handler failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}

	// With nil DB, database status should be "disconnected"
	if response.Database != "disconnected" {
		t.Errorf("expected database 'disconnected' with nil DB, got '%s'", response.Database)
	}
}
