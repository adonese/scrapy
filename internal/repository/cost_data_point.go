package repository

import (
	"context"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// CostDataPointRepository defines the interface for cost data point operations
type CostDataPointRepository interface {
	// Create inserts a new cost data point into the database
	Create(ctx context.Context, cdp *models.CostDataPoint) error

	// GetByID retrieves a cost data point by ID and recorded_at timestamp
	// Since the table uses composite primary key (id, recorded_at)
	GetByID(ctx context.Context, id string, recordedAt time.Time) (*models.CostDataPoint, error)

	// List retrieves cost data points based on the provided filter
	List(ctx context.Context, filter ListFilter) ([]*models.CostDataPoint, error)

	// Update updates an existing cost data point
	Update(ctx context.Context, cdp *models.CostDataPoint) error

	// Delete removes a cost data point by ID and recorded_at timestamp
	Delete(ctx context.Context, id string, recordedAt time.Time) error
}

// ListFilter defines filtering options for listing cost data points
type ListFilter struct {
	// ID filters by specific cost data point ID
	ID string

	// Category filters by category (exact match)
	Category string

	// SubCategory filters by sub category (exact match)
	SubCategory string

	// Emirate filters by location emirate (exact match)
	Emirate string

	// StartDate filters records where recorded_at >= StartDate
	StartDate *time.Time

	// EndDate filters records where recorded_at <= EndDate
	EndDate *time.Time

	// Limit specifies the maximum number of records to return
	Limit int

	// Offset specifies the number of records to skip
	Offset int
}
