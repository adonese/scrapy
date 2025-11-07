package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
	"github.com/adonese/cost-of-living/internal/repository/mock"
	"github.com/adonese/cost-of-living/pkg/logger"
)

type stubScraper struct {
	name      string
	points    []*models.CostDataPoint
	scrapeErr error
	canScrape bool
}

func (s *stubScraper) Name() string {
	return s.name
}

func (s *stubScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	if s.scrapeErr != nil {
		return nil, s.scrapeErr
	}
	return s.points, nil
}

func (s *stubScraper) CanScrape() bool {
	if !s.canScrape {
		return false
	}
	return true
}

func TestScraperServiceRunScraperSuccess(t *testing.T) {
	logger.Init()

	repo := mock.NewCostDataPointRepository()
	config := &ScraperServiceConfig{
		EnableValidation:   false,
		MinQualityScore:    0.7,
		FailOnValidation:   false,
		ValidateBeforeSave: false,
	}

	service := NewScraperServiceWithConfig(repo, config)

	points := []*models.CostDataPoint{
		newTestPoint("item-1"),
		newTestPoint("item-2"),
	}
	scraper := &stubScraper{name: "test", points: points, canScrape: true}
	service.RegisterScraper(scraper)

	result, err := service.RunScraper(context.Background(), "test")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test", result.ScraperName)
	assert.Equal(t, len(points), result.Fetched)
	assert.Equal(t, len(points), result.Saved)
	assert.Equal(t, 0, result.SaveFailures)
	assert.True(t, result.Validation.Skipped)
	assert.Empty(t, result.Errors)
	assert.Greater(t, result.Duration, time.Duration(0))

	// Repository should contain the saved points
	savedPoints, err := repo.List(context.Background(), mockListFilter())
	require.NoError(t, err)
	assert.Len(t, savedPoints, len(points))
}

func TestScraperServiceRunScraperHandlesSaveFailures(t *testing.T) {
	logger.Init()

	mockRepo := mock.NewCostDataPointRepository()
	failingRepo := &failingRepository{
		CostDataPointRepository: mockRepo,
		failAfter:               1,
		err:                     errors.New("forced failure"),
	}

	config := &ScraperServiceConfig{EnableValidation: false, ValidateBeforeSave: false}
	service := NewScraperServiceWithConfig(failingRepo, config)

	points := []*models.CostDataPoint{
		newTestPoint("item-1"),
		newTestPoint("item-2"),
	}
	scraper := &stubScraper{name: "test", points: points, canScrape: true}
	service.RegisterScraper(scraper)

	result, err := service.RunScraper(context.Background(), "test")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.SaveFailures)
	assert.Equal(t, 1, result.Saved)
	assert.Len(t, result.Errors, 1)
}

func TestScraperServiceRunAllScrapersAggregatesErrors(t *testing.T) {
	logger.Init()

	repo := mock.NewCostDataPointRepository()
	config := &ScraperServiceConfig{EnableValidation: false, ValidateBeforeSave: false}
	service := NewScraperServiceWithConfig(repo, config)

	goodScraper := &stubScraper{name: "good", points: []*models.CostDataPoint{newTestPoint("good")}, canScrape: true}
	badScraper := &stubScraper{name: "bad", scrapeErr: errors.New("boom"), canScrape: true}

	service.RegisterScraper(goodScraper)
	service.RegisterScraper(badScraper)

	results, err := service.RunAllScrapers(context.Background())
	require.Error(t, err)
	require.Len(t, results, 2)

	good := findResult(results, "good")
	require.NotNil(t, good)
	assert.Equal(t, 1, good.Saved)

	bad := findResult(results, "bad")
	require.NotNil(t, bad)
	assert.NotEmpty(t, bad.Errors)
}

func newTestPoint(name string) *models.CostDataPoint {
	now := time.Now()
	return &models.CostDataPoint{
		Category:   "Test",
		ItemName:   name,
		Price:      100,
		SampleSize: 1,
		RecordedAt: now,
		ValidFrom:  now,
		Source:     "test",
		Confidence: 1,
		Unit:       "AED",
	}
}

func mockListFilter() repository.ListFilter {
	return repository.ListFilter{Limit: 10, Offset: 0}
}

type failingRepository struct {
	repository.CostDataPointRepository
	failAfter int
	calls     int
	err       error
}

func (r *failingRepository) Create(ctx context.Context, cdp *models.CostDataPoint) error {
	r.calls++
	if r.failAfter > 0 && r.calls > r.failAfter {
		return r.err
	}
	return r.CostDataPointRepository.Create(ctx, cdp)
}

func findResult(results []*ScrapeResult, name string) *ScrapeResult {
	for _, r := range results {
		if r.ScraperName == name {
			return r
		}
	}
	return nil
}
