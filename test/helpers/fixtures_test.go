package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFixture(t *testing.T) {
	// Test loading a Bayut fixture
	content, err := LoadFixture("bayut", "dubai_listings.html")
	require.NoError(t, err, "Should load Bayut Dubai fixture")
	assert.NotEmpty(t, content, "Fixture content should not be empty")
	assert.Contains(t, content, "Dubai Marina", "Fixture should contain expected content")

	// Test loading a Dubizzle fixture
	content, err = LoadFixture("dubizzle", "apartments.html")
	require.NoError(t, err, "Should load Dubizzle apartments fixture")
	assert.NotEmpty(t, content, "Fixture content should not be empty")
	assert.Contains(t, content, "property-for-rent", "Fixture should contain expected content")

	// Test non-existent fixture
	_, err = LoadFixture("nonexistent", "file.html")
	assert.Error(t, err, "Should return error for non-existent fixture")
}

func TestFixtureExists(t *testing.T) {
	assert.True(t, FixtureExists("bayut", "dubai_listings.html"), "Dubai fixture should exist")
	assert.True(t, FixtureExists("dubizzle", "apartments.html"), "Apartments fixture should exist")
	assert.False(t, FixtureExists("bayut", "nonexistent.html"), "Non-existent fixture should not exist")
}

func TestListFixtures(t *testing.T) {
	// List Bayut fixtures
	bayutFixtures, err := ListFixtures("bayut")
	require.NoError(t, err, "Should list Bayut fixtures")
	assert.GreaterOrEqual(t, len(bayutFixtures), 4, "Should have at least 4 Bayut fixtures")
	assert.Contains(t, bayutFixtures, "dubai_listings.html", "Should contain Dubai listings")
	assert.Contains(t, bayutFixtures, "empty_results.html", "Should contain empty results")

	// List Dubizzle fixtures
	dubizzleFixtures, err := ListFixtures("dubizzle")
	require.NoError(t, err, "Should list Dubizzle fixtures")
	assert.GreaterOrEqual(t, len(dubizzleFixtures), 3, "Should have at least 3 Dubizzle fixtures")
	assert.Contains(t, dubizzleFixtures, "apartments.html", "Should contain apartments")
	assert.Contains(t, dubizzleFixtures, "bedspace.html", "Should contain bedspace")
	assert.Contains(t, dubizzleFixtures, "roomspace.html", "Should contain roomspace")
}

func TestMustLoadFixture(t *testing.T) {
	// Should not panic for valid fixture
	assert.NotPanics(t, func() {
		content := MustLoadFixture("bayut", "dubai_listings.html")
		assert.NotEmpty(t, content)
	}, "Should not panic for valid fixture")

	// Should panic for invalid fixture
	assert.Panics(t, func() {
		MustLoadFixture("invalid", "nonexistent.html")
	}, "Should panic for invalid fixture")
}
