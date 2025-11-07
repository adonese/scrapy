package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetFixturePath returns the absolute path to a fixture file
// Usage: GetFixturePath("bayut", "dubai_listings.html")
func GetFixturePath(directory, filename string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}

	// Get the directory of this helpers file
	helpersDir := filepath.Dir(file)

	// Go up one level to test/, then into fixtures/
	fixturesDir := filepath.Join(helpersDir, "..", "fixtures", directory)

	return filepath.Join(fixturesDir, filename)
}

// LoadFixture reads and returns the contents of a fixture file
// Usage: html, err := LoadFixture("bayut", "dubai_listings.html")
func LoadFixture(directory, filename string) (string, error) {
	path := GetFixturePath(directory, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read fixture %s/%s: %w", directory, filename, err)
	}

	return string(data), nil
}

// MustLoadFixture loads a fixture or panics if it fails
// Useful for test setup where failure should stop execution
func MustLoadFixture(directory, filename string) string {
	content, err := LoadFixture(directory, filename)
	if err != nil {
		panic(fmt.Sprintf("failed to load fixture: %v", err))
	}
	return content
}

// FixtureExists checks if a fixture file exists
func FixtureExists(directory, filename string) bool {
	path := GetFixturePath(directory, filename)
	_, err := os.Stat(path)
	return err == nil
}

// ListFixtures returns all fixture files in a directory
func ListFixtures(directory string) ([]string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get caller information")
	}

	helpersDir := filepath.Dir(file)
	fixturesDir := filepath.Join(helpersDir, "..", "fixtures", directory)

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixtures directory %s: %w", directory, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// GetProjectRoot returns the project root directory
func GetProjectRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}

	// From test/helpers/ go up two levels to project root
	helpersDir := filepath.Dir(file)
	return filepath.Join(helpersDir, "..", "..")
}
