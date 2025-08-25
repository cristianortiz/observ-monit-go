package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("USERS_SERVICE_PORT", "8081")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("USERS_SERVICE_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("LOG_LEVEL")
	}()

	config, err := Load("users-service")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test environment variables work
	if config.Service.Port != 8081 {
		t.Errorf("Expected port 8081, got %d", config.Service.Port)
	}

	if config.Database.Host != "testhost" {
		t.Errorf("Expected DB host 'testhost', got '%s'", config.Database.Host)
	}

	if config.Observability.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Observability.LogLevel)
	}
}

func TestDefaults(t *testing.T) {
	// No environment variables set
	config, err := Load("test-service")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test defaults
	if config.Service.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Service.Port)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected default DB host 'localhost', got '%s'", config.Database.Host)
	}
}
