package bayut

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
			name:     "Price with comma and year",
			input:    "AED 85,000/year",
			expected: 85000,
		},
		{
			name:     "Price without comma",
			input:    "AED 120000",
			expected: 120000,
		},
		{
			name:     "Price with comma and trailing AED",
			input:    "95,500 AED",
			expected: 95500,
		},
		{
			name:     "Price per month",
			input:    "AED 7,500 per month",
			expected: 7500,
		},
		{
			name:     "Invalid text",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Large price",
			input:    "AED 1,500,000/year",
			expected: 1500000,
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
			name:            "Full location with comma",
			input:           "Dubai Marina, Dubai",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Dubai Marina",
		},
		{
			name:            "Area only",
			input:           "Business Bay",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Business Bay",
		},
		{
			name:            "Location with emirate",
			input:           "Al Nahda, Sharjah",
			expectedEmirate: "Sharjah",
			expectedCity:    "Sharjah",
			expectedArea:    "Al Nahda",
		},
		{
			name:            "Empty string",
			input:           "",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "",
		},
		{
			name:            "Location with hyphen",
			input:           "Downtown Dubai - Dubai",
			expectedEmirate: "Dubai",
			expectedCity:    "Dubai",
			expectedArea:    "Downtown Dubai",
		},
		{
			name:            "Abu Dhabi location",
			input:           "Al Reem Island, Abu Dhabi",
			expectedEmirate: "Abu Dhabi",
			expectedCity:    "Abu Dhabi",
			expectedArea:    "Al Reem Island",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLocation(tt.input)
			assert.Equal(t, tt.expectedEmirate, result.Emirate, "Emirate mismatch")
			assert.Equal(t, tt.expectedCity, result.City, "City mismatch")
			assert.Equal(t, tt.expectedArea, result.Area, "Area mismatch")
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
			name:     "Abu Dhabi with mixed case",
			input:    "Abu Dhabi",
			expected: true,
		},
		{
			name:     "Sharjah lowercase",
			input:    "sharjah",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmirate(tt.input)
			assert.Equal(t, tt.expected, result)
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
			name:     "Three bedrooms",
			input:    "3 BR",
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
			assert.Equal(t, tt.expected, result)
		})
	}
}
