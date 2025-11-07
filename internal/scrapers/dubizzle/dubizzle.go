package dubizzle

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

// DubizzleScraper scrapes housing data from Dubizzle.com
type DubizzleScraper struct {
	config      scrapers.Config
	client      *http.Client
	rateLimiter *rate.Limiter
	emirate     string // Dubai, Sharjah, Ajman, Abu Dhabi, etc.
	category    string // apartmentflat, bedspace, roomspace
}

// NewDubizzleScraper creates a new Dubizzle scraper for Dubai apartments (default)
func NewDubizzleScraper(config scrapers.Config) *DubizzleScraper {
	return NewDubizzleScraperFor(config, "Dubai", "apartmentflat")
}

// NewDubizzleScraperFor creates a new Dubizzle scraper for a specific emirate and category
func NewDubizzleScraperFor(config scrapers.Config, emirate, category string) *DubizzleScraper {
	return &DubizzleScraper{
		config:   config,
		emirate:  emirate,
		category: category,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
	}
}

// Name returns the scraper identifier
func (s *DubizzleScraper) Name() string {
	if s.emirate == "Dubai" && s.category == "apartmentflat" {
		return "dubizzle" // Maintain backward compatibility
	}
	emirateName := strings.ToLower(strings.ReplaceAll(s.emirate, " ", "_"))
	if s.category == "apartmentflat" {
		return fmt.Sprintf("dubizzle_%s", emirateName)
	}
	return fmt.Sprintf("dubizzle_%s_%s", emirateName, s.category)
}

// CanScrape checks if scraping is possible (rate limit)
func (s *DubizzleScraper) CanScrape() bool {
	return s.rateLimiter.Allow()
}

// Scrape fetches housing data from Dubizzle
func (s *DubizzleScraper) Scrape(ctx context.Context) ([]*models.CostDataPoint, error) {
	logger.Info("Starting Dubizzle scrape", "emirate", s.emirate, "category", s.category)

	// Build URL for the specific emirate and category
	url := s.buildURL()
	logger.Info("Scraping URL", "url", url)

	// Wait for rate limit
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	// Fetch the page with retry logic
	dataPoints, err := s.fetchWithRetry(ctx, url)
	if err != nil {
		return nil, err
	}

	logger.Info("Completed Dubizzle scrape", "count", len(dataPoints))
	metrics.ScraperItemsScraped.WithLabelValues(s.Name()).Add(float64(len(dataPoints)))

	return dataPoints, nil
}

// fetchWithRetry attempts to fetch and parse the page with retries
func (s *DubizzleScraper) fetchWithRetry(ctx context.Context, url string) ([]*models.CostDataPoint, error) {
	var lastErr error

	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait between retries with exponential backoff
			waitTime := time.Duration(attempt) * time.Second
			logger.Info("Retrying fetch", "attempt", attempt+1, "wait", waitTime)
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Fetch the page
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		// Set headers to appear more like a real browser
		req.Header.Set("User-Agent", s.config.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		resp, err := s.client.Do(req)
		if err != nil {
			metrics.ScraperErrorsTotal.WithLabelValues("dubizzle", "fetch").Inc()
			lastErr = fmt.Errorf("fetch page: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check for bot detection or rate limiting
		if resp.StatusCode == 429 || resp.StatusCode == 403 {
			metrics.ScraperErrorsTotal.WithLabelValues("dubizzle", "blocked").Inc()
			lastErr = fmt.Errorf("blocked by anti-bot (status %d)", resp.StatusCode)
			continue
		}

		if resp.StatusCode != 200 {
			metrics.ScraperErrorsTotal.WithLabelValues("dubizzle", "status").Inc()
			lastErr = fmt.Errorf("bad status: %d", resp.StatusCode)
			continue
		}

		// Parse HTML
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("parse html: %w", err)
			continue
		}

		// Check if we got an error page or empty result
		if s.isErrorPage(doc) {
			metrics.ScraperErrorsTotal.WithLabelValues("dubizzle", "error_page").Inc()
			lastErr = fmt.Errorf("received error page (likely anti-bot)")
			continue
		}

		// Extract listings
		dataPoints := s.extractListings(doc, url)

		// If we got results, return them
		if len(dataPoints) > 0 {
			return dataPoints, nil
		}

		// If no results but no error, might be legitimate empty page
		lastErr = fmt.Errorf("no listings found")
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("failed after %d attempts", s.config.MaxRetries)
}

// isErrorPage checks if the document is an error/block page
func (s *DubizzleScraper) isErrorPage(doc *goquery.Document) bool {
	// Check for common anti-bot indicators
	pageText := strings.ToLower(doc.Text())
	indicators := []string{
		"incapsula",
		"access denied",
		"blocked",
		"suspicious activity",
		"captcha",
		"cloudflare",
		"ray id",
	}

	for _, indicator := range indicators {
		if strings.Contains(pageText, indicator) {
			return true
		}
	}

	return false
}

// extractListings extracts all listings from the page
func (s *DubizzleScraper) extractListings(doc *goquery.Document, baseURL string) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Try different selectors for property listings
	// Dubizzle uses various class names and structures
	selectors := []string{
		"li[data-testid='listing-item']",
		"div[data-testid='listing-card']",
		"article.listing",
		"li.listing",
		"div.listing-item",
		"div[class*='listing-card']",
	}

	for _, selector := range selectors {
		count := 0
		doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
			if i >= 10 { // Limit to first 10 for initial implementation
				return
			}

			cdp := s.extractListing(selection, baseURL)
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

	// If no listings found with specific selectors, try a general approach
	if len(dataPoints) == 0 {
		logger.Info("No listings found with specific selectors, trying general approach")
		dataPoints = s.extractWithGeneralApproach(doc, baseURL)
	}

	return dataPoints
}

// extractListing extracts a single listing from a goquery selection
func (scraper *DubizzleScraper) extractListing(sel *goquery.Selection, baseURL string) *models.CostDataPoint {
	// Extract price - try multiple selectors
	priceText := ""
	priceSelectors := []string{
		"[data-testid='listing-price']",
		"span[class*='price']",
		"span[class*='Price']",
		"div[class*='price']",
		"p[class*='price']",
	}

	for _, selector := range priceSelectors {
		priceText = sel.Find(selector).First().Text()
		if priceText != "" {
			break
		}
	}

	price := parsePrice(priceText)

	// Extract title/description
	title := ""
	titleSelectors := []string{
		"h2",
		"h3",
		"[data-testid='listing-title']",
		"a[title]",
		"div[class*='title']",
	}

	for _, selector := range titleSelectors {
		if selector == "a[title]" {
			title, _ = sel.Find(selector).First().Attr("title")
		} else {
			title = sel.Find(selector).First().Text()
		}
		if title != "" {
			break
		}
	}

	// Extract location
	locationText := ""
	locationSelectors := []string{
		"[data-testid='listing-location']",
		"span[class*='location']",
		"div[class*='location']",
		"p[class*='location']",
	}

	for _, selector := range locationSelectors {
		locationText = sel.Find(selector).First().Text()
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
		"[data-testid='bedrooms']",
		"span[class*='bed']",
		"div[class*='bed']",
	}

	for _, selector := range bedroomSelectors {
		bedrooms = sel.Find(selector).First().Text()
		if bedrooms != "" {
			break
		}
	}

	bathrooms := ""
	bathroomSelectors := []string{
		"[data-testid='bathrooms']",
		"span[class*='bath']",
		"div[class*='bath']",
	}

	for _, selector := range bathroomSelectors {
		bathrooms = sel.Find(selector).First().Text()
		if bathrooms != "" {
			break
		}
	}

	// Extract URL
	propertyURL := ""
	if href, exists := sel.Find("a").First().Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			propertyURL = href
		} else if strings.HasPrefix(href, "/") {
			propertyURL = "https://dubai.dubizzle.com" + href
		}
	}

	// Validate we have minimum required data
	if price == 0 || title == "" {
		return nil
	}

	now := time.Now()

	// Determine tags based on category
	tags := []string{"dubizzle", scraper.emirate}
	if scraper.category == "bedspace" || scraper.category == "roomspace" {
		tags = append(tags, "shared", "budget", scraper.category)
	} else {
		tags = append(tags, "rent", "apartment")
	}

	return &models.CostDataPoint{
		Category:    "Housing",
		SubCategory: scraper.getSubCategoryFromCategory(),
		ItemName:    strings.TrimSpace(title),
		Price:       price,
		Location:    location,
		Source:      "dubizzle",
		SourceURL:   propertyURL,
		Confidence:  0.75, // Slightly lower than Bayut due to potential bot detection issues
		Unit:        "AED",
		RecordedAt:  now,
		ValidFrom:   now,
		SampleSize:  1,
		Tags:        tags,
		Attributes: map[string]interface{}{
			"bedrooms":  strings.TrimSpace(bedrooms),
			"bathrooms": strings.TrimSpace(bathrooms),
		},
	}
}

// extractWithGeneralApproach tries to extract listings using a more general approach
func (s *DubizzleScraper) extractWithGeneralApproach(doc *goquery.Document, baseURL string) []*models.CostDataPoint {
	dataPoints := []*models.CostDataPoint{}

	// Look for any links that seem to be property listings
	doc.Find("a[href*='/property-for-rent/']").Each(func(i int, link *goquery.Selection) {
		if i >= 10 {
			return
		}

		// Get the parent container (likely the property card)
		card := link.Parent().Parent()

		// Try to extract price from the card
		priceText := card.Find("span, div, p").FilterFunction(func(_ int, sel *goquery.Selection) bool {
			text := sel.Text()
			return strings.Contains(text, "AED") || strings.Contains(text, "aed") ||
				   strings.Contains(text, "Dhs") || strings.Contains(text, "DHS")
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
		} else if strings.HasPrefix(href, "/") {
			propertyURL = "https://dubai.dubizzle.com" + href
		}

		now := time.Now()

		// Determine tags based on category
		tags := []string{"dubizzle", s.emirate}
		if s.category == "bedspace" || s.category == "roomspace" {
			tags = append(tags, "shared", "budget", s.category)
		} else {
			tags = append(tags, "rent", "apartment")
		}

		dataPoints = append(dataPoints, &models.CostDataPoint{
			Category:    "Housing",
			SubCategory: s.getSubCategoryFromCategory(),
			ItemName:    title,
			Price:       price,
			Location: models.Location{
				Emirate: s.emirate,
				City:    s.emirate,
			},
			Source:      "dubizzle",
			SourceURL:   propertyURL,
			Confidence:  0.5, // Lower confidence for general approach
			Unit:        "AED",
			RecordedAt:  now,
			ValidFrom:   now,
			SampleSize:  1,
			Tags:        tags,
			Attributes:  map[string]interface{}{},
		})
	})

	return dataPoints
}

// buildURL constructs the Dubizzle URL for the emirate and category
func (s *DubizzleScraper) buildURL() string {
	emiratePath := s.emirateToSubdomain(s.emirate)
	categoryPath := s.categoryToPath(s.category)
	return fmt.Sprintf("https://%s.dubizzle.com/property-for-rent/residential/%s/", emiratePath, categoryPath)
}

// emirateToSubdomain converts emirate name to Dubizzle subdomain format
func (s *DubizzleScraper) emirateToSubdomain(emirate string) string {
	// Dubizzle subdomain patterns:
	// Dubai -> dubai
	// Abu Dhabi -> abudhabi
	// Sharjah -> sharjah
	// Ajman -> ajman
	// Ras Al Khaimah -> rak
	// Fujairah -> fujairah
	// Umm Al Quwain -> uaq

	emirate = strings.ToLower(emirate)
	emirate = strings.ReplaceAll(emirate, " ", "")

	// Handle special cases
	if strings.Contains(emirate, "rasalkhaimah") || strings.Contains(emirate, "rak") {
		return "rak"
	}
	if strings.Contains(emirate, "ummalquwain") || strings.Contains(emirate, "uaq") {
		return "uaq"
	}

	return emirate
}

// categoryToPath converts category to Dubizzle URL path
func (s *DubizzleScraper) categoryToPath(category string) string {
	// Category patterns:
	// apartmentflat -> apartmentflat
	// bedspace -> bedspace
	// roomspace -> roomspace

	return category
}

// getSubCategoryFromCategory returns the subcategory based on the category
func (s *DubizzleScraper) getSubCategoryFromCategory() string {
	switch s.category {
	case "bedspace", "roomspace":
		return "Shared Accommodation"
	default:
		return "Rent"
	}
}
