package dto

import (
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// CreateCostDataPointRequest represents the request body for creating a cost data point
type CreateCostDataPointRequest struct {
	Category    string                 `json:"category" validate:"required"`
	SubCategory string                 `json:"sub_category"`
	ItemName    string                 `json:"item_name" validate:"required"`
	Price       float64                `json:"price" validate:"required,gt=0"`
	MinPrice    float64                `json:"min_price,omitempty"`
	MaxPrice    float64                `json:"max_price,omitempty"`
	MedianPrice float64                `json:"median_price,omitempty"`
	SampleSize  int                    `json:"sample_size,omitempty"`
	Location    LocationDTO            `json:"location" validate:"required"`
	RecordedAt  *time.Time             `json:"recorded_at,omitempty"`
	ValidFrom   *time.Time             `json:"valid_from,omitempty"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	Source      string                 `json:"source" validate:"required"`
	SourceURL   string                 `json:"source_url,omitempty"`
	Confidence  float32                `json:"confidence,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// UpdateCostDataPointRequest represents the request body for updating a cost data point
type UpdateCostDataPointRequest struct {
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category"`
	ItemName    string                 `json:"item_name"`
	Price       float64                `json:"price"`
	MinPrice    float64                `json:"min_price,omitempty"`
	MaxPrice    float64                `json:"max_price,omitempty"`
	MedianPrice float64                `json:"median_price,omitempty"`
	SampleSize  int                    `json:"sample_size,omitempty"`
	Location    *LocationDTO           `json:"location"`
	ValidFrom   *time.Time             `json:"valid_from,omitempty"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	Source      string                 `json:"source"`
	SourceURL   string                 `json:"source_url,omitempty"`
	Confidence  float32                `json:"confidence,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// LocationDTO represents the location information
type LocationDTO struct {
	Emirate     string        `json:"emirate" validate:"required"`
	City        string        `json:"city,omitempty"`
	Area        string        `json:"area,omitempty"`
	Coordinates *GeoPointDTO  `json:"coordinates,omitempty"`
}

// GeoPointDTO represents geographic coordinates
type GeoPointDTO struct {
	Lat float64 `json:"lat" validate:"required"`
	Lon float64 `json:"lon" validate:"required"`
}

// CostDataPointResponse represents the response body for a cost data point
type CostDataPointResponse struct {
	ID          string                 `json:"id"`
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category,omitempty"`
	ItemName    string                 `json:"item_name"`
	Price       float64                `json:"price"`
	MinPrice    float64                `json:"min_price,omitempty"`
	MaxPrice    float64                `json:"max_price,omitempty"`
	MedianPrice float64                `json:"median_price,omitempty"`
	SampleSize  int                    `json:"sample_size"`
	Location    LocationDTO            `json:"location"`
	RecordedAt  time.Time              `json:"recorded_at"`
	ValidFrom   time.Time              `json:"valid_from"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	Source      string                 `json:"source"`
	SourceURL   string                 `json:"source_url,omitempty"`
	Confidence  float32                `json:"confidence"`
	Unit        string                 `json:"unit"`
	Tags        []string               `json:"tags,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data       []CostDataPointResponse `json:"data"`
	TotalCount int                     `json:"total_count"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
}

// ToModel converts CreateCostDataPointRequest to models.CostDataPoint
func (r *CreateCostDataPointRequest) ToModel() *models.CostDataPoint {
	cdp := &models.CostDataPoint{
		Category:    r.Category,
		SubCategory: r.SubCategory,
		ItemName:    r.ItemName,
		Price:       r.Price,
		MinPrice:    r.MinPrice,
		MaxPrice:    r.MaxPrice,
		MedianPrice: r.MedianPrice,
		SampleSize:  r.SampleSize,
		Location:    r.Location.ToModel(),
		Source:      r.Source,
		SourceURL:   r.SourceURL,
		Confidence:  r.Confidence,
		Unit:        r.Unit,
		Tags:        r.Tags,
		Attributes:  r.Attributes,
	}

	if r.RecordedAt != nil {
		cdp.RecordedAt = *r.RecordedAt
	}
	if r.ValidFrom != nil {
		cdp.ValidFrom = *r.ValidFrom
	}
	if r.ValidTo != nil {
		cdp.ValidTo = r.ValidTo
	}

	return cdp
}

// ToModel converts LocationDTO to models.Location
func (l *LocationDTO) ToModel() models.Location {
	loc := models.Location{
		Emirate: l.Emirate,
		City:    l.City,
		Area:    l.Area,
	}

	if l.Coordinates != nil {
		loc.Coordinates = &models.GeoPoint{
			Lat: l.Coordinates.Lat,
			Lon: l.Coordinates.Lon,
		}
	}

	return loc
}

// FromModel converts models.CostDataPoint to CostDataPointResponse
func FromModel(cdp *models.CostDataPoint) CostDataPointResponse {
	return CostDataPointResponse{
		ID:          cdp.ID,
		Category:    cdp.Category,
		SubCategory: cdp.SubCategory,
		ItemName:    cdp.ItemName,
		Price:       cdp.Price,
		MinPrice:    cdp.MinPrice,
		MaxPrice:    cdp.MaxPrice,
		MedianPrice: cdp.MedianPrice,
		SampleSize:  cdp.SampleSize,
		Location:    FromLocationModel(cdp.Location),
		RecordedAt:  cdp.RecordedAt,
		ValidFrom:   cdp.ValidFrom,
		ValidTo:     cdp.ValidTo,
		Source:      cdp.Source,
		SourceURL:   cdp.SourceURL,
		Confidence:  cdp.Confidence,
		Unit:        cdp.Unit,
		Tags:        cdp.Tags,
		Attributes:  cdp.Attributes,
		CreatedAt:   cdp.CreatedAt,
		UpdatedAt:   cdp.UpdatedAt,
	}
}

// FromLocationModel converts models.Location to LocationDTO
func FromLocationModel(loc models.Location) LocationDTO {
	dto := LocationDTO{
		Emirate: loc.Emirate,
		City:    loc.City,
		Area:    loc.Area,
	}

	if loc.Coordinates != nil {
		dto.Coordinates = &GeoPointDTO{
			Lat: loc.Coordinates.Lat,
			Lon: loc.Coordinates.Lon,
		}
	}

	return dto
}

// ApplyUpdate applies UpdateCostDataPointRequest to an existing CostDataPoint
func (r *UpdateCostDataPointRequest) ApplyUpdate(cdp *models.CostDataPoint) {
	// Only update non-zero/non-empty fields
	if r.Category != "" {
		cdp.Category = r.Category
	}
	if r.SubCategory != "" {
		cdp.SubCategory = r.SubCategory
	}
	if r.ItemName != "" {
		cdp.ItemName = r.ItemName
	}
	if r.Price > 0 {
		cdp.Price = r.Price
	}
	if r.MinPrice > 0 {
		cdp.MinPrice = r.MinPrice
	}
	if r.MaxPrice > 0 {
		cdp.MaxPrice = r.MaxPrice
	}
	if r.MedianPrice > 0 {
		cdp.MedianPrice = r.MedianPrice
	}
	if r.SampleSize > 0 {
		cdp.SampleSize = r.SampleSize
	}
	if r.Location != nil {
		cdp.Location = r.Location.ToModel()
	}
	if r.ValidFrom != nil {
		cdp.ValidFrom = *r.ValidFrom
	}
	if r.ValidTo != nil {
		cdp.ValidTo = r.ValidTo
	}
	if r.Source != "" {
		cdp.Source = r.Source
	}
	if r.SourceURL != "" {
		cdp.SourceURL = r.SourceURL
	}
	if r.Confidence > 0 {
		cdp.Confidence = r.Confidence
	}
	if r.Unit != "" {
		cdp.Unit = r.Unit
	}
	if len(r.Tags) > 0 {
		cdp.Tags = r.Tags
	}
	if len(r.Attributes) > 0 {
		cdp.Attributes = r.Attributes
	}
}
