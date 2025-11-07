package careem

import (
	"testing"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func TestParseRatesToDataPoints(t *testing.T) {
	tests := []struct {
		name          string
		rates         *CareemRates
		wantCount     int
		wantErr       bool
		checkItemName string
	}{
		{
			name: "complete rates",
			rates: &CareemRates{
				ServiceType:             "careem_go",
				Emirate:                 "Dubai",
				BaseFare:                8.0,
				PerKm:                   1.97,
				PerMinuteWait:           0.5,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 1.5,
				AirportSurcharge:        20.0,
				SalikToll:               5.0,
				EffectiveDate:           "2025-01-01",
				Source:                  "test",
				Confidence:              0.8,
			},
			wantCount:     7, // base, per_km, per_minute, minimum, peak, airport, salik
			wantErr:       false,
			checkItemName: "Base Fare",
		},
		{
			name: "rates with service types",
			rates: &CareemRates{
				ServiceType:   "careem_go",
				Emirate:       "Dubai",
				BaseFare:      8.0,
				PerKm:         1.97,
				MinimumFare:   12.0,
				EffectiveDate: "2025-01-01",
				Source:        "test",
				Confidence:    0.8,
				Rates: []ServiceRate{
					{
						ServiceType: "careem_go_plus",
						Description: "Premium option",
						BaseFare:    10.0,
						PerKm:       2.45,
						MinimumFare: 15.0,
					},
				},
			},
			wantCount: 6, // 3 main rates (base, per_km, minimum) + 3 service rates
			wantErr:   false,
		},
		{
			name:      "nil rates",
			rates:     nil,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "minimal rates",
			rates: &CareemRates{
				ServiceType:   "careem_go",
				Emirate:       "Dubai",
				BaseFare:      8.0,
				EffectiveDate: "2025-01-01",
				Source:        "test",
				Confidence:    0.8,
			},
			wantCount: 1, // only base fare
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataPoints, err := ParseRatesToDataPoints(tt.rates)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRatesToDataPoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(dataPoints) != tt.wantCount {
					t.Errorf("ParseRatesToDataPoints() got %d data points, want %d", len(dataPoints), tt.wantCount)
				}

				if len(dataPoints) > 0 {
					// Check first data point
					dp := dataPoints[0]
					if dp.Category != "Transportation" {
						t.Errorf("Category = %v, want Transportation", dp.Category)
					}
					if dp.SubCategory != "Ride Sharing" {
						t.Errorf("SubCategory = %v, want Ride Sharing", dp.SubCategory)
					}
					if dp.Unit != "AED" {
						t.Errorf("Unit = %v, want AED", dp.Unit)
					}
					if dp.Location.Emirate != tt.rates.Emirate {
						t.Errorf("Location.Emirate = %v, want %v", dp.Location.Emirate, tt.rates.Emirate)
					}
				}
			}
		})
	}
}

func TestParseServiceRate(t *testing.T) {
	parentRates := &CareemRates{
		ServiceType:   "careem_go",
		Emirate:       "Dubai",
		EffectiveDate: "2025-01-01",
		Source:        "test",
		Confidence:    0.85,
	}

	serviceRate := ServiceRate{
		ServiceType:   "careem_go_plus",
		Description:   "Premium service",
		BaseFare:      10.0,
		PerKm:         2.45,
		PerMinuteWait: 0.6,
		MinimumFare:   15.0,
	}

	location := models.Location{Emirate: "Dubai", City: "Dubai"}
	now := time.Now()

	dataPoints := parseServiceRate(serviceRate, parentRates, location, now, now)

	if len(dataPoints) != 4 {
		t.Errorf("parseServiceRate() got %d data points, want 4", len(dataPoints))
	}

	// Check that service type is in attributes
	for _, dp := range dataPoints {
		if serviceType, ok := dp.Attributes["service_type"]; !ok || serviceType != "careem_go_plus" {
			t.Errorf("service_type attribute = %v, want careem_go_plus", serviceType)
		}
	}
}

func TestCreateRateDataPoint(t *testing.T) {
	rates := &CareemRates{
		ServiceType:   "careem_go",
		Emirate:       "Dubai",
		EffectiveDate: "2025-01-01",
		Source:        "test_source",
		Confidence:    0.9,
	}

	location := models.Location{Emirate: "Dubai", City: "Dubai"}
	now := time.Now()

	dp := createRateDataPoint(
		"Test Rate",
		10.0,
		BaseFareComponent,
		rates,
		location,
		now,
		now,
	)

	if dp.ItemName != "Test Rate" {
		t.Errorf("ItemName = %v, want Test Rate", dp.ItemName)
	}
	if dp.Price != 10.0 {
		t.Errorf("Price = %v, want 10.0", dp.Price)
	}
	if dp.Category != "Transportation" {
		t.Errorf("Category = %v, want Transportation", dp.Category)
	}
	if dp.SubCategory != "Ride Sharing" {
		t.Errorf("SubCategory = %v, want Ride Sharing", dp.SubCategory)
	}
	if dp.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", dp.Confidence)
	}
	if dp.Attributes["rate_type"] != string(BaseFareComponent) {
		t.Errorf("rate_type attribute = %v, want %v", dp.Attributes["rate_type"], BaseFareComponent)
	}
}

func TestValidateRates(t *testing.T) {
	tests := []struct {
		name    string
		rates   *CareemRates
		wantErr bool
	}{
		{
			name: "valid rates",
			rates: &CareemRates{
				BaseFare:                8.0,
				PerKm:                   1.97,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 1.5,
			},
			wantErr: false,
		},
		{
			name:    "nil rates",
			rates:   nil,
			wantErr: true,
		},
		{
			name: "zero base fare",
			rates: &CareemRates{
				BaseFare:    0,
				PerKm:       1.97,
				MinimumFare: 12.0,
			},
			wantErr: true,
		},
		{
			name: "zero per km",
			rates: &CareemRates{
				BaseFare:    8.0,
				PerKm:       0,
				MinimumFare: 12.0,
			},
			wantErr: true,
		},
		{
			name: "minimum fare less than base fare",
			rates: &CareemRates{
				BaseFare:    10.0,
				PerKm:       1.97,
				MinimumFare: 5.0,
			},
			wantErr: true,
		},
		{
			name: "invalid peak surcharge",
			rates: &CareemRates{
				BaseFare:                8.0,
				PerKm:                   1.97,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 0.5,
			},
			wantErr: true,
		},
		{
			name: "rates too high",
			rates: &CareemRates{
				BaseFare:    150.0,
				PerKm:       1.97,
				MinimumFare: 160.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRates(tt.rates)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectRateChange(t *testing.T) {
	oldRates := &CareemRates{
		BaseFare:                8.0,
		PerKm:                   1.97,
		MinimumFare:             12.0,
		PeakSurchargeMultiplier: 1.5,
	}

	tests := []struct {
		name      string
		oldRates  *CareemRates
		newRates  *CareemRates
		threshold float64
		wantCount int
	}{
		{
			name:     "no change",
			oldRates: oldRates,
			newRates: &CareemRates{
				BaseFare:                8.0,
				PerKm:                   1.97,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 1.5,
			},
			threshold: 10.0,
			wantCount: 0,
		},
		{
			name:     "base fare increased",
			oldRates: oldRates,
			newRates: &CareemRates{
				BaseFare:                10.0, // +25%
				PerKm:                   1.97,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 1.5,
			},
			threshold: 10.0,
			wantCount: 1,
		},
		{
			name:     "multiple changes",
			oldRates: oldRates,
			newRates: &CareemRates{
				BaseFare:                10.0,  // +25%
				PerKm:                   2.5,   // +26%
				MinimumFare:             15.0,  // +25%
				PeakSurchargeMultiplier: 1.5,
			},
			threshold: 10.0,
			wantCount: 3,
		},
		{
			name:     "change below threshold",
			oldRates: oldRates,
			newRates: &CareemRates{
				BaseFare:                8.5, // +6.25%
				PerKm:                   1.97,
				MinimumFare:             12.0,
				PeakSurchargeMultiplier: 1.5,
			},
			threshold: 10.0,
			wantCount: 0,
		},
		{
			name:      "nil old rates",
			oldRates:  nil,
			newRates:  oldRates,
			threshold: 10.0,
			wantCount: 0,
		},
		{
			name:      "nil new rates",
			oldRates:  oldRates,
			newRates:  nil,
			threshold: 10.0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := DetectRateChange(tt.oldRates, tt.newRates, tt.threshold)
			if len(changes) != tt.wantCount {
				t.Errorf("DetectRateChange() got %d changes, want %d", len(changes), tt.wantCount)
				t.Logf("Changes: %v", changes)
			}
		})
	}
}

func TestCalculateChangePercent(t *testing.T) {
	tests := []struct {
		name     string
		oldValue float64
		newValue float64
		want     float64
	}{
		{
			name:     "no change",
			oldValue: 10.0,
			newValue: 10.0,
			want:     0.0,
		},
		{
			name:     "increase",
			oldValue: 10.0,
			newValue: 15.0,
			want:     50.0,
		},
		{
			name:     "decrease",
			oldValue: 10.0,
			newValue: 5.0,
			want:     50.0, // absolute value
		},
		{
			name:     "zero to non-zero",
			oldValue: 0.0,
			newValue: 10.0,
			want:     100.0,
		},
		{
			name:     "both zero",
			oldValue: 0.0,
			newValue: 0.0,
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateChangePercent(tt.oldValue, tt.newValue)
			if got != tt.want {
				t.Errorf("calculateChangePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateFare(t *testing.T) {
	rates := &CareemRates{
		BaseFare:                8.0,
		PerKm:                   1.97,
		PerMinuteWait:           0.5,
		MinimumFare:             12.0,
		PeakSurchargeMultiplier: 1.5,
		AirportSurcharge:        20.0,
		SalikToll:               5.0,
	}

	tests := []struct {
		name            string
		distanceKm      float64
		waitTimeMinutes float64
		isPeakHour      bool
		isAirport       bool
		salikGates      int
		wantMin         float64
		wantMax         float64
	}{
		{
			name:            "short trip minimum fare",
			distanceKm:      1.0,
			waitTimeMinutes: 0,
			isPeakHour:      false,
			isAirport:       false,
			salikGates:      0,
			wantMin:         12.0,
			wantMax:         12.0,
		},
		{
			name:            "normal trip",
			distanceKm:      10.0,
			waitTimeMinutes: 5.0,
			isPeakHour:      false,
			isAirport:       false,
			salikGates:      0,
			wantMin:         28.0, // 8 + 19.7 + 2.5 = 30.2
			wantMax:         32.0,
		},
		{
			name:            "peak hour trip",
			distanceKm:      10.0,
			waitTimeMinutes: 5.0,
			isPeakHour:      true,
			isAirport:       false,
			salikGates:      0,
			wantMin:         43.0, // (8 + 19.7 + 2.5) * 1.5 = 45.3
			wantMax:         48.0,
		},
		{
			name:            "airport trip",
			distanceKm:      10.0,
			waitTimeMinutes: 5.0,
			isPeakHour:      false,
			isAirport:       true,
			salikGates:      2,
			wantMin:         50.0, // 8 + 19.7 + 2.5 + 20 + 10 = 60.2
			wantMax:         62.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fare := EstimateFare(rates, tt.distanceKm, tt.waitTimeMinutes, tt.isPeakHour, tt.isAirport, tt.salikGates)
			if fare < tt.wantMin || fare > tt.wantMax {
				t.Errorf("EstimateFare() = %v, want between %v and %v", fare, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEstimateFare_NilRates(t *testing.T) {
	fare := EstimateFare(nil, 10.0, 5.0, false, false, 0)
	if fare != 0 {
		t.Errorf("EstimateFare(nil) = %v, want 0", fare)
	}
}
