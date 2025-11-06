package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/handlers/dto"
	"github.com/adonese/cost-of-living/internal/repository/mock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCostDataPointHandler_Create(t *testing.T) {
	e := echo.New()
	mockRepo := mock.NewCostDataPointRepository()
	handler := NewCostDataPointHandler(mockRepo)

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.Reset()

		reqBody := `{
			"category": "Housing",
			"item_name": "1BR Apartment Marina",
			"price": 85000,
			"location": {
				"emirate": "Dubai",
				"city": "Dubai",
				"area": "Marina"
			},
			"source": "manual"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cost-data-points", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response dto.CostDataPointResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "Housing", response.Category)
		assert.Equal(t, "1BR Apartment Marina", response.ItemName)
		assert.Equal(t, 85000.0, response.Price)
		assert.Equal(t, "Dubai", response.Location.Emirate)
		assert.Equal(t, "manual", response.Source)
		assert.Equal(t, 1, mockRepo.GetCallCount("Create"))
	})

	t.Run("validation error - missing required fields", func(t *testing.T) {
		mockRepo.Reset()

		reqBody := `{
			"category": "Housing",
			"price": 85000
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cost-data-points", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("validation error - invalid price", func(t *testing.T) {
		mockRepo.Reset()

		reqBody := `{
			"category": "Housing",
			"item_name": "1BR Apartment",
			"price": 0,
			"location": {
				"emirate": "Dubai"
			},
			"source": "manual"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cost-data-points", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo.Reset()

		reqBody := `{invalid json`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cost-data-points", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestCostDataPointHandler_GetByID(t *testing.T) {
	e := echo.New()
	mockRepo := mock.NewCostDataPointRepository()
	handler := NewCostDataPointHandler(mockRepo)

	t.Run("successful get by ID with recorded_at", func(t *testing.T) {
		mockRepo.Reset()

		// Create a test record
		now := time.Now()
		createReq := dto.CreateCostDataPointRequest{
			Category: "Housing",
			ItemName: "Test Apartment",
			Price:    50000,
			Location: dto.LocationDTO{
				Emirate: "Dubai",
			},
			Source:     "manual",
			RecordedAt: &now,
		}
		cdp := createReq.ToModel()
		cdp.ID = "test-id-1"
		err := mockRepo.Create(nil, cdp)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points/test-id-1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("test-id-1")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err = handler.GetByID(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.CostDataPointResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "test-id-1", response.ID)
		assert.Equal(t, "Housing", response.Category)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.Reset()

		now := time.Now()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err := handler.GetByID(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("missing ID", func(t *testing.T) {
		mockRepo.Reset()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.GetByID(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestCostDataPointHandler_List(t *testing.T) {
	e := echo.New()
	mockRepo := mock.NewCostDataPointRepository()
	handler := NewCostDataPointHandler(mockRepo)

	t.Run("successful list", func(t *testing.T) {
		mockRepo.Reset()

		// Create test records
		for i := 1; i <= 5; i++ {
			createReq := dto.CreateCostDataPointRequest{
				Category: "Housing",
				ItemName: "Test Apartment",
				Price:    50000.0,
				Location: dto.LocationDTO{
					Emirate: "Dubai",
				},
				Source: "manual",
			}
			cdp := createReq.ToModel()
			err := mockRepo.Create(nil, cdp)
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.List(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.ListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 5, len(response.Data))
		assert.Equal(t, 10, response.Limit)
		assert.Equal(t, 0, response.Offset)
	})

	t.Run("list with filters", func(t *testing.T) {
		mockRepo.Reset()

		// Create test records with different categories
		housingReq := dto.CreateCostDataPointRequest{
			Category: "Housing",
			ItemName: "Apartment",
			Price:    50000,
			Location: dto.LocationDTO{Emirate: "Dubai"},
			Source:   "manual",
		}
		err := mockRepo.Create(nil, housingReq.ToModel())
		require.NoError(t, err)

		foodReq := dto.CreateCostDataPointRequest{
			Category: "Food",
			ItemName: "Meal",
			Price:    50,
			Location: dto.LocationDTO{Emirate: "Abu Dhabi"},
			Source:   "manual",
		}
		err = mockRepo.Create(nil, foodReq.ToModel())
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points?category=Housing", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.List(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.ListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 1, len(response.Data))
		assert.Equal(t, "Housing", response.Data[0].Category)
	})

	t.Run("list with pagination", func(t *testing.T) {
		mockRepo.Reset()

		// Create test records
		for i := 1; i <= 15; i++ {
			createReq := dto.CreateCostDataPointRequest{
				Category: "Housing",
				ItemName: "Test Apartment",
				Price:    50000.0,
				Location: dto.LocationDTO{Emirate: "Dubai"},
				Source:   "manual",
			}
			err := mockRepo.Create(nil, createReq.ToModel())
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/cost-data-points?limit=5&offset=0", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.List(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.ListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 5, len(response.Data))
		assert.Equal(t, 5, response.Limit)
		assert.Equal(t, 0, response.Offset)
	})
}

func TestCostDataPointHandler_Update(t *testing.T) {
	e := echo.New()
	mockRepo := mock.NewCostDataPointRepository()
	handler := NewCostDataPointHandler(mockRepo)

	t.Run("successful update", func(t *testing.T) {
		mockRepo.Reset()

		// Create a test record
		now := time.Now()
		createReq := dto.CreateCostDataPointRequest{
			Category: "Housing",
			ItemName: "Test Apartment",
			Price:    50000,
			Location: dto.LocationDTO{Emirate: "Dubai"},
			Source:   "manual",
		}
		cdp := createReq.ToModel()
		cdp.ID = "test-id-1"
		cdp.RecordedAt = now
		err := mockRepo.Create(nil, cdp)
		require.NoError(t, err)

		updateBody := `{"price": 60000}`

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cost-data-points/test-id-1", strings.NewReader(updateBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("test-id-1")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err = handler.Update(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response dto.CostDataPointResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "test-id-1", response.ID)
		assert.Equal(t, 60000.0, response.Price)
	})

	t.Run("update not found", func(t *testing.T) {
		mockRepo.Reset()

		now := time.Now()
		updateBody := `{"price": 60000}`

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cost-data-points/nonexistent", strings.NewReader(updateBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err := handler.Update(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("missing recorded_at", func(t *testing.T) {
		mockRepo.Reset()

		updateBody := `{"price": 60000}`

		req := httptest.NewRequest(http.MethodPut, "/api/v1/cost-data-points/test-id-1", strings.NewReader(updateBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("test-id-1")

		err := handler.Update(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestCostDataPointHandler_Delete(t *testing.T) {
	e := echo.New()
	mockRepo := mock.NewCostDataPointRepository()
	handler := NewCostDataPointHandler(mockRepo)

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.Reset()

		// Create a test record
		now := time.Now()
		createReq := dto.CreateCostDataPointRequest{
			Category: "Housing",
			ItemName: "Test Apartment",
			Price:    50000,
			Location: dto.LocationDTO{Emirate: "Dubai"},
			Source:   "manual",
		}
		cdp := createReq.ToModel()
		cdp.ID = "test-id-1"
		cdp.RecordedAt = now
		err := mockRepo.Create(nil, cdp)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/cost-data-points/test-id-1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("test-id-1")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err = handler.Delete(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, 1, mockRepo.GetCallCount("Delete"))
	})

	t.Run("delete not found", func(t *testing.T) {
		mockRepo.Reset()

		now := time.Now()
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/cost-data-points/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")
		c.QueryParams().Add("recorded_at", now.Format(time.RFC3339))

		err := handler.Delete(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("missing recorded_at", func(t *testing.T) {
		mockRepo.Reset()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/cost-data-points/test-id-1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("test-id-1")

		err := handler.Delete(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}
