package helpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockServer represents a test HTTP server that serves fixture files
type MockServer struct {
	Server   *httptest.Server
	fixtures map[string]string // path -> fixture content
}

// NewMockServer creates a new mock HTTP server
func NewMockServer() *MockServer {
	ms := &MockServer{
		fixtures: make(map[string]string),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handleRequest))
	return ms
}

// AddFixture adds a fixture to be served at a specific path
// path: URL path (e.g., "/to-rent/apartments/dubai/")
// fixtureDir: fixture directory (e.g., "bayut")
// fixtureFile: fixture filename (e.g., "dubai_listings.html")
func (ms *MockServer) AddFixture(path, fixtureDir, fixtureFile string) error {
	content, err := LoadFixture(fixtureDir, fixtureFile)
	if err != nil {
		return err
	}

	ms.fixtures[path] = content
	return nil
}

// AddFixtureContent adds raw HTML content to be served at a specific path
func (ms *MockServer) AddFixtureContent(path, content string) {
	ms.fixtures[path] = content
}

// handleRequest handles incoming HTTP requests and serves appropriate fixtures
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Look for exact match first
	if content, ok := ms.fixtures[path]; ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
		return
	}

	// Try prefix matching (useful for dynamic URLs)
	for fixturePath, content := range ms.fixtures {
		if strings.HasPrefix(path, fixturePath) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
			return
		}
	}

	// No fixture found, return 404
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Mock server: fixture not found for path: " + path))
}

// Close shuts down the mock server
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// URL returns the base URL of the mock server
func (ms *MockServer) URL() string {
	return ms.Server.URL
}

// MockServerBuilder helps build a mock server with fluent API
type MockServerBuilder struct {
	server *MockServer
	err    error
}

// NewMockServerBuilder creates a new builder
func NewMockServerBuilder() *MockServerBuilder {
	return &MockServerBuilder{
		server: NewMockServer(),
	}
}

// WithFixture adds a fixture to the server
func (b *MockServerBuilder) WithFixture(path, fixtureDir, fixtureFile string) *MockServerBuilder {
	if b.err != nil {
		return b
	}

	err := b.server.AddFixture(path, fixtureDir, fixtureFile)
	if err != nil {
		b.err = err
	}

	return b
}

// WithContent adds raw content to the server
func (b *MockServerBuilder) WithContent(path, content string) *MockServerBuilder {
	b.server.AddFixtureContent(path, content)
	return b
}

// Build returns the configured server or an error
func (b *MockServerBuilder) Build() (*MockServer, error) {
	if b.err != nil {
		b.server.Close()
		return nil, b.err
	}
	return b.server, nil
}

// MustBuild returns the configured server or panics
func (b *MockServerBuilder) MustBuild() *MockServer {
	server, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build mock server: %v", err))
	}
	return server
}

// NewBayutMockServer creates a pre-configured mock server for Bayut testing
func NewBayutMockServer() (*MockServer, error) {
	return NewMockServerBuilder().
		WithFixture("/to-rent/apartments/dubai/", "bayut", "dubai_listings.html").
		WithFixture("/to-rent/apartments/sharjah/", "bayut", "sharjah_listings.html").
		WithFixture("/to-rent/apartments/ajman/", "bayut", "ajman_listings.html").
		WithFixture("/to-rent/apartments/abu-dhabi/", "bayut", "abudhabi_listings.html").
		WithFixture("/to-rent/apartments/empty/", "bayut", "empty_results.html").
		Build()
}

// NewDubizzleMockServer creates a pre-configured mock server for Dubizzle testing
func NewDubizzleMockServer() (*MockServer, error) {
	return NewMockServerBuilder().
		WithFixture("/property-for-rent/residential/apartmentflat/", "dubizzle", "apartments.html").
		WithFixture("/property-for-rent/residential/bedspace/", "dubizzle", "bedspace.html").
		WithFixture("/property-for-rent/residential/roomspace/", "dubizzle", "roomspace.html").
		WithFixture("/blocked/", "dubizzle", "error_page.html").
		Build()
}
