package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/stretchr/testify/require"
)

// TestDB creates a test database connection
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Use test database configuration
	config := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "cost_of_living_test",
		SSLMode:  "disable",
	}

	db, err := database.Connect(config)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		if db != nil {
			db.Close()
		}
	}

	return db, cleanup
}

// setupTestRepository creates a test repository
func setupTestRepository(t *testing.T, db *database.DB) repository.CostDataPointRepository {
	t.Helper()
	return postgres.NewCostDataPointRepository(db.GetConn())
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, repo repository.CostDataPointRepository) {
	t.Helper()
	ctx := context.Background()

	// Delete all test data
	// Note: In production, you'd want a more targeted cleanup
	// For now, we'll rely on the repository's methods
	filter := repository.ListFilter{
		Limit:  1000,
		Offset: 0,
	}
	items, err := repo.List(ctx, filter)
	if err != nil {
		t.Logf("Warning: failed to list items for cleanup: %v", err)
		return
	}

	for _, item := range items {
		if err := repo.Delete(ctx, item.ID, item.RecordedAt); err != nil {
			t.Logf("Warning: failed to delete item %s: %v", item.ID, err)
		}
	}
}

// MockHTTPServer creates a mock HTTP server for testing scrapers
type MockHTTPServer struct {
	Server *httptest.Server
	URLs   map[string]string
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer(handler http.HandlerFunc) *MockHTTPServer {
	server := httptest.NewServer(handler)
	return &MockHTTPServer{
		Server: server,
		URLs:   make(map[string]string),
	}
}

// Close closes the mock server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// URL returns the base URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}

// ValidateCostDataPoint validates that a CostDataPoint has all required fields
func ValidateCostDataPoint(t *testing.T, cdp *models.CostDataPoint) {
	t.Helper()

	require.NotNil(t, cdp, "CostDataPoint should not be nil")
	require.NotEmpty(t, cdp.Category, "Category should not be empty")
	require.NotEmpty(t, cdp.ItemName, "ItemName should not be empty")
	require.Greater(t, cdp.Price, 0.0, "Price should be greater than 0")
	require.NotEmpty(t, cdp.Location.Emirate, "Emirate should not be empty")
	require.NotEmpty(t, cdp.Source, "Source should not be empty")
	require.Greater(t, cdp.Confidence, float32(0), "Confidence should be greater than 0")
	require.NotEmpty(t, cdp.Unit, "Unit should not be empty")
	require.NotZero(t, cdp.RecordedAt, "RecordedAt should not be zero")
	require.NotZero(t, cdp.ValidFrom, "ValidFrom should not be zero")
}

// ValidateDataPoints validates a slice of CostDataPoints
func ValidateDataPoints(t *testing.T, dataPoints []*models.CostDataPoint, minCount int) {
	t.Helper()

	require.NotNil(t, dataPoints, "DataPoints should not be nil")
	require.GreaterOrEqual(t, len(dataPoints), minCount, "Should have at least %d data points", minCount)

	for i, cdp := range dataPoints {
		t.Run("DataPoint_"+string(rune(i)), func(t *testing.T) {
			ValidateCostDataPoint(t, cdp)
		})
	}
}

// WaitForCondition waits for a condition to be true with a timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				t.Fatalf("Timeout waiting for condition: %s", message)
			}
		}
	}
}

// AssertDataPointsEqual asserts that two data points are equal
func AssertDataPointsEqual(t *testing.T, expected, actual *models.CostDataPoint) {
	t.Helper()

	require.Equal(t, expected.Category, actual.Category, "Category mismatch")
	require.Equal(t, expected.SubCategory, actual.SubCategory, "SubCategory mismatch")
	require.Equal(t, expected.ItemName, actual.ItemName, "ItemName mismatch")
	require.Equal(t, expected.Price, actual.Price, "Price mismatch")
	require.Equal(t, expected.Location.Emirate, actual.Location.Emirate, "Emirate mismatch")
	require.Equal(t, expected.Location.City, actual.Location.City, "City mismatch")
	require.Equal(t, expected.Location.Area, actual.Location.Area, "Area mismatch")
	require.Equal(t, expected.Source, actual.Source, "Source mismatch")
	require.Equal(t, expected.Unit, actual.Unit, "Unit mismatch")
}

// CreateTestDataPoint creates a test CostDataPoint
func CreateTestDataPoint(category, itemName, source string, price float64) *models.CostDataPoint {
	now := time.Now()
	return &models.CostDataPoint{
		Category:    category,
		SubCategory: "Test",
		ItemName:    itemName,
		Price:       price,
		Location: models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
			Area:    "Test Area",
		},
		Source:      source,
		SourceURL:   "https://example.com/test",
		Confidence:  0.8,
		Unit:        "AED",
		RecordedAt:  now,
		ValidFrom:   now,
		SampleSize:  1,
		Tags:        []string{"test"},
		Attributes:  map[string]interface{}{},
	}
}
