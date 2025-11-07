package mock

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
)

// CostDataPointRepository is a mock implementation of repository.CostDataPointRepository
type CostDataPointRepository struct {
	mu    sync.RWMutex
	data  map[string]*models.CostDataPoint // key is "id:recordedAt"
	calls map[string]int                   // track method calls for testing
}

// NewCostDataPointRepository creates a new mock repository
func NewCostDataPointRepository() *CostDataPointRepository {
	return &CostDataPointRepository{
		data:  make(map[string]*models.CostDataPoint),
		calls: make(map[string]int),
	}
}

// Create implements repository.CostDataPointRepository
func (m *CostDataPointRepository) Create(ctx context.Context, cdp *models.CostDataPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls["Create"]++

	// Generate ID if not provided
	if cdp.ID == "" {
		cdp.ID = fmt.Sprintf("mock-id-%d", len(m.data)+1)
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

	cdp.CreatedAt = time.Now()
	cdp.UpdatedAt = time.Now()

	key := makeKey(cdp.ID, cdp.RecordedAt)
	m.data[key] = cdp

	return nil
}

// GetByID implements repository.CostDataPointRepository
func (m *CostDataPointRepository) GetByID(ctx context.Context, id string, recordedAt time.Time) (*models.CostDataPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.calls["GetByID"]++

	key := makeKey(id, recordedAt)
	cdp, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("cost data point not found")
	}

	return cdp, nil
}

// List implements repository.CostDataPointRepository
func (m *CostDataPointRepository) List(ctx context.Context, filter repository.ListFilter) ([]*models.CostDataPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.calls["List"]++

	var results []*models.CostDataPoint

	// Collect all data points
	for _, cdp := range m.data {
		if filter.ID != "" && cdp.ID != filter.ID {
			continue
		}
		// Apply filters
		if filter.Category != "" && cdp.Category != filter.Category {
			continue
		}
		if filter.Emirate != "" && cdp.Location.Emirate != filter.Emirate {
			continue
		}
		if filter.StartDate != nil && cdp.RecordedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && cdp.RecordedAt.After(*filter.EndDate) {
			continue
		}

		results = append(results, cdp)
	}

	// Sort by recorded_at descending to match real repository behavior
	sort.Slice(results, func(i, j int) bool {
		return results[i].RecordedAt.After(results[j].RecordedAt)
	})

	// Apply pagination
	start := filter.Offset
	if start > len(results) {
		return []*models.CostDataPoint{}, nil
	}

	end := start + filter.Limit
	if filter.Limit > 0 && end < len(results) {
		results = results[start:end]
	} else if start < len(results) {
		results = results[start:]
	}

	return results, nil
}

// Update implements repository.CostDataPointRepository
func (m *CostDataPointRepository) Update(ctx context.Context, cdp *models.CostDataPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls["Update"]++

	key := makeKey(cdp.ID, cdp.RecordedAt)
	if _, exists := m.data[key]; !exists {
		return fmt.Errorf("cost data point not found")
	}

	cdp.UpdatedAt = time.Now()
	m.data[key] = cdp

	return nil
}

// Delete implements repository.CostDataPointRepository
func (m *CostDataPointRepository) Delete(ctx context.Context, id string, recordedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls["Delete"]++

	key := makeKey(id, recordedAt)
	if _, exists := m.data[key]; !exists {
		return fmt.Errorf("cost data point not found")
	}

	delete(m.data, key)
	return nil
}

// GetCallCount returns the number of times a method was called
func (m *CostDataPointRepository) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.calls[method]
}

// Reset clears all data and call counts
func (m *CostDataPointRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]*models.CostDataPoint)
	m.calls = make(map[string]int)
}

// makeKey creates a composite key from id and recordedAt
func makeKey(id string, recordedAt time.Time) string {
	return fmt.Sprintf("%s:%d", id, recordedAt.Unix())
}
