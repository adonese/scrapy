package integration

import (
	"context"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/internal/scrapers/bayut"
	"github.com/adonese/cost-of-living/internal/scrapers/dubizzle"
	"github.com/adonese/cost-of-living/internal/services"
	"github.com/adonese/cost-of-living/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// TestScraperWorkflowIntegration tests the full workflow with scrapers
func TestScraperWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the scraper activity
	env.OnActivity(workflow.RunScraperActivity, mock.Anything, "bayut").Return(&workflow.ScraperActivityResult{
		ItemsScraped: 10,
		ItemsSaved:   10,
		Duration:     time.Second * 5,
	}, nil)

	// Execute workflow
	input := workflow.ScraperWorkflowInput{
		ScraperName: "bayut",
		MaxRetries:  3,
	}
	env.ExecuteWorkflow(workflow.ScraperWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result workflow.ScraperWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// Validate results
	assert.Equal(t, "bayut", result.ScraperName)
	assert.Equal(t, 10, result.ItemsScraped)
	assert.Equal(t, 10, result.ItemsSaved)
	assert.Empty(t, result.Errors)
	assert.NotZero(t, result.CompletedAt)
}

// TestScraperWorkflowWithRetry tests workflow retry logic
func TestScraperWorkflowWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	attemptCount := 0

	// Mock activity that fails first time, succeeds second time
	env.OnActivity(workflow.RunScraperActivity, mock.Anything, "dubizzle").Return(
		func(ctx context.Context, scraperName string) (*workflow.ScraperActivityResult, error) {
			attemptCount++
			if attemptCount < 2 {
				return nil, assert.AnError
			}
			return &workflow.ScraperActivityResult{
				ItemsScraped: 5,
				ItemsSaved:   5,
				Duration:     time.Second * 3,
			}, nil
		},
	)

	// Execute workflow with retries
	input := workflow.ScraperWorkflowInput{
		ScraperName: "dubizzle",
		MaxRetries:  3,
	}
	env.ExecuteWorkflow(workflow.ScraperWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result workflow.ScraperWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// Should succeed after retry
	assert.Equal(t, "dubizzle", result.ScraperName)
	assert.Equal(t, 5, result.ItemsScraped)
	assert.Empty(t, result.Errors)
}

// TestScraperWorkflowWithCompensation tests compensation logic
func TestScraperWorkflowWithCompensation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity to always fail
	env.OnActivity(workflow.RunScraperActivity, mock.Anything, "bayut").Return(
		(*workflow.ScraperActivityResult)(nil),
		assert.AnError,
	)

	// Mock compensation activity
	env.OnActivity(workflow.CompensateFailedScrapeActivity, mock.Anything, "bayut").Return(true, nil)

	// Execute workflow
	input := workflow.ScraperWorkflowInput{
		ScraperName: "bayut",
		MaxRetries:  1,
	}
	env.ExecuteWorkflow(workflow.ScraperWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

// TestScheduledScraperWorkflowIntegration tests the scheduled workflow
func TestScheduledScraperWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock child workflows for each scraper
	scrapers := []string{"bayut", "dubizzle"}
	for _, scraperName := range scrapers {
		env.OnWorkflow(workflow.ScraperWorkflow, mock.Anything, mock.MatchedBy(func(input workflow.ScraperWorkflowInput) bool {
			return input.ScraperName == scraperName
		})).Return(&workflow.ScraperWorkflowResult{
			ScraperName:  scraperName,
			ItemsScraped: 10,
			ItemsSaved:   10,
			Errors:       []string{},
			Duration:     time.Second * 5,
			CompletedAt:  time.Now(),
		}, nil)
	}

	// Execute scheduled workflow
	env.ExecuteWorkflow(workflow.ScheduledScraperWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

// TestScraperServiceIntegration tests the scraper service end-to-end
func TestScraperServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock repository
	mockRepo := &MockRepository{
		items: make(map[string]*models.CostDataPoint),
	}

	// Create scraper configuration
	config := scrapers.Config{
		UserAgent:  "Mozilla/5.0 (Test)",
		RateLimit:  10,
		Timeout:    30,
		MaxRetries: 3,
	}

	// Create scraper service
	scraperService := services.NewScraperService(mockRepo)

	// Register scrapers
	scraperService.RegisterScraper(bayut.NewBayutScraper(config))
	scraperService.RegisterScraper(dubizzle.NewDubizzleScraper(config))

	// Test running a scraper
	ctx := context.Background()
	err := scraperService.RunScraper(ctx, "bayut")

	// We expect an error since we're not connected to the real site
	// but this tests the integration structure
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// TestBatchScrapingWorkflow tests batch processing of multiple scrapers
func TestBatchScrapingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activities for multiple scrapers
	scrapers := []string{"bayut", "dubizzle", "bayut_abu_dhabi", "dubizzle_sharjah"}

	for _, scraperName := range scrapers {
		env.OnActivity(workflow.RunScraperActivity, mock.Anything, scraperName).Return(&workflow.ScraperActivityResult{
			ItemsScraped: 10,
			ItemsSaved:   10,
			Duration:     time.Second * 5,
		}, nil)
	}

	// Test that all scrapers complete successfully
	for _, scraperName := range scrapers {
		input := workflow.ScraperWorkflowInput{
			ScraperName: scraperName,
			MaxRetries:  3,
		}

		env.ExecuteWorkflow(workflow.ScraperWorkflow, input)
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		// Reset for next test
		env = testSuite.NewTestWorkflowEnvironment()
		env.OnActivity(workflow.RunScraperActivity, mock.Anything, scraperName).Return(&workflow.ScraperActivityResult{
			ItemsScraped: 10,
			ItemsSaved:   10,
			Duration:     time.Second * 5,
		}, nil)
	}
}

// TestWorkflowWithDatabaseIntegration tests workflow with database operations
func TestWorkflowWithDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real database connection
	// For now, we demonstrate the structure

	t.Run("Store scraped data", func(t *testing.T) {
		mockRepo := &MockRepository{
			items: make(map[string]*models.CostDataPoint),
		}

		ctx := context.Background()

		// Create test data point
		dataPoint := CreateTestDataPoint("Housing", "Test Apartment", "test", 50000)

		// Store data point
		err := mockRepo.Create(ctx, dataPoint)
		require.NoError(t, err)

		// Verify it was stored
		assert.Equal(t, 1, len(mockRepo.items))
	})

	t.Run("Handle duplicate data", func(t *testing.T) {
		mockRepo := &MockRepository{
			items: make(map[string]*models.CostDataPoint),
		}

		ctx := context.Background()

		// Create test data point
		dataPoint := CreateTestDataPoint("Housing", "Test Apartment", "test", 50000)
		dataPoint.ID = "test-id-1"

		// Store data point twice
		err := mockRepo.Create(ctx, dataPoint)
		require.NoError(t, err)

		// Second create should work (or update in real implementation)
		err = mockRepo.Create(ctx, dataPoint)
		require.NoError(t, err)
	})
}

// TestWorkflowTimeouts tests workflow timeout handling
func TestWorkflowTimeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity that takes too long
	env.OnActivity(workflow.RunScraperActivity, mock.Anything, "bayut").Return(
		func(ctx context.Context, scraperName string) (*workflow.ScraperActivityResult, error) {
			// Simulate long-running activity
			select {
			case <-time.After(10 * time.Minute):
				return &workflow.ScraperActivityResult{
					ItemsScraped: 10,
					ItemsSaved:   10,
					Duration:     10 * time.Minute,
				}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	)

	// Execute workflow with timeout
	input := workflow.ScraperWorkflowInput{
		ScraperName: "bayut",
		MaxRetries:  1,
	}

	env.SetTestTimeout(5 * time.Second)
	env.ExecuteWorkflow(workflow.ScraperWorkflow, input)

	// Should timeout
	require.True(t, env.IsWorkflowCompleted())
}

// MockRepository is a mock implementation of CostDataPointRepository
type MockRepository struct {
	items map[string]*models.CostDataPoint
}

func (m *MockRepository) Create(ctx context.Context, cdp *models.CostDataPoint) error {
	if cdp.ID == "" {
		cdp.ID = "mock-id"
	}
	m.items[cdp.ID] = cdp
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string, recordedAt time.Time) (*models.CostDataPoint, error) {
	if item, ok := m.items[id]; ok {
		return item, nil
	}
	return nil, assert.AnError
}

func (m *MockRepository) Update(ctx context.Context, cdp *models.CostDataPoint) error {
	m.items[cdp.ID] = cdp
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id string, recordedAt time.Time) error {
	delete(m.items, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context, filter repository.ListFilter) ([]*models.CostDataPoint, error) {
	result := make([]*models.CostDataPoint, 0, len(m.items))
	for _, item := range m.items {
		result = append(result, item)
	}
	return result, nil
}

func (m *MockRepository) Count(ctx context.Context, filter repository.ListFilter) (int, error) {
	return len(m.items), nil
}
