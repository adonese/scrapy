package estimator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/repository"
)

// Service exposes persona-driven cost aggregation logic.
type Service struct {
	repo   repository.CostDataPointRepository
	config Config
}

// NewService builds an estimator service with sane defaults.
func NewService(repo repository.CostDataPointRepository, cfg *Config) *Service {
	if repo == nil {
		panic("estimator: repository cannot be nil")
	}
	finalCfg := DefaultConfig()
	if cfg != nil {
		if cfg.LookbackDays > 0 {
			finalCfg.LookbackDays = cfg.LookbackDays
		}
		if cfg.HousingSampleLimit > 0 {
			finalCfg.HousingSampleLimit = cfg.HousingSampleLimit
		}
		if cfg.UtilitySampleLimit > 0 {
			finalCfg.UtilitySampleLimit = cfg.UtilitySampleLimit
		}
		if cfg.TransportSampleLimit > 0 {
			finalCfg.TransportSampleLimit = cfg.TransportSampleLimit
		}
		if cfg.Currency != "" {
			finalCfg.Currency = cfg.Currency
		}
		if cfg.LifestyleMultipliers != nil {
			finalCfg.LifestyleMultipliers = cfg.LifestyleMultipliers
		}
		if cfg.HousingTypeMultipliers != nil {
			finalCfg.HousingTypeMultipliers = cfg.HousingTypeMultipliers
		}
		if cfg.BedroomStepPercent > 0 {
			finalCfg.BedroomStepPercent = cfg.BedroomStepPercent
		}
	}

	return &Service{repo: repo, config: finalCfg}
}

// Estimate returns a monthly budget breakdown for the supplied persona.
func (s *Service) Estimate(ctx context.Context, persona PersonaInput) (*EstimateResult, error) {
	persona = persona.Normalize()
	if errs := persona.Validate(); len(errs) > 0 {
		return nil, combineErrors(errs)
	}

	since := time.Now().AddDate(0, 0, -s.config.LookbackDays)
	tracker := newDataTracker()

	housing, err := s.buildHousingEstimate(ctx, persona, since, tracker)
	if err != nil {
		return nil, err
	}

	utilities, err := s.buildUtilitiesEstimate(ctx, persona, since, tracker)
	if err != nil {
		return nil, err
	}

	transport, err := s.buildTransportEstimate(ctx, persona, since, tracker)
	if err != nil {
		return nil, err
	}

	groceries := s.buildGroceriesEstimate(persona)
	buffer := s.buildBufferEstimate(persona, []CategoryEstimate{housing, utilities, transport, groceries})

	breakdown := []CategoryEstimate{housing, utilities, transport, groceries, buffer}
	sort.SliceStable(breakdown, func(i, j int) bool {
		return breakdown[i].MonthlyAED > breakdown[j].MonthlyAED
	})

	total := 0.0
	for i := range breakdown {
		breakdown[i].MonthlyAED = roundCurrency(breakdown[i].MonthlyAED)
		breakdown[i].RangeLowAED = roundCurrency(breakdown[i].RangeLowAED)
		breakdown[i].RangeHighAED = roundCurrency(breakdown[i].RangeHighAED)
		total += breakdown[i].MonthlyAED
	}

	res := &EstimateResult{
		Persona:         persona,
		Currency:        s.config.Currency,
		MonthlyTotalAED: roundCurrency(total),
		Breakdown:       breakdown,
		Recommendations: s.buildRecommendations(breakdown, persona),
		Dataset:         tracker.Snapshot(),
		GeneratedAt:     time.Now(),
	}
	return res, nil
}

// Summary returns dataset coverage info for UI/monitoring cards.
func (s *Service) Summary(ctx context.Context, emirate string) (DatasetSnapshot, error) {
	emirate = strings.TrimSpace(emirate)
	if emirate == "" {
		return DatasetSnapshot{}, errors.New("emirate is required")
	}

	persona := PersonaInput{
		Adults:        2,
		Children:      0,
		Bedrooms:      2,
		HousingType:   HousingApartment,
		Lifestyle:     LifestyleModerate,
		Emirate:       emirate,
		TransportMode: TransportMixed,
	}
	persona = persona.Normalize()

	since := time.Now().AddDate(0, 0, -s.config.LookbackDays)
	tracker := newDataTracker()

	if _, err := s.buildHousingEstimate(ctx, persona, since, tracker); err != nil {
		return DatasetSnapshot{}, err
	}
	if _, err := s.buildUtilitiesEstimate(ctx, persona, since, tracker); err != nil {
		return DatasetSnapshot{}, err
	}
	if _, err := s.buildTransportEstimate(ctx, persona, since, tracker); err != nil {
		return DatasetSnapshot{}, err
	}

	// Summary ignores heuristics for groceries/buffer.
	return tracker.Snapshot(), nil
}

func (s *Service) fetchData(ctx context.Context, category string, subCategory string, emirate string, limit int, since time.Time) ([]*models.CostDataPoint, error) {
	filter := repository.ListFilter{
		Category: category,
		Limit:    limit,
	}
	if subCategory != "" {
		filter.SubCategory = subCategory
	}
	if emirate != "" {
		filter.Emirate = emirate
	}
	if !since.IsZero() {
		filter.StartDate = &since
	}

	data, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list %s data: %w", category, err)
	}

	if len(data) == 0 && emirate != "" {
		filter.Emirate = ""
		data, err = s.repo.List(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("list fallback %s data: %w", category, err)
		}
	}

	return data, nil
}

func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var parts []string
	for _, err := range errs {
		parts = append(parts, err.Error())
	}
	return fmt.Errorf("invalid persona: %s", strings.Join(parts, "; "))
}
