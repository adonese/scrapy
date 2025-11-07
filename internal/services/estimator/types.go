package estimator

import (
	"fmt"
	"strings"
	"time"
)

// Lifestyle represents the qualitative spending style supplied by the user.
type Lifestyle string

const (
	LifestyleBudget   Lifestyle = "budget"
	LifestyleModerate Lifestyle = "moderate"
	LifestylePremium  Lifestyle = "premium"
)

// HousingType captures the high-level shelter choice.
type HousingType string

const (
	HousingApartment HousingType = "apartment"
	HousingVilla     HousingType = "villa"
	HousingShared    HousingType = "shared"
)

// TransportMode is how the persona primarily commutes.
type TransportMode string

const (
	TransportPublic    TransportMode = "public"
	TransportRideshare TransportMode = "rideshare"
	TransportMixed     TransportMode = "mixed"
)

// PersonaInput is supplied by end users (UI/API) to model their household costs.
type PersonaInput struct {
	Adults            int           `json:"adults"`
	Children          int           `json:"children"`
	Bedrooms          int           `json:"bedrooms"`
	HousingType       HousingType   `json:"housing_type"`
	Lifestyle         Lifestyle     `json:"lifestyle"`
	Emirate           string        `json:"emirate"`
	TransportMode     TransportMode `json:"transport_mode"`
	CommuteDistanceKM float64       `json:"commute_distance_km"`
	WorkDaysPerWeek   int           `json:"work_days_per_week"`
}

// Normalize ensures baseline defaults to simplify later logic.
func (p PersonaInput) Normalize() PersonaInput {
	if p.Adults <= 0 {
		p.Adults = 1
	}
	if p.Bedrooms <= 0 {
		p.Bedrooms = 1
	}
	if p.WorkDaysPerWeek <= 0 {
		p.WorkDaysPerWeek = 5
	}
	if p.CommuteDistanceKM <= 0 {
		p.CommuteDistanceKM = 18 // average Dubai commute (km)
	}
	if p.TransportMode == "" {
		p.TransportMode = TransportMixed
	}
	if p.HousingType == "" {
		p.HousingType = HousingApartment
	}
	if p.Lifestyle == "" {
		p.Lifestyle = LifestyleModerate
	}
	p.Emirate = strings.TrimSpace(p.Emirate)
	return p
}

// Validate ensures the persona is usable. Returns slice to keep UX friendly.
func (p PersonaInput) Validate() []error {
	var errs []error
	if strings.TrimSpace(p.Emirate) == "" {
		errs = append(errs, fmt.Errorf("emirate is required"))
	}
	if p.Adults <= 0 {
		errs = append(errs, fmt.Errorf("at least one adult is required"))
	}
	if p.Children < 0 {
		errs = append(errs, fmt.Errorf("children cannot be negative"))
	}
	if p.Bedrooms <= 0 {
		errs = append(errs, fmt.Errorf("bedrooms must be >= 1"))
	}
	if !isValidLifestyle(p.Lifestyle) {
		errs = append(errs, fmt.Errorf("unsupported lifestyle %q", p.Lifestyle))
	}
	if !isValidHousingType(p.HousingType) {
		errs = append(errs, fmt.Errorf("unsupported housing_type %q", p.HousingType))
	}
	if !isValidTransportMode(p.TransportMode) {
		errs = append(errs, fmt.Errorf("unsupported transport_mode %q", p.TransportMode))
	}
	return errs
}

func isValidLifestyle(l Lifestyle) bool {
	switch l {
	case LifestyleBudget, LifestyleModerate, LifestylePremium:
		return true
	default:
		return false
	}
}

func isValidHousingType(ht HousingType) bool {
	switch ht {
	case HousingApartment, HousingVilla, HousingShared:
		return true
	default:
		return false
	}
}

func isValidTransportMode(tm TransportMode) bool {
	switch tm {
	case TransportPublic, TransportRideshare, TransportMixed:
		return true
	default:
		return false
	}
}

// CategoryEstimate represents one budget slice returned to clients.
type CategoryEstimate struct {
	Category     string    `json:"category"`
	MonthlyAED   float64   `json:"monthly_aed"`
	RangeLowAED  float64   `json:"range_low_aed"`
	RangeHighAED float64   `json:"range_high_aed"`
	SampleSize   int       `json:"sample_size"`
	Sources      []string  `json:"sources"`
	Confidence   float32   `json:"confidence"`
	Method       string    `json:"method"`
	Notes        []string  `json:"notes,omitempty"`
	LastUpdated  time.Time `json:"last_updated"`
}

// DatasetSnapshot helps the UI show freshness + coverage.
type DatasetSnapshot struct {
	TotalSamples int            `json:"total_samples"`
	Categories   map[string]int `json:"categories"`
	LastUpdated  time.Time      `json:"last_updated"`
	Coverage     []string       `json:"coverage"`
	Warnings     []string       `json:"warnings"`
}

// EstimateResult is the response returned by the estimator service/API.
type EstimateResult struct {
	Persona         PersonaInput       `json:"persona"`
	Currency        string             `json:"currency"`
	MonthlyTotalAED float64            `json:"monthly_total_aed"`
	Breakdown       []CategoryEstimate `json:"breakdown"`
	Recommendations []string           `json:"recommendations"`
	Dataset         DatasetSnapshot    `json:"dataset"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

// Config tweaks the estimator behaviour.
type Config struct {
	LookbackDays           int
	HousingSampleLimit     int
	UtilitySampleLimit     int
	TransportSampleLimit   int
	Currency               string
	LifestyleMultipliers   map[Lifestyle]float64
	HousingTypeMultipliers map[HousingType]float64
	BedroomStepPercent     float64
}

// DefaultConfig wires pragmatic defaults.
func DefaultConfig() Config {
	return Config{
		LookbackDays:         45,
		HousingSampleLimit:   60,
		UtilitySampleLimit:   80,
		TransportSampleLimit: 80,
		Currency:             "AED",
		LifestyleMultipliers: map[Lifestyle]float64{
			LifestyleBudget:   0.9,
			LifestyleModerate: 1.0,
			LifestylePremium:  1.2,
		},
		HousingTypeMultipliers: map[HousingType]float64{
			HousingApartment: 1.0,
			HousingVilla:     1.35,
			HousingShared:    0.45,
		},
		BedroomStepPercent: 0.12, // each bedroom beyond 1 adds 12%
	}
}
