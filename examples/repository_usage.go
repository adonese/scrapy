package examples

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/repository/postgres"
)

// ExampleRepositoryUsage demonstrates how to use the CostDataPointRepository
// This is not a runnable example but serves as documentation
func ExampleRepositoryUsage(db *sql.DB) error {
	// Initialize the repository
	repo := postgres.NewCostDataPointRepository(db)
	ctx := context.Background()

	// 1. Create a new cost data point
	cdp := &models.CostDataPoint{
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
		Source:     "example_source",
		SourceURL:  "https://example.com/listing",
		Confidence: 0.9,
		Unit:       "AED",
		Tags:       []string{"rent", "apartment", "marina"},
		Attributes: map[string]interface{}{
			"bedrooms":  1,
			"bathrooms": 1,
			"furnished": true,
			"amenities": []string{"gym", "pool", "parking"},
		},
	}

	// Create will auto-generate ID and timestamps
	err := repo.Create(ctx, cdp)
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}
	fmt.Printf("Created cost data point with ID: %s\n", cdp.ID)

	// 2. Get by ID and timestamp
	retrieved, err := repo.GetByID(ctx, cdp.ID, cdp.RecordedAt)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	fmt.Printf("Retrieved: %s - %s - AED %.2f\n",
		retrieved.Category, retrieved.ItemName, retrieved.Price)

	// 3. List with filters

	// List all housing items in Dubai
	housingFilter := repository.ListFilter{
		Category: "Housing",
		Emirate:  "Dubai",
		Limit:    10,
		Offset:   0,
	}
	housingList, err := repo.List(ctx, housingFilter)
	if err != nil {
		return fmt.Errorf("failed to list: %w", err)
	}
	fmt.Printf("Found %d housing items in Dubai\n", len(housingList))

	// List recent data (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	now := time.Now()
	recentFilter := repository.ListFilter{
		StartDate: &sevenDaysAgo,
		EndDate:   &now,
		Limit:     50,
	}
	recentList, err := repo.List(ctx, recentFilter)
	if err != nil {
		return fmt.Errorf("failed to list recent: %w", err)
	}
	fmt.Printf("Found %d recent items\n", len(recentList))

	// Pagination example
	page1Filter := repository.ListFilter{
		Category: "Housing",
		Limit:    20,
		Offset:   0,
	}
	page1, err := repo.List(ctx, page1Filter)
	if err != nil {
		return fmt.Errorf("failed to get page 1: %w", err)
	}

	page2Filter := repository.ListFilter{
		Category: "Housing",
		Limit:    20,
		Offset:   20,
	}
	page2, err := repo.List(ctx, page2Filter)
	if err != nil {
		return fmt.Errorf("failed to get page 2: %w", err)
	}
	fmt.Printf("Page 1: %d items, Page 2: %d items\n", len(page1), len(page2))

	// 4. Update a cost data point
	retrieved.Price = 90000.00
	retrieved.MaxPrice = 95000.00
	retrieved.Tags = append(retrieved.Tags, "updated")
	retrieved.Attributes["price_updated"] = true

	err = repo.Update(ctx, retrieved)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	fmt.Printf("Updated price to AED %.2f\n", retrieved.Price)

	// 5. Delete a cost data point
	err = repo.Delete(ctx, cdp.ID, cdp.RecordedAt)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	fmt.Printf("Deleted cost data point %s\n", cdp.ID)

	return nil
}

// ExampleBulkInsert shows how to efficiently insert multiple records
func ExampleBulkInsert(db *sql.DB) error {
	repo := postgres.NewCostDataPointRepository(db)
	ctx := context.Background()

	// Prepare multiple cost data points
	items := []struct {
		category string
		item     string
		price    float64
		emirate  string
	}{
		{"Food", "Milk (1L)", 6.50, "Dubai"},
		{"Food", "Bread", 5.00, "Dubai"},
		{"Food", "Eggs (12)", 12.00, "Dubai"},
		{"Transportation", "Taxi (per km)", 2.50, "Dubai"},
		{"Transportation", "Metro ticket", 5.00, "Dubai"},
	}

	for _, item := range items {
		cdp := &models.CostDataPoint{
			Category: item.category,
			ItemName: item.item,
			Price:    item.price,
			Location: models.Location{
				Emirate: item.emirate,
			},
			Source:     "bulk_import",
			Confidence: 1.0,
			Unit:       "AED",
		}

		if err := repo.Create(ctx, cdp); err != nil {
			return fmt.Errorf("failed to create %s: %w", item.item, err)
		}
	}

	fmt.Printf("Successfully inserted %d items\n", len(items))
	return nil
}

// ExampleComplexQuery shows how to use multiple filters together
func ExampleComplexQuery(db *sql.DB) error {
	repo := postgres.NewCostDataPointRepository(db)
	ctx := context.Background()

	// Find all food items in Dubai from the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	now := time.Now()

	filter := repository.ListFilter{
		Category:  "Food",
		Emirate:   "Dubai",
		StartDate: &thirtyDaysAgo,
		EndDate:   &now,
		Limit:     100,
		Offset:    0,
	}

	results, err := repo.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	// Process results
	if len(results) == 0 {
		fmt.Println("No food items found in Dubai from the last 30 days")
		return nil
	}

	// Calculate average price
	var total float64
	for _, item := range results {
		total += item.Price
	}
	avg := total / float64(len(results))

	fmt.Printf("Found %d food items in Dubai\n", len(results))
	fmt.Printf("Average price: AED %.2f\n", avg)
	fmt.Printf("Price range: AED %.2f - AED %.2f\n",
		results[len(results)-1].Price, results[0].Price)

	return nil
}
