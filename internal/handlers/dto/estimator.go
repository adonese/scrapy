package dto

import "github.com/adonese/cost-of-living/internal/services/estimator"

// EstimateRequest is the payload accepted by /api/v1/estimates.
type EstimateRequest struct {
	Adults            int     `json:"adults" validate:"required,min=1"`
	Children          int     `json:"children" validate:"min=0"`
	Bedrooms          int     `json:"bedrooms" validate:"required,min=1"`
	HousingType       string  `json:"housing_type" validate:"required,oneof=apartment villa shared"`
	Lifestyle         string  `json:"lifestyle" validate:"required,oneof=budget moderate premium"`
	Emirate           string  `json:"emirate" validate:"required"`
	TransportMode     string  `json:"transport_mode" validate:"required,oneof=public rideshare mixed"`
	CommuteDistanceKM float64 `json:"commute_distance_km"`
	WorkDaysPerWeek   int     `json:"work_days_per_week"`
}

// ToPersona converts request payload into the estimator domain input.
func (r EstimateRequest) ToPersona() estimator.PersonaInput {
	return estimator.PersonaInput{
		Adults:            r.Adults,
		Children:          r.Children,
		Bedrooms:          r.Bedrooms,
		HousingType:       estimator.HousingType(r.HousingType),
		Lifestyle:         estimator.Lifestyle(r.Lifestyle),
		Emirate:           r.Emirate,
		TransportMode:     estimator.TransportMode(r.TransportMode),
		CommuteDistanceKM: r.CommuteDistanceKM,
		WorkDaysPerWeek:   r.WorkDaysPerWeek,
	}
}
