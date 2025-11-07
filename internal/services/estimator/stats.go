package estimator

import (
	"math"
	"sort"
	"time"

	"github.com/adonese/cost-of-living/internal/models"
)

type summaryStats struct {
	Median      float64
	P25         float64
	P75         float64
	Average     float64
	SampleSize  int
	Sources     []string
	Confidence  float64
	LastUpdated time.Time
}

func computeStats(data []*models.CostDataPoint, transform func(*models.CostDataPoint) float64) summaryStats {
	if len(data) == 0 {
		return summaryStats{}
	}

	values := make([]float64, 0, len(data))
	var sum float64
	var sources = map[string]struct{}{}
	var confidence float64
	var last time.Time

	for _, dp := range data {
		if dp == nil {
			continue
		}
		v := transform(dp)
		if v <= 0 {
			continue
		}
		values = append(values, v)
		sum += v
		if dp.Source != "" {
			sources[dp.Source] = struct{}{}
		}
		confidence += float64(dp.Confidence)
		if dp.RecordedAt.After(last) {
			last = dp.RecordedAt
		}
	}

	if len(values) == 0 {
		return summaryStats{}
	}

	sort.Float64s(values)
	median := percentile(values, 50)
	p25 := percentile(values, 25)
	p75 := percentile(values, 75)

	srcs := make([]string, 0, len(sources))
	for s := range sources {
		srcs = append(srcs, s)
	}
	sort.Strings(srcs)

	avg := sum / float64(len(values))
	conf := confidence / float64(len(values))

	return summaryStats{
		Median:      median,
		P25:         p25,
		P75:         p75,
		Average:     avg,
		SampleSize:  len(values),
		Sources:     srcs,
		Confidence:  conf,
		LastUpdated: last,
	}
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 100 {
		return values[len(values)-1]
	}
	k := (p / 100) * float64(len(values)-1)
	f := int(k)
	c := f + 1
	if c >= len(values) {
		return values[f]
	}
	d := k - float64(f)
	return values[f] + d*(values[c]-values[f])
}

type dataTracker struct {
	total      int
	categories map[string]int
	coverage   map[string]struct{}
	lastUpdate time.Time
	warnings   []string
}

func newDataTracker() *dataTracker {
	return &dataTracker{
		categories: map[string]int{},
		coverage:   map[string]struct{}{},
	}
}

func (d *dataTracker) Track(category string, stats summaryStats) {
	if stats.SampleSize == 0 {
		return
	}
	d.total += stats.SampleSize
	d.categories[category] += stats.SampleSize
	d.coverage[category] = struct{}{}
	if stats.LastUpdated.After(d.lastUpdate) {
		d.lastUpdate = stats.LastUpdated
	}
}

func (d *dataTracker) Warn(msg string) {
	if msg == "" {
		return
	}
	d.warnings = append(d.warnings, msg)
}

func (d *dataTracker) Snapshot() DatasetSnapshot {
	categories := make(map[string]int, len(d.categories))
	for k, v := range d.categories {
		categories[k] = v
	}
	coverage := make([]string, 0, len(d.coverage))
	for k := range d.coverage {
		coverage = append(coverage, k)
	}
	sort.Strings(coverage)

	warnings := make([]string, len(d.warnings))
	copy(warnings, d.warnings)

	return DatasetSnapshot{
		TotalSamples: d.total,
		Categories:   categories,
		LastUpdated:  d.lastUpdate,
		Coverage:     coverage,
		Warnings:     warnings,
	}
}

func roundCurrency(v float64) float64 {
	return math.Round(v*100) / 100
}
