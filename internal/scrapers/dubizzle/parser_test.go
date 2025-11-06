package dubizzle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Price with AED and comma",
			input:    "AED 85,000",
			expected: 85000,
		},
		{
			name:     "Price with Dhs",
			input:    "75,000 Dhs",
			expected: 75000,
		},
		{
			name:     "Price with DHS uppercase",
			input:    "120,000 DHS",
			expected: 120000,
		},
		{
			name:     "Price per year",
			input:    "AED 95,000/year",
			expected: 95000,
		},
		{
			name:     "Price per month",
			input:    "7,500 AED/month",
			expected: 7500,
		},
		{
			name:     "Price without comma",
			input:    "AED 50000",
			expected: 50000,
		},
		{
			name:     "Price with yearly",
			input:    "100,000 AED/yearly",
			expected: 100000,
		},
		{
			name:     "Price with dirhams",
			input:    "85,000 Dirhams",
			expected: 85000,
		},
		{
			name:     "Invalid text",
			input:    "Contact for price",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Large price",
			input:    "AED 2,500,000",
			expected: 2500000,
		},
		{
			name:     "Price with decimal",
			input:    "85,500 AED",
			expected: 85500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrice(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedEmirate string
		expectedCity    string
		expectedArea    string
	}{
		{
			name:            "Location with comma",
			input:           "Dubai Marina, Dubai",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Dubai Marina",
		},
		{
			name:            "Location with hyphen",
			input:           "Business Bay - Dubai",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Business Bay",
		},
		{
			name:            "Location with pipe",
			input:           "Downtown Dubai | Dubai",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Downtown Dubai",
		},
		{
			name:            "Area only",
			input:           "Jumeirah Lake Towers",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Jumeirah Lake Towers",
		},
		{
			name:            "Sharjah location",
			input:           "Al Nahda, Sharjah",
			expectedEmirate: "Sharjah",
			expectedCity:    "Sharjah",
			expectedArea:    "Al Nahda",
		},
		{
			name:            "Abu Dhabi location",
			input:           "Al Reem Island, Abu Dhabi",
			expectedEmirate: "Abu Dhabi",
			expectedCity:    "Abu Dhabi",
			expectedArea:    "Al Reem Island",
		},
		{
			name:            "Ajman location",
			input:           "Al Rashidiya - Ajman",
			expectedEmirate: "Ajman",
			expectedCity:    "Ajman",
			expectedArea:    "Al Rashidiya",
		},
		{
			name:            "RAK abbreviation",
			input:           "Al Hamra, RAK",
			expectedEmirate: "RAK",
			expectedCity:    "RAK",
			expectedArea:    "Al Hamra",
		},
		{
			name:            "Empty string",
			input:           "",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "",
		},
		{
			name:            "Emirate only",
			input:           "Fujairah",
			expectedEmirate: "Fujairah",
			expectedCity:    "Fujairah",
			expectedArea:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLocation(tt.input)
			assert.Equal(t, tt.expectedEmirate, result.Emirate, "Emirate mismatch for input: %s", tt.input)
			assert.Equal(t, tt.expectedCity, result.City, "City mismatch for input: %s", tt.input)
			assert.Equal(t, tt.expectedArea, result.Area, "Area mismatch for input: %s", tt.input)
		})
	}
}

func TestIsEmirate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Dubai",
			input:    "Dubai",
			expected: true,
		},
		{
			name:     "Abu Dhabi",
			input:    "Abu Dhabi",
			expected: true,
		},
		{
			name:     "Sharjah lowercase",
			input:    "sharjah",
			expected: true,
		},
		{
			name:     "Ajman",
			input:    "Ajman",
			expected: true,
		},
		{
			name:     "RAK abbreviation",
			input:    "RAK",
			expected: true,
		},
		{
			name:     "Ras Al Khaimah full",
			input:    "Ras Al Khaimah",
			expected: true,
		},
		{
			name:     "Fujairah",
			input:    "Fujairah",
			expected: true,
		},
		{
			name:     "UAQ abbreviation",
			input:    "UAQ",
			expected: true,
		},
		{
			name:     "Umm Al Quwain full",
			input:    "Umm Al Quwain",
			expected: true,
		},
		{
			name:     "Not an emirate",
			input:    "New York",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Random area",
			input:    "Marina",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmirate(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestParseBedrooms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "One bedroom",
			input:    "1 Bedroom",
			expected: "1",
		},
		{
			name:     "Two bedrooms with BR",
			input:    "2 BR",
			expected: "2",
		},
		{
			name:     "Three bedrooms",
			input:    "3 Bed",
			expected: "3",
		},
		{
			name:     "Studio",
			input:    "Studio",
			expected: "Studio",
		},
		{
			name:     "Studio lowercase",
			input:    "studio apartment",
			expected: "Studio",
		},
		{
			name:     "Studio with extra text",
			input:    "Spacious studio",
			expected: "Studio",
		},
		{
			name:     "Five bedrooms",
			input:    "5BR",
			expected: "5",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No number",
			input:    "Bedroom",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBedrooms(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestParseBathrooms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "One bathroom",
			input:    "1 Bathroom",
			expected: "1",
		},
		{
			name:     "Two bathrooms",
			input:    "2 Bath",
			expected: "2",
		},
		{
			name:     "Three bathrooms",
			input:    "3 Baths",
			expected: "3",
		},
		{
			name:     "Bathroom with number only",
			input:    "2",
			expected: "2",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No number",
			input:    "Bathroom",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBathrooms(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestParseArea(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Area in sqft",
			input:    "1200 sqft",
			expected: 1200,
		},
		{
			name:     "Area with comma",
			input:    "1,500 sq.ft",
			expected: 1500,
		},
		{
			name:     "Area in sqm",
			input:    "120 sqm",
			expected: 120,
		},
		{
			name:     "Area with period",
			input:    "2,000 sq. ft.",
			expected: 2000,
		},
		{
			name:     "Area without unit",
			input:    "800",
			expected: 800,
		},
		{
			name:     "Large area",
			input:    "5,000 sq ft",
			expected: 5000,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "No number",
			input:    "sqft",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArea(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Title with extra spaces",
			input:    "Spacious  1BR   Apartment",
			expected: "Spacious 1BR Apartment",
		},
		{
			name:     "Title with leading/trailing spaces",
			input:    "  Luxury 2BR Villa  ",
			expected: "Luxury 2BR Villa",
		},
		{
			name:     "Normal title",
			input:    "Studio in Marina",
			expected: "Studio in Marina",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Title with tabs and newlines",
			input:    "Modern\t3BR\nApartment",
			expected: "Modern 3BR Apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTitle(tt.input)
			assert.Equal(t, tt.expected, result, "Failed for input: %s", tt.input)
		})
	}
}
