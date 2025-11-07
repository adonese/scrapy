package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if testDB != nil {
		return testDB
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "cost_of_living"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	testDB = db
	return testDB
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec("DELETE FROM cost_data_points WHERE source = 'test'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", err)
	}
}

// createTestCostDataPoint creates a sample cost data point for testing
func createTestCostDataPoint() *models.CostDataPoint {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &models.CostDataPoint{
		Category:    "Housing",
		SubCategory: "Rent",
		ItemName:    "1BR Apartment in Marina",
		Price:       85000.00,
		MinPrice:    80000.00,
		MaxPrice:    90000.00,
		MedianPrice: 85000.00,
		SampleSize:  5,
		Location: models.Location{
			Emirate: "Dubai",
			City:    "Dubai",
			Area:    "Marina",
			Coordinates: &models.GeoPoint{
				Lat: 25.0803,
				Lon: 55.1396,
			},
		},
		RecordedAt: now,
		ValidFrom:  now,
		Source:     "test",
		SourceURL:  "https://example.com/test",
		Confidence: 1.0,
		Unit:       "AED",
		Tags:       []string{"rent", "apartment", "marina"},
		Attributes: map[string]interface{}{
			"bedrooms":  1,
			"bathrooms": 1,
			"furnished": true,
		},
	}
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	repo := NewCostDataPointRepository(db)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		cdp := createTestCostDataPoint()

		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create cost data point: %v", err)
		}

		// Verify ID was generated
		if cdp.ID == "" {
			t.Error("Expected ID to be generated")
		}

		// Verify timestamps were set
		if cdp.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
		if cdp.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}

		// Clean up
		_ = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
	})

	t.Run("create with custom ID", func(t *testing.T) {
		cdp := createTestCostDataPoint()

		// Get a UUID from the database
		var customID string
		err := db.QueryRow("SELECT uuid_generate_v4()").Scan(&customID)
		if err != nil {
			t.Fatalf("Failed to generate UUID: %v", err)
		}
		cdp.ID = customID

		err = repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create cost data point: %v", err)
		}

		if cdp.ID != customID {
			t.Errorf("Expected ID to be %s, got %s", customID, cdp.ID)
		}

		// Clean up
		_ = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
	})

	t.Run("create with minimal fields", func(t *testing.T) {
		cdp := &models.CostDataPoint{
			Category: "Food",
			ItemName: "Bread",
			Price:    5.00,
			Location: models.Location{
				Emirate: "Abu Dhabi",
			},
			Source: "test",
		}

		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create cost data point: %v", err)
		}

		// Verify defaults were set
		if cdp.ID == "" {
			t.Error("Expected ID to be generated")
		}
		if cdp.RecordedAt.IsZero() {
			t.Error("Expected RecordedAt to be set")
		}
		if cdp.ValidFrom.IsZero() {
			t.Error("Expected ValidFrom to be set")
		}
		if cdp.SampleSize != 1 {
			t.Errorf("Expected SampleSize to be 1, got %d", cdp.SampleSize)
		}
		if cdp.Confidence != 1.0 {
			t.Errorf("Expected Confidence to be 1.0, got %f", cdp.Confidence)
		}
		if cdp.Unit != "AED" {
			t.Errorf("Expected Unit to be AED, got %s", cdp.Unit)
		}

		// Clean up
		_ = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
	})
}

func TestGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	repo := NewCostDataPointRepository(db)
	ctx := context.Background()

	t.Run("get existing record", func(t *testing.T) {
		// Create a test record
		cdp := createTestCostDataPoint()
		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}
		defer func() { _ = repo.Delete(ctx, cdp.ID, cdp.RecordedAt) }()

		// Retrieve the record
		retrieved, err := repo.GetByID(ctx, cdp.ID, cdp.RecordedAt)
		if err != nil {
			t.Fatalf("Failed to get cost data point: %v", err)
		}

		// Verify fields
		if retrieved.ID != cdp.ID {
			t.Errorf("Expected ID %s, got %s", cdp.ID, retrieved.ID)
		}
		if retrieved.Category != cdp.Category {
			t.Errorf("Expected Category %s, got %s", cdp.Category, retrieved.Category)
		}
		if retrieved.ItemName != cdp.ItemName {
			t.Errorf("Expected ItemName %s, got %s", cdp.ItemName, retrieved.ItemName)
		}
		if retrieved.Price != cdp.Price {
			t.Errorf("Expected Price %f, got %f", cdp.Price, retrieved.Price)
		}
		if retrieved.Location.Emirate != cdp.Location.Emirate {
			t.Errorf("Expected Emirate %s, got %s", cdp.Location.Emirate, retrieved.Location.Emirate)
		}
		if retrieved.Location.Coordinates.Lat != cdp.Location.Coordinates.Lat {
			t.Errorf("Expected Lat %f, got %f", cdp.Location.Coordinates.Lat, retrieved.Location.Coordinates.Lat)
		}
		if len(retrieved.Tags) != len(cdp.Tags) {
			t.Errorf("Expected %d tags, got %d", len(cdp.Tags), len(retrieved.Tags))
		}
		if retrieved.Attributes["bedrooms"] != float64(1) { // JSON numbers are float64
			t.Errorf("Expected bedrooms attribute to be 1, got %v", retrieved.Attributes["bedrooms"])
		}
	})

	t.Run("get non-existent record", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000", time.Now())
		if err == nil {
			t.Error("Expected error when getting non-existent record")
		}
		if err.Error() != "cost data point not found" {
			t.Errorf("Expected 'cost data point not found' error, got: %v", err)
		}
	})
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	repo := NewCostDataPointRepository(db)
	ctx := context.Background()

	// Create multiple test records
	records := []*models.CostDataPoint{
		createTestCostDataPoint(),
	}
	records[0].Category = "Housing"
	records[0].Location.Emirate = "Dubai"

	cdp2 := createTestCostDataPoint()
	cdp2.Category = "Food"
	cdp2.Location.Emirate = "Abu Dhabi"
	cdp2.RecordedAt = time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond)
	records = append(records, cdp2)

	cdp3 := createTestCostDataPoint()
	cdp3.Category = "Housing"
	cdp3.Location.Emirate = "Sharjah"
	cdp3.RecordedAt = time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	records = append(records, cdp3)

	// Create all test records
	for _, cdp := range records {
		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}
		defer func(id string, recordedAt time.Time) {
			_ = repo.Delete(ctx, id, recordedAt)
		}(cdp.ID, cdp.RecordedAt)
	}

	t.Run("list all", func(t *testing.T) {
		filter := repository.ListFilter{
			Limit: 100,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		if len(results) < 3 {
			t.Errorf("Expected at least 3 results, got %d", len(results))
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		filter := repository.ListFilter{
			Category: "Housing",
			Limit:    100,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		if len(results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(results))
		}

		for _, r := range results {
			if r.Category != "Housing" {
				t.Errorf("Expected category Housing, got %s", r.Category)
			}
		}
	})

	t.Run("filter by emirate", func(t *testing.T) {
		filter := repository.ListFilter{
			Emirate: "Dubai",
			Limit:   100,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		if len(results) < 1 {
			t.Errorf("Expected at least 1 result, got %d", len(results))
		}

		for _, r := range results {
			if r.Location.Emirate != "Dubai" {
				t.Errorf("Expected emirate Dubai, got %s", r.Location.Emirate)
			}
		}
	})

	t.Run("filter by id", func(t *testing.T) {
		target := records[0]
		filter := repository.ListFilter{
			ID:    target.ID,
			Limit: 10,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected results for specific ID filter")
		}

		for _, r := range results {
			if r.ID != target.ID {
				t.Fatalf("Expected only ID %s, got %s", target.ID, r.ID)
			}
		}
	})

	t.Run("filter by date range", func(t *testing.T) {
		now := time.Now().UTC()
		startDate := now.Add(-3 * time.Hour)
		endDate := now.Add(-30 * time.Minute)

		filter := repository.ListFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
			Limit:     100,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		if len(results) < 1 {
			t.Errorf("Expected at least 1 result, got %d", len(results))
		}

		for _, r := range results {
			if r.RecordedAt.Before(startDate) || r.RecordedAt.After(endDate) {
				t.Errorf("Result outside date range: %v", r.RecordedAt)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filter := repository.ListFilter{
			Limit:  2,
			Offset: 0,
		}

		page1, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		filter.Offset = 2
		page2, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		// Verify pages are different
		if len(page1) > 0 && len(page2) > 0 {
			if page1[0].ID == page2[0].ID {
				t.Error("Expected different results on different pages")
			}
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		filter := repository.ListFilter{
			Category: "Housing",
			Emirate:  "Dubai",
			Limit:    100,
		}

		results, err := repo.List(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list cost data points: %v", err)
		}

		for _, r := range results {
			if r.Category != "Housing" {
				t.Errorf("Expected category Housing, got %s", r.Category)
			}
			if r.Location.Emirate != "Dubai" {
				t.Errorf("Expected emirate Dubai, got %s", r.Location.Emirate)
			}
		}
	})
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	repo := NewCostDataPointRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		// Create a test record
		cdp := createTestCostDataPoint()
		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}
		defer func() { _ = repo.Delete(ctx, cdp.ID, cdp.RecordedAt) }()

		// Update the record
		cdp.Price = 90000.00
		cdp.ItemName = "Updated Item Name"
		cdp.Tags = []string{"updated", "tags"}
		cdp.Attributes["updated"] = true

		err = repo.Update(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to update cost data point: %v", err)
		}

		// Retrieve and verify
		retrieved, err := repo.GetByID(ctx, cdp.ID, cdp.RecordedAt)
		if err != nil {
			t.Fatalf("Failed to get updated record: %v", err)
		}

		if retrieved.Price != 90000.00 {
			t.Errorf("Expected Price 90000.00, got %f", retrieved.Price)
		}
		if retrieved.ItemName != "Updated Item Name" {
			t.Errorf("Expected ItemName 'Updated Item Name', got %s", retrieved.ItemName)
		}
		if len(retrieved.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(retrieved.Tags))
		}
		if retrieved.Attributes["updated"] != true {
			t.Errorf("Expected updated attribute to be true")
		}
	})

	t.Run("update non-existent record", func(t *testing.T) {
		cdp := createTestCostDataPoint()
		cdp.ID = "00000000-0000-0000-0000-000000000000"

		err := repo.Update(ctx, cdp)
		if err == nil {
			t.Error("Expected error when updating non-existent record")
		}
		if err.Error() != "cost data point not found" {
			t.Errorf("Expected 'cost data point not found' error, got: %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	repo := NewCostDataPointRepository(db)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		// Create a test record
		cdp := createTestCostDataPoint()
		err := repo.Create(ctx, cdp)
		if err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}

		// Delete the record
		err = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
		if err != nil {
			t.Fatalf("Failed to delete cost data point: %v", err)
		}

		// Verify it's gone
		_, err = repo.GetByID(ctx, cdp.ID, cdp.RecordedAt)
		if err == nil {
			t.Error("Expected error when getting deleted record")
		}
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		err := repo.Delete(ctx, "00000000-0000-0000-0000-000000000000", time.Now())
		if err == nil {
			t.Error("Expected error when deleting non-existent record")
		}
		if err.Error() != "cost data point not found" {
			t.Errorf("Expected 'cost data point not found' error, got: %v", err)
		}
	})
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
