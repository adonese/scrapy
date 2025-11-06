package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewConfigFromEnv(t *testing.T) {
	// Test with default values
	cfg := NewConfigFromEnv()

	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != "5432" {
		t.Errorf("expected port '5432', got '%s'", cfg.Port)
	}

	if cfg.DBName != "cost_of_living" {
		t.Errorf("expected dbname 'cost_of_living', got '%s'", cfg.DBName)
	}
}

func TestNewConfigFromEnvWithCustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")

	defer func() {
		// Clean up
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg := NewConfigFromEnv()

	if cfg.Host != "testhost" {
		t.Errorf("expected host 'testhost', got '%s'", cfg.Host)
	}

	if cfg.Port != "5433" {
		t.Errorf("expected port '5433', got '%s'", cfg.Port)
	}

	if cfg.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", cfg.User)
	}

	if cfg.Password != "testpass" {
		t.Errorf("expected password 'testpass', got '%s'", cfg.Password)
	}

	if cfg.DBName != "testdb" {
		t.Errorf("expected dbname 'testdb', got '%s'", cfg.DBName)
	}

	if cfg.SSLMode != "require" {
		t.Errorf("expected sslmode 'require', got '%s'", cfg.SSLMode)
	}
}

// TestConnectIntegration tests actual database connection
// This test requires a running PostgreSQL instance
// Run with: go test -tags=integration ./pkg/database
func TestConnectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := NewConfigFromEnv()
	db, err := Connect(cfg)
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
		return
	}
	defer db.Close()

	// Test Ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// Test HealthCheck
	if err := db.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test GetConn
	if conn := db.GetConn(); conn == nil {
		t.Error("GetConn returned nil")
	}
}

func TestDatabaseNilConnection(t *testing.T) {
	db := &DB{conn: nil}

	ctx := context.Background()

	// Test Ping with nil connection
	if err := db.Ping(ctx); err == nil {
		t.Error("expected error when pinging nil connection")
	}

	// Test HealthCheck with nil connection
	if err := db.HealthCheck(ctx); err == nil {
		t.Error("expected error when health checking nil connection")
	}

	// Test Close with nil connection (should not panic)
	if err := db.Close(); err != nil {
		t.Errorf("Close should not error on nil connection, got: %v", err)
	}
}
