package estimator

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

func (s *Service) buildHousingEstimate(ctx context.Context, persona PersonaInput, since time.Time, tracker *dataTracker) (CategoryEstimate, error) {
	subCat := housingSubCategory(persona.HousingType)
	data, err := s.fetchData(ctx, "Housing", subCat, persona.Emirate, s.config.HousingSampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}

	stats := computeStats(data, func(dp *models.CostDataPoint) float64 {
		price := dp.Price
		if strings.EqualFold(dp.SubCategory, "Rent") {
			price = price / 12
		}
		return price
	})

	tracker.Track("Housing", stats)

	estimate := CategoryEstimate{
		Category:    "Housing",
		Sources:     stats.Sources,
		SampleSize:  stats.SampleSize,
		Confidence:  float32(stats.Confidence),
		Method:      "scraped",
		LastUpdated: stats.LastUpdated,
	}

	if stats.SampleSize == 0 {
		fallback := s.fallbackHousing(persona)
		estimate.MonthlyAED = fallback
		estimate.RangeLowAED = fallback * 0.9
		estimate.RangeHighAED = fallback * 1.15
		estimate.Confidence = 0.4
		estimate.Method = "heuristic"
		estimate.LastUpdated = time.Now()
		estimate.Notes = append(estimate.Notes, "No fresh listings matched filters; using heuristic floor tied to bedrooms and lifestyle.")
		tracker.Warn("Housing data fell back to heuristic due to empty dataset.")
		return estimate, nil
	}

	lifestyleMult := s.config.LifestyleMultipliers[persona.Lifestyle]
	if lifestyleMult == 0 {
		lifestyleMult = 1.0
	}

	housingMult := s.config.HousingTypeMultipliers[persona.HousingType]
	if housingMult == 0 {
		housingMult = 1.0
	}

	bedroomMult := 1 + float64(maxInt(persona.Bedrooms-1, 0))*s.config.BedroomStepPercent
	householdMult := 1 + float64(maxInt(persona.Adults+persona.Children-2, 0))*0.05

	apply := func(base float64) float64 {
		return base * lifestyleMult * housingMult * bedroomMult * householdMult
	}

	estimate.MonthlyAED = apply(stats.Median)
	estimate.RangeLowAED = apply(stats.P25)
	estimate.RangeHighAED = apply(stats.P75)
	if estimate.RangeLowAED == 0 {
		estimate.RangeLowAED = estimate.MonthlyAED * 0.9
	}
	if estimate.RangeHighAED == 0 {
		estimate.RangeHighAED = estimate.MonthlyAED * 1.15
	}

	return estimate, nil
}

func (s *Service) buildUtilitiesEstimate(ctx context.Context, persona PersonaInput, since time.Time, tracker *dataTracker) (CategoryEstimate, error) {
	electricityData, err := s.fetchData(ctx, "Utilities", "Electricity", persona.Emirate, s.config.UtilitySampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}
	waterData, err := s.fetchData(ctx, "Utilities", "Water", persona.Emirate, s.config.UtilitySampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}
	surchargeData, err := s.fetchData(ctx, "Utilities", "Fuel Surcharge", persona.Emirate, s.config.UtilitySampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}

	elecStats := computeStats(electricityData, func(dp *models.CostDataPoint) float64 { return dp.Price })
	waterStats := computeStats(waterData, func(dp *models.CostDataPoint) float64 { return dp.Price })
	surchargeStats := computeStats(surchargeData, func(dp *models.CostDataPoint) float64 { return dp.Price })

	tracker.Track("Utilities", elecStats)
	tracker.Track("Utilities", waterStats)
	tracker.Track("Utilities", surchargeStats)

	kwh := float64(persona.Adults)*350 + float64(persona.Children)*180 + float64(persona.Bedrooms)*65
	if kwh < 450 {
		kwh = 450
	}
	waterM3 := float64(persona.Adults)*4.5 + float64(persona.Children)*3.2 + float64(persona.Bedrooms)*1.1
	if waterM3 < 8 {
		waterM3 = 8
	}

	electricityCost := elecStats.Median * kwh
	waterCost := waterStats.Median * waterM3
	surchargeCost := surchargeStats.Median * kwh * 0.1
	serviceFees := 45.0 // DEWA fixed meter fee

	if electricityCost == 0 {
		electricityCost = 0.38 * kwh
	}
	if waterCost == 0 {
		waterCost = 3.0 * waterM3
	}
	if surchargeCost == 0 {
		surchargeCost = 0.05 * kwh
	}

	base := electricityCost + waterCost + surchargeCost + serviceFees

	lifestyleMult := s.config.LifestyleMultipliers[persona.Lifestyle]
	if lifestyleMult == 0 {
		lifestyleMult = 1
	}

	estimate := CategoryEstimate{
		Category:     "Utilities",
		MonthlyAED:   base * lifestyleMult,
		RangeLowAED:  base * 0.9 * lifestyleMult,
		RangeHighAED: base * 1.15 * lifestyleMult,
		SampleSize:   elecStats.SampleSize + waterStats.SampleSize + surchargeStats.SampleSize,
		Sources:      mergeSources(elecStats.Sources, waterStats.Sources, surchargeStats.Sources),
		Confidence:   float32(math.Min(1, (elecStats.Confidence+waterStats.Confidence+surchargeStats.Confidence)/3)),
		Method:       "scraped",
		LastUpdated:  maxTime(elecStats.LastUpdated, maxTime(waterStats.LastUpdated, surchargeStats.LastUpdated)),
	}

	if estimate.SampleSize == 0 {
		estimate.Method = "heuristic"
		estimate.LastUpdated = time.Now()
		estimate.Confidence = 0.5
		estimate.Notes = append(estimate.Notes, "Using heuristic DEWA reference slab because no fresh utility data was available.")
		tracker.Warn("Utilities fell back to heuristic slab rate.")
	}

	return estimate, nil
}

func (s *Service) buildTransportEstimate(ctx context.Context, persona PersonaInput, since time.Time, tracker *dataTracker) (CategoryEstimate, error) {
	publicData, err := s.fetchData(ctx, "Transportation", "Public Transport", persona.Emirate, s.config.TransportSampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}
	taxiData, err := s.fetchData(ctx, "Transportation", "Taxi", persona.Emirate, s.config.TransportSampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}
	rideShareData, err := s.fetchData(ctx, "Transportation", "Ride Sharing", persona.Emirate, s.config.TransportSampleLimit, since)
	if err != nil {
		return CategoryEstimate{}, err
	}

	publicStats := computeStats(publicData, func(dp *models.CostDataPoint) float64 { return dp.Price })
	taxiStats := computeStats(taxiData, func(dp *models.CostDataPoint) float64 { return dp.Price })
	rideStats := computeStats(rideShareData, func(dp *models.CostDataPoint) float64 { return dp.Price })

	tracker.Track("Transportation", publicStats)
	tracker.Track("Transportation", taxiStats)
	tracker.Track("Transportation", rideStats)

	commuteTrips := float64(persona.WorkDaysPerWeek*2) * 4.3
	if commuteTrips < 30 {
		commuteTrips = 30
	}

	publicFare := publicStats.Median
	if publicFare == 0 {
		publicFare = 4.0
	}

	taxiPerKm := taxiStats.Median
	if taxiPerKm == 0 {
		taxiPerKm = 2.5
	}

	rideFare := estimateRideShareTrip(rideShareData, persona.CommuteDistanceKM)
	if rideFare == 0 {
		rideFare = taxiPerKm*persona.CommuteDistanceKM + 8
	}

	monthly := 0.0
	switch persona.TransportMode {
	case TransportPublic:
		monthly = publicFare * commuteTrips
	case TransportRideshare:
		monthly = rideFare*commuteTrips + 4*rideFare // errands
	default:
		monthly = publicFare*commuteTrips*0.65 + rideFare*commuteTrips*0.35
	}

	monthly *= s.config.LifestyleMultipliers[persona.Lifestyle]

	estimate := CategoryEstimate{
		Category:     "Transportation",
		MonthlyAED:   monthly,
		RangeLowAED:  monthly * 0.85,
		RangeHighAED: monthly * 1.2,
		SampleSize:   publicStats.SampleSize + taxiStats.SampleSize + rideStats.SampleSize,
		Sources:      mergeSources(publicStats.Sources, taxiStats.Sources, rideStats.Sources),
		Confidence:   float32(math.Min(1, (publicStats.Confidence+taxiStats.Confidence+rideStats.Confidence)/3)),
		Method:       "scraped",
		LastUpdated:  maxTime(publicStats.LastUpdated, maxTime(taxiStats.LastUpdated, rideStats.LastUpdated)),
	}

	if estimate.SampleSize == 0 {
		estimate.Method = "heuristic"
		estimate.LastUpdated = time.Now()
		estimate.Confidence = 0.45
		estimate.Notes = append(estimate.Notes, "Transport numbers derived from RTA card + typical Careem trip assumptions.")
		tracker.Warn("Transportation fell back to heuristic mixture of RTA + Careem fares.")
	}

	return estimate, nil
}

func (s *Service) buildGroceriesEstimate(persona PersonaInput) CategoryEstimate {
	adults := float64(persona.Adults)
	children := float64(persona.Children)
	base := adults*1100 + children*650
	lifestyleMult := s.config.LifestyleMultipliers[persona.Lifestyle]
	if lifestyleMult == 0 {
		lifestyleMult = 1
	}
	monthly := base * lifestyleMult

	return CategoryEstimate{
		Category:     "Groceries & Essentials",
		MonthlyAED:   monthly,
		RangeLowAED:  monthly * 0.9,
		RangeHighAED: monthly * 1.15,
		Method:       "heuristic",
		Confidence:   0.55,
		SampleSize:   0,
		Sources:      nil,
		LastUpdated:  time.Now(),
		Notes: []string{
			"Scaled per-adult (AED 1100) and per-child (AED 650) basket using Carrefour/Lulu reference carts.",
		},
	}
}

func (s *Service) buildBufferEstimate(persona PersonaInput, deps []CategoryEstimate) CategoryEstimate {
	subtotal := 0.0
	for _, dep := range deps {
		subtotal += dep.MonthlyAED
	}
	lifestyleMult := s.config.LifestyleMultipliers[persona.Lifestyle]
	if lifestyleMult == 0 {
		lifestyleMult = 1
	}

	buffer := math.Max(300, subtotal*0.08*lifestyleMult)
	return CategoryEstimate{
		Category:     "Safety & Lifestyle Buffer",
		MonthlyAED:   buffer,
		RangeLowAED:  buffer * 0.8,
		RangeHighAED: buffer * 1.2,
		Method:       "heuristic",
		Confidence:   0.4,
		LastUpdated:  time.Now(),
		Notes: []string{
			"Covers telecom, healthcare, visa fees, and surprise runs (8% of core spend, min AED 300).",
		},
	}
}

func (s *Service) buildRecommendations(breakdown []CategoryEstimate, persona PersonaInput) []string {
	if len(breakdown) == 0 {
		return nil
	}
	total := 0.0
	for _, b := range breakdown {
		total += b.MonthlyAED
	}
	if total == 0 {
		return nil
	}

	var recs []string
	for _, b := range breakdown {
		share := b.MonthlyAED / total
		if b.Category == "Housing" && share > 0.45 {
			recs = append(recs, "Housing exceeds 45% of spend. Consider exploring outer communities or smaller units.")
		}
		if b.Category == "Transportation" && persona.TransportMode == TransportRideshare && share > 0.18 {
			recs = append(recs, "Ride sharing dominates mobility costs. Switching to RTA weekly passes could save ~30%. ")
		}
		if b.Category == "Utilities" && persona.Bedrooms >= 3 && share > 0.12 {
			recs = append(recs, "Utilities are spiking. Smart thermostats and DEWA efficiency tips usually trim 10-15%.")
		}
	}

	if len(recs) == 0 {
		recs = append(recs, "Mix looks balanced for this lifestyle. Track actual invoices for two months to calibrate further.")
	}
	return recs
}

func housingSubCategory(ht HousingType) string {
	switch ht {
	case HousingShared:
		return "Shared Accommodation"
	default:
		return "Rent"
	}
}

func (s *Service) fallbackHousing(persona PersonaInput) float64 {
	base := 4200.0
	switch persona.Emirate {
	case "Dubai":
		base = 5200
	case "Abu Dhabi":
		base = 5000
	case "Sharjah":
		base = 3600
	case "Ajman":
		base = 3200
	}

	lifestyleMult := s.config.LifestyleMultipliers[persona.Lifestyle]
	if lifestyleMult == 0 {
		lifestyleMult = 1
	}
	housingMult := s.config.HousingTypeMultipliers[persona.HousingType]
	if housingMult == 0 {
		housingMult = 1
	}
	bedroomMult := 1 + float64(maxInt(persona.Bedrooms-1, 0))*s.config.BedroomStepPercent
	return base * lifestyleMult * housingMult * bedroomMult
}

func mergeSources(groups ...[]string) []string {
	uniq := map[string]struct{}{}
	for _, g := range groups {
		for _, src := range g {
			if src == "" {
				continue
			}
			uniq[src] = struct{}{}
		}
	}
	out := make([]string, 0, len(uniq))
	for src := range uniq {
		out = append(out, src)
	}
	sort.Strings(out)
	return out
}

func estimateRideShareTrip(data []*models.CostDataPoint, distance float64) float64 {
	if distance <= 0 {
		distance = 15
	}
	base, perKm, perMinute, minFare := 0.0, 0.0, 0.0, 0.0
	for _, dp := range data {
		rt, _ := dp.Attributes["rate_type"].(string)
		switch rt {
		case "base_fare":
			base = dp.Price
		case "per_km":
			perKm = dp.Price
		case "per_minute_wait":
			perMinute = dp.Price
		case "minimum_fare":
			minFare = dp.Price
		}
	}
	if base == 0 {
		base = 7.0
	}
	if perKm == 0 {
		perKm = 2.6
	}
	if perMinute == 0 {
		perMinute = 0.6
	}
	if minFare == 0 {
		minFare = 14
	}

	durationMinutes := math.Max(18, distance/28*60)
	fare := base + perKm*distance + perMinute*(durationMinutes/3) // assume 1/3 of trip is slow traffic
	if fare < minFare {
		fare = minFare
	}
	return fare
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
