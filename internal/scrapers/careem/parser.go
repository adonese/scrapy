package careem

import (
	"fmt"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

// RateComponent represents a type of rate component
type RateComponent string

const (
	BaseFareComponent       RateComponent = "base_fare"
	PerKmComponent          RateComponent = "per_km"
	PerMinuteComponent      RateComponent = "per_minute_wait"
	MinimumFareComponent    RateComponent = "minimum_fare"
	PeakSurchargeComponent  RateComponent = "peak_surcharge"
	AirportSurchargeComponent RateComponent = "airport_surcharge"
	SalikTollComponent      RateComponent = "salik_toll"
)

// ParseRatesToDataPoints converts CareemRates to cost data points
func ParseRatesToDataPoints(rates *CareemRates) ([]*models.CostDataPoint, error) {
	if rates == nil {
		return nil, fmt.Errorf("rates cannot be nil")
	}

	dataPoints := []*models.CostDataPoint{}
	now := time.Now()

	// Parse effective date
	effectiveDate := now
	if rates.EffectiveDate != "" {
		if parsed, err := time.Parse("2006-01-02", rates.EffectiveDate); err == nil {
			effectiveDate = parsed
		}
	}

	location := models.Location{
		Emirate: rates.Emirate,
		City:    rates.Emirate,
	}

	// Base fare
	if rates.BaseFare > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Base Fare",
			rates.BaseFare,
			BaseFareComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Per kilometer rate
	if rates.PerKm > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Per Kilometer Rate",
			rates.PerKm,
			PerKmComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Per minute waiting rate
	if rates.PerMinuteWait > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Per Minute Wait",
			rates.PerMinuteWait,
			PerMinuteComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Minimum fare
	if rates.MinimumFare > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Minimum Fare",
			rates.MinimumFare,
			MinimumFareComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Peak surcharge (stored as multiplier, convert to example surcharge)
	if rates.PeakSurchargeMultiplier > 1.0 {
		// Calculate example surcharge based on base fare
		exampleSurcharge := rates.BaseFare * (rates.PeakSurchargeMultiplier - 1.0)
		dp := createRateDataPoint(
			"Peak Hour Surcharge",
			exampleSurcharge,
			PeakSurchargeComponent,
			rates,
			location,
			effectiveDate,
			now,
		)
		dp.Attributes["multiplier"] = rates.PeakSurchargeMultiplier
		dp.Attributes["description"] = "Example surcharge on base fare during peak hours"
		dataPoints = append(dataPoints, dp)
	}

	// Airport surcharge
	if rates.AirportSurcharge > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Airport Pickup Surcharge",
			rates.AirportSurcharge,
			AirportSurchargeComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Salik toll
	if rates.SalikToll > 0 {
		dataPoints = append(dataPoints, createRateDataPoint(
			"Salik Toll (per gate)",
			rates.SalikToll,
			SalikTollComponent,
			rates,
			location,
			effectiveDate,
			now,
		))
	}

	// Parse individual service rates if available
	for _, serviceRate := range rates.Rates {
		serviceDataPoints := parseServiceRate(serviceRate, rates, location, effectiveDate, now)
		dataPoints = append(dataPoints, serviceDataPoints...)
	}

	return dataPoints, nil
}

// parseServiceRate converts a ServiceRate to data points
func parseServiceRate(serviceRate ServiceRate, parentRates *CareemRates, location models.Location, effectiveDate, now time.Time) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	prefix := fmt.Sprintf("%s - ", serviceRate.ServiceType)

	if serviceRate.BaseFare > 0 {
		dp := createRateDataPoint(
			prefix+"Base Fare",
			serviceRate.BaseFare,
			BaseFareComponent,
			parentRates,
			location,
			effectiveDate,
			now,
		)
		dp.Attributes["service_type"] = serviceRate.ServiceType
		dp.Attributes["description"] = serviceRate.Description
		dataPoints = append(dataPoints, dp)
	}

	if serviceRate.PerKm > 0 {
		dp := createRateDataPoint(
			prefix+"Per Kilometer",
			serviceRate.PerKm,
			PerKmComponent,
			parentRates,
			location,
			effectiveDate,
			now,
		)
		dp.Attributes["service_type"] = serviceRate.ServiceType
		dataPoints = append(dataPoints, dp)
	}

	if serviceRate.PerMinuteWait > 0 {
		dp := createRateDataPoint(
			prefix+"Per Minute Wait",
			serviceRate.PerMinuteWait,
			PerMinuteComponent,
			parentRates,
			location,
			effectiveDate,
			now,
		)
		dp.Attributes["service_type"] = serviceRate.ServiceType
		dataPoints = append(dataPoints, dp)
	}

	if serviceRate.MinimumFare > 0 {
		dp := createRateDataPoint(
			prefix+"Minimum Fare",
			serviceRate.MinimumFare,
			MinimumFareComponent,
			parentRates,
			location,
			effectiveDate,
			now,
		)
		dp.Attributes["service_type"] = serviceRate.ServiceType
		dataPoints = append(dataPoints, dp)
	}

	return dataPoints
}

// createRateDataPoint creates a cost data point for a rate component
func createRateDataPoint(
	itemName string,
	price float64,
	component RateComponent,
	rates *CareemRates,
	location models.Location,
	effectiveDate, now time.Time,
) *models.CostDataPoint {
	return &models.CostDataPoint{
		Category:    "Transportation",
		SubCategory: "Ride Sharing",
		ItemName:    itemName,
		Price:       price,
		Location:    location,
		Source:      "careem_rates",
		SourceURL:   "aggregated_from_multiple_sources",
		Confidence:  rates.Confidence,
		Unit:        "AED",
		RecordedAt:  now,
		ValidFrom:   effectiveDate,
		SampleSize:  1,
		Tags:        []string{"careem", "ride_sharing", "transportation", string(component)},
		Attributes: map[string]interface{}{
			"rate_type":      string(component),
			"service":        rates.ServiceType,
			"effective_date": rates.EffectiveDate,
			"data_source":    rates.Source,
		},
	}
}

// ValidateRates checks if the rates are reasonable and complete
func ValidateRates(rates *CareemRates) error {
	if rates == nil {
		return fmt.Errorf("rates cannot be nil")
	}

	if rates.BaseFare <= 0 {
		return fmt.Errorf("base fare must be positive")
	}

	if rates.PerKm <= 0 {
		return fmt.Errorf("per km rate must be positive")
	}

	if rates.MinimumFare <= 0 {
		return fmt.Errorf("minimum fare must be positive")
	}

	if rates.MinimumFare < rates.BaseFare {
		return fmt.Errorf("minimum fare cannot be less than base fare")
	}

	if rates.PeakSurchargeMultiplier < 1.0 && rates.PeakSurchargeMultiplier != 0 {
		return fmt.Errorf("peak surcharge multiplier must be >= 1.0 or 0 (no surcharge)")
	}

	// Validate ranges - rates should be within reasonable bounds for UAE
	if rates.BaseFare > 100 {
		return fmt.Errorf("base fare too high: %.2f (expected < 100 AED)", rates.BaseFare)
	}

	if rates.PerKm > 10 {
		return fmt.Errorf("per km rate too high: %.2f (expected < 10 AED)", rates.PerKm)
	}

	if rates.MinimumFare > 200 {
		return fmt.Errorf("minimum fare too high: %.2f (expected < 200 AED)", rates.MinimumFare)
	}

	return nil
}

// DetectRateChange compares old and new rates and returns significant changes
func DetectRateChange(oldRates, newRates *CareemRates, threshold float64) []string {
	if oldRates == nil || newRates == nil {
		return nil
	}

	changes := []string{}

	// Check base fare
	if changePercent := calculateChangePercent(oldRates.BaseFare, newRates.BaseFare); changePercent >= threshold {
		changes = append(changes, fmt.Sprintf("Base fare changed by %.1f%%: %.2f -> %.2f AED",
			changePercent, oldRates.BaseFare, newRates.BaseFare))
	}

	// Check per km rate
	if changePercent := calculateChangePercent(oldRates.PerKm, newRates.PerKm); changePercent >= threshold {
		changes = append(changes, fmt.Sprintf("Per km rate changed by %.1f%%: %.2f -> %.2f AED",
			changePercent, oldRates.PerKm, newRates.PerKm))
	}

	// Check minimum fare
	if changePercent := calculateChangePercent(oldRates.MinimumFare, newRates.MinimumFare); changePercent >= threshold {
		changes = append(changes, fmt.Sprintf("Minimum fare changed by %.1f%%: %.2f -> %.2f AED",
			changePercent, oldRates.MinimumFare, newRates.MinimumFare))
	}

	// Check peak surcharge
	if changePercent := calculateChangePercent(oldRates.PeakSurchargeMultiplier, newRates.PeakSurchargeMultiplier); changePercent >= threshold {
		changes = append(changes, fmt.Sprintf("Peak surcharge multiplier changed by %.1f%%: %.2fx -> %.2fx",
			changePercent, oldRates.PeakSurchargeMultiplier, newRates.PeakSurchargeMultiplier))
	}

	return changes
}

// calculateChangePercent calculates the percentage change between old and new values
func calculateChangePercent(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		if newValue == 0 {
			return 0
		}
		return 100 // New value appeared
	}

	change := ((newValue - oldValue) / oldValue) * 100
	if change < 0 {
		change = -change // Return absolute value
	}
	return change
}

// EstimateFare calculates an estimated fare based on rates
func EstimateFare(rates *CareemRates, distanceKm float64, waitTimeMinutes float64, isPeakHour bool, isAirport bool, salikGates int) float64 {
	if rates == nil {
		return 0
	}

	// Base calculation
	fare := rates.BaseFare
	fare += rates.PerKm * distanceKm
	fare += rates.PerMinuteWait * waitTimeMinutes

	// Apply minimum fare
	if fare < rates.MinimumFare {
		fare = rates.MinimumFare
	}

	// Apply peak hour surcharge
	if isPeakHour && rates.PeakSurchargeMultiplier > 1.0 {
		fare *= rates.PeakSurchargeMultiplier
	}

	// Add airport surcharge
	if isAirport {
		fare += rates.AirportSurcharge
	}

	// Add Salik tolls
	fare += float64(salikGates) * rates.SalikToll

	return fare
}
