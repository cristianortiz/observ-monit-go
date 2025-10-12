package health

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	serviceName := "test-service"
	version := "1.0.0"

	health := New(serviceName, version)

	if health == nil {
		t.Fatal("New() returned nil")
	}

	if health.serviceName != serviceName {
		t.Errorf("Expected service name %s, got %s", serviceName, health.serviceName)
	}

	if health.version != version {
		t.Errorf("Expected version %s, got %s", version, health.version)
	}
}

func TestCheck(t *testing.T) {
	health := New("test-service", "1.0.0")

	response := health.Check()

	if response.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, response.Status)
	}

	if response.Service != "test-service" {
		t.Errorf("Expected service test-service, got %s", response.Service)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response.Version)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	// check recent timestamp
	if time.Since(response.Timestamp) > time.Second {
		t.Error("Timestamp is too old")
	}
}

func TestCheckLiveness(t *testing.T) {
	health := New("test-service", "1.0.0")

	response := health.CheckLiveness()

	if response.Status != StatusHealthy {
		t.Errorf("Expected healthy status, got %s", response.Status)
	}

	if response.Service != "test-service" {
		t.Errorf("Expected service test-service, got %s", response.Service)
	}
}

func TestCheckReadiness(t *testing.T) {
	health := New("test-service", "1.0.0")

	response := health.CheckReadiness()

	if response.Status != StatusHealthy {
		t.Errorf("Expected healthy status, got %s", response.Status)
	}

	if response.Service != "test-service" {
		t.Errorf("Expected service test-service, got %s", response.Service)
	}
}

func TestMultipleChecks(t *testing.T) {
	health := New("test-service", "1.0.0")

	// Ejecutar múltiples checks
	for i := 0; i < 5; i++ {
		response := health.Check()

		if response.Status != StatusHealthy {
			t.Errorf("Check %d: Expected healthy status, got %s", i, response.Status)
		}
	}
}
