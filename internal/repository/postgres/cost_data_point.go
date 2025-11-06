package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/lib/pq"
)

// CostDataPointRepository implements the repository.CostDataPointRepository interface
type CostDataPointRepository struct {
	db *sql.DB
}

// NewCostDataPointRepository creates a new instance of CostDataPointRepository
func NewCostDataPointRepository(db *sql.DB) *CostDataPointRepository {
	return &CostDataPointRepository{db: db}
}

// Create inserts a new cost data point into the database
func (r *CostDataPointRepository) Create(ctx context.Context, cdp *models.CostDataPoint) error {
	// Generate UUID if not provided
	if cdp.ID == "" {
		var id string
		err := r.db.QueryRowContext(ctx, "SELECT uuid_generate_v4()").Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to generate UUID: %w", err)
		}
		cdp.ID = id
	}

	// Set default values if not provided
	if cdp.RecordedAt.IsZero() {
		cdp.RecordedAt = time.Now()
	}
	if cdp.ValidFrom.IsZero() {
		cdp.ValidFrom = time.Now()
	}
	if cdp.SampleSize == 0 {
		cdp.SampleSize = 1
	}
	if cdp.Confidence == 0 {
		cdp.Confidence = 1.0
	}
	if cdp.Unit == "" {
		cdp.Unit = "AED"
	}

	// Marshal location to JSON
	locationJSON, err := json.Marshal(cdp.Location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	// Marshal attributes to JSON
	attributesJSON, err := json.Marshal(cdp.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		INSERT INTO cost_data_points (
			id, category, sub_category, item_name, price, min_price, max_price,
			median_price, sample_size, location, recorded_at, valid_from, valid_to,
			source, source_url, confidence, unit, tags, attributes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		cdp.ID,
		cdp.Category,
		nullString(cdp.SubCategory),
		cdp.ItemName,
		cdp.Price,
		nullFloat64(cdp.MinPrice),
		nullFloat64(cdp.MaxPrice),
		nullFloat64(cdp.MedianPrice),
		cdp.SampleSize,
		locationJSON,
		cdp.RecordedAt,
		cdp.ValidFrom,
		nullTime(cdp.ValidTo),
		cdp.Source,
		nullString(cdp.SourceURL),
		cdp.Confidence,
		cdp.Unit,
		pq.Array(cdp.Tags),
		attributesJSON,
	).Scan(&cdp.CreatedAt, &cdp.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create cost data point: %w", err)
	}

	return nil
}

// GetByID retrieves a cost data point by ID and recorded_at timestamp
func (r *CostDataPointRepository) GetByID(ctx context.Context, id string, recordedAt time.Time) (*models.CostDataPoint, error) {
	query := `
		SELECT
			id, category, sub_category, item_name, price, min_price, max_price,
			median_price, sample_size, location, recorded_at, valid_from, valid_to,
			source, source_url, confidence, unit, tags, attributes, created_at, updated_at
		FROM cost_data_points
		WHERE id = $1 AND recorded_at = $2
	`

	cdp := &models.CostDataPoint{}
	var locationJSON []byte
	var attributesJSON []byte
	var subCategory, sourceURL sql.NullString
	var minPrice, maxPrice, medianPrice sql.NullFloat64
	var validTo sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, recordedAt).Scan(
		&cdp.ID,
		&cdp.Category,
		&subCategory,
		&cdp.ItemName,
		&cdp.Price,
		&minPrice,
		&maxPrice,
		&medianPrice,
		&cdp.SampleSize,
		&locationJSON,
		&cdp.RecordedAt,
		&cdp.ValidFrom,
		&validTo,
		&cdp.Source,
		&sourceURL,
		&cdp.Confidence,
		&cdp.Unit,
		pq.Array(&cdp.Tags),
		&attributesJSON,
		&cdp.CreatedAt,
		&cdp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cost data point not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cost data point: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(locationJSON, &cdp.Location); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %w", err)
	}

	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &cdp.Attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
		}
	}

	// Handle nullable fields
	if subCategory.Valid {
		cdp.SubCategory = subCategory.String
	}
	if sourceURL.Valid {
		cdp.SourceURL = sourceURL.String
	}
	if minPrice.Valid {
		cdp.MinPrice = minPrice.Float64
	}
	if maxPrice.Valid {
		cdp.MaxPrice = maxPrice.Float64
	}
	if medianPrice.Valid {
		cdp.MedianPrice = medianPrice.Float64
	}
	if validTo.Valid {
		cdp.ValidTo = &validTo.Time
	}

	return cdp, nil
}

// List retrieves cost data points based on the provided filter
func (r *CostDataPointRepository) List(ctx context.Context, filter repository.ListFilter) ([]*models.CostDataPoint, error) {
	query := `
		SELECT
			id, category, sub_category, item_name, price, min_price, max_price,
			median_price, sample_size, location, recorded_at, valid_from, valid_to,
			source, source_url, confidence, unit, tags, attributes, created_at, updated_at
		FROM cost_data_points
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, filter.Category)
		argPos++
	}

	if filter.Emirate != "" {
		query += fmt.Sprintf(" AND location->>'emirate' = $%d", argPos)
		args = append(args, filter.Emirate)
		argPos++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argPos)
		args = append(args, *filter.StartDate)
		argPos++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argPos)
		args = append(args, *filter.EndDate)
		argPos++
	}

	// Order by recorded_at DESC (most recent first)
	query += " ORDER BY recorded_at DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cost data points: %w", err)
	}
	defer rows.Close()

	var results []*models.CostDataPoint

	for rows.Next() {
		cdp := &models.CostDataPoint{}
		var locationJSON []byte
		var attributesJSON []byte
		var subCategory, sourceURL sql.NullString
		var minPrice, maxPrice, medianPrice sql.NullFloat64
		var validTo sql.NullTime

		err := rows.Scan(
			&cdp.ID,
			&cdp.Category,
			&subCategory,
			&cdp.ItemName,
			&cdp.Price,
			&minPrice,
			&maxPrice,
			&medianPrice,
			&cdp.SampleSize,
			&locationJSON,
			&cdp.RecordedAt,
			&cdp.ValidFrom,
			&validTo,
			&cdp.Source,
			&sourceURL,
			&cdp.Confidence,
			&cdp.Unit,
			pq.Array(&cdp.Tags),
			&attributesJSON,
			&cdp.CreatedAt,
			&cdp.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(locationJSON, &cdp.Location); err != nil {
			return nil, fmt.Errorf("failed to unmarshal location: %w", err)
		}

		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &cdp.Attributes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
			}
		}

		// Handle nullable fields
		if subCategory.Valid {
			cdp.SubCategory = subCategory.String
		}
		if sourceURL.Valid {
			cdp.SourceURL = sourceURL.String
		}
		if minPrice.Valid {
			cdp.MinPrice = minPrice.Float64
		}
		if maxPrice.Valid {
			cdp.MaxPrice = maxPrice.Float64
		}
		if medianPrice.Valid {
			cdp.MedianPrice = medianPrice.Float64
		}
		if validTo.Valid {
			cdp.ValidTo = &validTo.Time
		}

		results = append(results, cdp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Update updates an existing cost data point
func (r *CostDataPointRepository) Update(ctx context.Context, cdp *models.CostDataPoint) error {
	// Marshal location to JSON
	locationJSON, err := json.Marshal(cdp.Location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	// Marshal attributes to JSON
	attributesJSON, err := json.Marshal(cdp.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		UPDATE cost_data_points SET
			category = $1,
			sub_category = $2,
			item_name = $3,
			price = $4,
			min_price = $5,
			max_price = $6,
			median_price = $7,
			sample_size = $8,
			location = $9,
			valid_from = $10,
			valid_to = $11,
			source = $12,
			source_url = $13,
			confidence = $14,
			unit = $15,
			tags = $16,
			attributes = $17
		WHERE id = $18 AND recorded_at = $19
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		cdp.Category,
		nullString(cdp.SubCategory),
		cdp.ItemName,
		cdp.Price,
		nullFloat64(cdp.MinPrice),
		nullFloat64(cdp.MaxPrice),
		nullFloat64(cdp.MedianPrice),
		cdp.SampleSize,
		locationJSON,
		cdp.ValidFrom,
		nullTime(cdp.ValidTo),
		cdp.Source,
		nullString(cdp.SourceURL),
		cdp.Confidence,
		cdp.Unit,
		pq.Array(cdp.Tags),
		attributesJSON,
		cdp.ID,
		cdp.RecordedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update cost data point: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cost data point not found")
	}

	return nil
}

// Delete removes a cost data point by ID and recorded_at timestamp
func (r *CostDataPointRepository) Delete(ctx context.Context, id string, recordedAt time.Time) error {
	query := `DELETE FROM cost_data_points WHERE id = $1 AND recorded_at = $2`

	result, err := r.db.ExecContext(ctx, query, id, recordedAt)
	if err != nil {
		return fmt.Errorf("failed to delete cost data point: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cost data point not found")
	}

	return nil
}

// Helper functions to handle nullable fields

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
