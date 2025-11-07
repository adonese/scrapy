package careem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adonese/cost-of-living/pkg/logger"
)

// CareemRates represents the rate structure for Careem services
type CareemRates struct {
	ServiceType              string                 `json:"service_type"`
	Emirate                  string                 `json:"emirate"`
	BaseFare                 float64                `json:"base_fare"`
	PerKm                    float64                `json:"per_km"`
	PerMinuteWait            float64                `json:"per_minute_wait"`
	MinimumFare              float64                `json:"minimum_fare"`
	PeakSurchargeMultiplier  float64                `json:"peak_surcharge_multiplier"`
	AirportSurcharge         float64                `json:"airport_surcharge"`
	SalikToll                float64                `json:"salik_toll"`
	EffectiveDate            string                 `json:"effective_date"`
	Source                   string                 `json:"source"`
	LastUpdated              time.Time              `json:"last_updated"`
	Confidence               float32                `json:"confidence"`
	Rates                    []ServiceRate          `json:"rates,omitempty"`
	Surcharges               map[string]interface{} `json:"surcharges,omitempty"`
	Notes                    []string               `json:"notes,omitempty"`
}

// ServiceRate represents rates for a specific service type
type ServiceRate struct {
	ServiceType   string  `json:"service_type"`
	Description   string  `json:"description"`
	BaseFare      float64 `json:"base_fare"`
	PerKm         float64 `json:"per_km"`
	PerMinuteWait float64 `json:"per_minute_wait"`
	MinimumFare   float64 `json:"minimum_fare"`
}

// RateSource defines the interface for fetching Careem rates from different sources
type RateSource interface {
	// Name returns the source identifier
	Name() string

	// FetchRates fetches rates from this source
	FetchRates(ctx context.Context) (*CareemRates, error)

	// Confidence returns the confidence score for this source (0.0-1.0)
	Confidence() float32

	// IsAvailable checks if this source is currently available
	IsAvailable(ctx context.Context) bool
}

// StaticSource provides fallback rates from a JSON file
type StaticSource struct {
	filePath   string
	confidence float32
}

// NewStaticSource creates a new static source from a file
func NewStaticSource(filePath string) *StaticSource {
	return &StaticSource{
		filePath:   filePath,
		confidence: 0.7, // Lower confidence for static data
	}
}

func (s *StaticSource) Name() string {
	return "static_file"
}

func (s *StaticSource) Confidence() float32 {
	return s.confidence
}

func (s *StaticSource) IsAvailable(ctx context.Context) bool {
	_, err := os.Stat(s.filePath)
	return err == nil
}

func (s *StaticSource) FetchRates(ctx context.Context) (*CareemRates, error) {
	logger.Info("Fetching rates from static source", "file", s.filePath)

	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var rates CareemRates
	if err := json.Unmarshal(data, &rates); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	rates.Source = s.Name()
	rates.Confidence = s.Confidence()
	rates.LastUpdated = time.Now()

	return &rates, nil
}

// HelpCenterSource fetches rates from Careem help center pages
type HelpCenterSource struct {
	baseURL    string
	client     *http.Client
	userAgent  string
	confidence float32
}

// NewHelpCenterSource creates a new help center source
func NewHelpCenterSource(userAgent string) *HelpCenterSource {
	return &HelpCenterSource{
		baseURL:    "https://help.careem.com",
		userAgent:  userAgent,
		confidence: 0.85,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *HelpCenterSource) Name() string {
	return "careem_help_center"
}

func (h *HelpCenterSource) Confidence() float32 {
	return h.confidence
}

func (h *HelpCenterSource) IsAvailable(ctx context.Context) bool {
	// Try to reach the help center
	req, err := http.NewRequestWithContext(ctx, "HEAD", h.baseURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (h *HelpCenterSource) FetchRates(ctx context.Context) (*CareemRates, error) {
	logger.Info("Fetching rates from help center")

	// URLs that might contain rate information
	urls := []string{
		h.baseURL + "/hc/en-us/articles/pricing",
		h.baseURL + "/hc/en-us/articles/fare-estimate",
		h.baseURL + "/hc/en-us/articles/rates",
	}

	for _, url := range urls {
		rates, err := h.scrapeHelpPage(ctx, url)
		if err == nil && rates != nil {
			rates.Source = h.Name()
			rates.Confidence = h.Confidence()
			rates.LastUpdated = time.Now()
			return rates, nil
		}
		logger.Info("Failed to fetch from URL", "url", url, "error", err)
	}

	return nil, fmt.Errorf("no rates found in help center")
}

func (h *HelpCenterSource) scrapeHelpPage(ctx context.Context, url string) (*CareemRates, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Look for rate information in the text
	text := doc.Find("article, .article-body, .content").Text()

	// Parse rates from text (simple pattern matching)
	// This is a simplified version - real implementation would be more robust
	if strings.Contains(strings.ToLower(text), "base fare") {
		// Extract rates using patterns
		// For now, return nil to use fallback
		return nil, fmt.Errorf("rate parsing not implemented")
	}

	return nil, fmt.Errorf("no rates found on page")
}

// NewsSource fetches rates from news articles and press releases
type NewsSource struct {
	searchURLs []string
	client     *http.Client
	userAgent  string
	confidence float32
}

// NewNewsSource creates a new news source
func NewNewsSource(userAgent string) *NewsSource {
	return &NewsSource{
		searchURLs: []string{
			"https://www.google.com/search?q=careem+rates+dubai+2025",
			"https://www.thenationalnews.com/search?q=careem+rates",
		},
		userAgent:  userAgent,
		confidence: 0.75,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (n *NewsSource) Name() string {
	return "news_articles"
}

func (n *NewsSource) Confidence() float32 {
	return n.confidence
}

func (n *NewsSource) IsAvailable(ctx context.Context) bool {
	// News sources are generally available, but we might get blocked
	// For now, return true
	return true
}

func (n *NewsSource) FetchRates(ctx context.Context) (*CareemRates, error) {
	logger.Info("Fetching rates from news sources")

	// In a real implementation, we would:
	// 1. Search for recent news articles about Careem rates
	// 2. Parse the articles for rate information
	// 3. Aggregate and validate the data

	// For now, return an error to use fallback
	return nil, fmt.Errorf("news source scraping not implemented")
}

// APISource would fetch from official Careem API if available
type APISource struct {
	apiKey     string
	baseURL    string
	client     *http.Client
	confidence float32
}

// NewAPISource creates a new API source (if Careem provides an API)
func NewAPISource(apiKey string) *APISource {
	return &APISource{
		apiKey:     apiKey,
		baseURL:    "https://api.careem.com/v1",
		confidence: 0.95,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *APISource) Name() string {
	return "careem_api"
}

func (a *APISource) Confidence() float32 {
	return a.confidence
}

func (a *APISource) IsAvailable(ctx context.Context) bool {
	// Check if API key is provided
	return a.apiKey != ""
}

func (a *APISource) FetchRates(ctx context.Context) (*CareemRates, error) {
	logger.Info("Fetching rates from Careem API")

	// In a real implementation, we would make API calls
	// For now, return an error since there's no official public API
	return nil, fmt.Errorf("careem public API not available")
}

// SourceAggregator tries multiple sources and returns the best result
type SourceAggregator struct {
	sources []RateSource
}

// NewSourceAggregator creates a new source aggregator
func NewSourceAggregator(sources []RateSource) *SourceAggregator {
	return &SourceAggregator{
		sources: sources,
	}
}

// FetchBestRates tries all sources and returns the result with highest confidence
func (sa *SourceAggregator) FetchBestRates(ctx context.Context) (*CareemRates, error) {
	logger.Info("Fetching rates from multiple sources", "count", len(sa.sources))

	var bestRates *CareemRates
	var bestConfidence float32 = 0.0

	for _, source := range sa.sources {
		// Check if source is available
		if !source.IsAvailable(ctx) {
			logger.Info("Source not available", "source", source.Name())
			continue
		}

		// Try to fetch rates
		rates, err := source.FetchRates(ctx)
		if err != nil {
			logger.Info("Failed to fetch from source", "source", source.Name(), "error", err)
			continue
		}

		// Keep the rates with highest confidence
		if rates != nil && rates.Confidence > bestConfidence {
			bestRates = rates
			bestConfidence = rates.Confidence
		}
	}

	if bestRates == nil {
		return nil, fmt.Errorf("no sources available")
	}

	logger.Info("Best rates found", "source", bestRates.Source, "confidence", bestRates.Confidence)
	return bestRates, nil
}

// GetDefaultStaticSourcePath returns the default path to the static rates file
func GetDefaultStaticSourcePath() string {
	// Try to find the fixtures directory
	possiblePaths := []string{
		"test/fixtures/careem/rates.json",
		"../../../test/fixtures/careem/rates.json",
		"../../test/fixtures/careem/rates.json",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Default to relative path
	return "test/fixtures/careem/rates.json"
}
