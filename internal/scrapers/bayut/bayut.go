package bayut

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"

	"github.com/adonese/cost-of-living/internal/models"
	"github.com/adonese/cost-of-living/internal/scrapers"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/adonese/cost-of-living/pkg/metrics"
)

// BayutScraper scrapes housing data from Bayut.com
type BayutScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	emirate     string // Dubai, Sharjah, Ajman, Abu Dhabi, etc.
}

// NewBayutScraper creates a new Bayut scraper for Dubai (default)
func NewBayutScraper(config scrapers.Config) *BayutScraper {
	return NewBayutScraperForEmirate(config, "Dubai")
}

// NewBayutScraperForEmirate creates a new Bayut scraper for a specific emirate
func NewBayutScraperForEmirate(config scrapers.Config, emirate string) *BayutScraper {
	return &BayutScraper{
		config:  config,
		emirate: emirate,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
	}
}

// Name returns the scraper identifier
func (s *BayutScraper) Name() string {
	if s.emirate == "Dubai" {
		return "bayut" // Maintain backward compatibility
	}
	return fmt.Sprintf("bayut_%s", strings.ToLower(strings.ReplaceAll(s.emirate, " ", "_")))
}

// CanScrape checks if scraping is possible (rate limit)
func (s *BayutScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches housing data from Bayut
func (s *BayutScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting Bayut scrape", "emirate", s.emirate)

	// Build URL for the specific emirate
	url := s.buildURL()
	logger.Info("Scraping URL", "url", url)

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch the page
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		metrics.ScraperErrorsTotal.WithLabelValues("bayut", "fetch").Inc()
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		metrics.ScraperErrorsTotal.WithLabelValues("bayut", "status").Inc()
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Extract listings
	dataPoints := []*models.CostDataPoint{}

	// Try different selectors for property cards
	// Bayut uses various class names, so we try multiple approaches
	selectors := []string{
		"article[data-testid='property-card']",
		"article.ca2f3a4c",
		"div[aria-label='Property Card']",
		"li[aria-label*='Property']",
	}

	for _, selector := range selectors {
		count := 0
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if i >= 10 { // Limit to first 10 for initial implementation
				return
			}

			cdp := s.extractListing(selection, url)
			if cdp != nil {
				dataPoints = append(dataPoints, cdp)
				count++
			}
		})

		// If we found listings with this selector, stop trying others
		if count > 0 {
			logger.Info("Found listings with selector", "selector", selector, "count", count)
			break
		}
	}

	// If no listings found with specific selectors, try a more general approach
	if len(dataPoints) == 0 {
		logger.Info("No listings found with specific selectors, trying general approach")
		dataPoints = s.extractWithGeneralApproach(doc, url)
	}

	logger.Info("Completed Bayut scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues("bayut").Add(float64(len(dataPoints)))

	return dataPoints, nil
}

// extractListing extracts a single listing from a property card
func (scraper *BayutScraper) extractListing(s *goquery.Selection, baseURL string) *models.CostDataPoint {
	// Extract price - try multiple selectors
	priceText := ""
	priceSelectors := []string{
		"[aria-label='Price']",
		"span[class*='price']",
		"span[class*='Price']",
		"div[class*='price']",
	}

	for _, sel := range priceSelectors {
		priceText = s.Find(sel).First().Text()
		if priceText != "" {
			break
		}
	}

	price := parsePrice(priceText)

	// Extract title/description
	title := ""
	titleSelectors := []string{
		"h2",
		"[aria-label='Property title']",
		"a[title]",
	}

	for _, sel := range titleSelectors {
		if sel == "a[title]" {
			title, _ = s.Find(sel).First().Attr("title")
		} else {
			title = s.Find(sel).First().Text()
		}
		if title != "" {
			break
		}
	}

	// Extract location
	locationText := ""
	locationSelectors := []string{
		"[aria-label='Location']",
		"div[class*='location']",
		"span[class*='location']",
	}

	for _, sel := range locationSelectors {
		locationText = s.Find(sel).First().Text()
		if locationText != "" {
			break
		}
	}

	location := parseLocation(locationText)
	// Ensure emirate is set correctly
	if location.Emirate == "" || location.Emirate == "Dubai" {
		location.Emirate = scraper.emirate
	}
	if location.City == "" {
		location.City = scraper.emirate
	}

	// Extract property details
	bedrooms := ""
	bedroomSelectors := []string{
		"[aria-label='Bedrooms']",
		"span[class*='bed']",
		"div[class*='bed']",
	}

	for _, sel := range bedroomSelectors {
		bedrooms = s.Find(sel).First().Text()
		if bedrooms != "" {
			break
		}
	}

	// Extract URL
	propertyURL := ""
	if href, exists := s.Find("a").First().Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			propertyURL = href
		} else {
			propertyURL = "https://www.bayut.com" + href
		}
	}

	// Validate we have minimum required data
	if price == 0 || title == "" {
		return nil
	}

	now := time.Now()
	return &models.CostDataPoint{
		Category:    "Housing",
		SubCategory: "Rent",
		ItemName:    strings.TrimSpace(title),
		Price:       price,
		Location:    location,
		Source:      "bayut",
		SourceURL:   propertyURL,
		Confidence:  0.8,
		Unit:        "AED",
		RecordedAt:  now,
		ValidFrom:   now,
		SampleSize:  1,
		Tags:        []string{"rent", "apartment", "bayut"},
		Attributes: map[string]interface{}{
			"bedrooms": strings.TrimSpace(bedrooms),
		},
	}
}

// extractWithGeneralApproach tries to extract listings using a more general approach
func (s *BayutScraper) extractWithGeneralApproach(doc *goquery.Document, baseURL string) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Look for any links that seem to be property listings
	doc.Find("a[href*='/property/']").Each(func(i int, link *goquery.Selection) {
		if i >= 10 {
			return
		}

		// Get the parent container (likely the property card)
		card := link.Parent().Parent()

		// Try to extract price from the card
		priceText := card.Find("span, div").FilterFunction(func(_ int, s *goquery.Selection) bool {
			text := s.Text()
			return strings.Contains(text, "AED") || strings.Contains(text, "aed")
		}).First().Text()

		price := parsePrice(priceText)
		if price == 0 {
			return
		}

		// Get title from link
		title, _ := link.Attr("title")
		if title == "" {
			title = link.Text()
		}
		title = strings.TrimSpace(title)

		if title == "" {
			return
		}

		// Get URL
		href, _ := link.Attr("href")
		propertyURL := ""
		if strings.HasPrefix(href, "http") {
			propertyURL = href
		} else {
			propertyURL = "https://www.bayut.com" + href
		}

		now := time.Now()
		dataPoints = append(dataPoints, &models.CostDataPoint{
			Category:    "Housing",
			SubCategory: "Rent",
			ItemName:    title,
			Price:       price,
			Location: models.Location{
				Emirate: s.emirate,
				City:    s.emirate,
			},
			Source:      "bayut",
			SourceURL:   propertyURL,
			Confidence:  0.6, // Lower confidence for general approach
			Unit:        "AED",
			RecordedAt:  now,
			ValidFrom:   now,
			SampleSize:  1,
			Tags:        []string{"rent", "apartment", "bayut"},
			Attributes:  map[string]interface{}{},
		})
	})

	return dataPoints
}

// buildURL constructs the Bayut URL for the emirate
func (s *BayutScraper) buildURL() string {
	emiratePath := s.emirateToURLPath(s.emirate)
	return fmt.Sprintf("https://www.bayut.com/to-rent/apartments/%s/", emiratePath)
}

// emirateToURLPath converts emirate name to Bayut URL format
func (s *BayutScraper) emirateToURLPath(emirate string) string {
	// Bayut URL patterns:
	// Dubai -> dubai
	// Abu Dhabi -> abu-dhabi
	// Sharjah -> sharjah
	// Ajman -> ajman
	// Ras Al Khaimah -> ras-al-khaimah
	// Fujairah -> fujairah
	// Umm Al Quwain -> umm-al-quwain

	emirate = strings.ToLower(emirate)
	emirate = strings.ReplaceAll(emirate, " ", "-")
	return emirate
}
