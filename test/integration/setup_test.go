package integration

import (
	"os"
	"testing"

	"github.com/adonese/cost-of-living/pkg/logger"
)

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.Init()

	// Run tests
	exitCode := m.Run()

	// Cleanup if needed

	os.Exit(exitCode)
}
