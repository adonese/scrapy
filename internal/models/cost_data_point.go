package models

import "time"

// CostDataPoint represents a single cost data point with temporal characteristics
type CostDataPoint struct {
	ID          string                 `json:"id"`
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category,omitempty"`
	ItemName    string                 `json:"item_name"`
	Price       float64                `json:"price"`
	MinPrice    float64                `json:"min_price,omitempty"`
	MaxPrice    float64                `json:"max_price,omitempty"`
	MedianPrice float64                `json:"median_price,omitempty"`
	SampleSize  int                    `json:"sample_size"`
	Location    Location               `json:"location"`
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

// Location represents the geographic location of a cost data point
type Location struct {
	Emirate     string    `json:"emirate"`
	City        string    `json:"city,omitempty"`
	Area        string    `json:"area,omitempty"`
	Coordinates *GeoPoint `json:"coordinates,omitempty"`
}

// GeoPoint represents geographic coordinates
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
