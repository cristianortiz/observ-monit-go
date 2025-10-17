package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockDatabase implements DatabaseChecker for testing
type MockDatabase struct {
	PingError error
	PingDelay time.Duration
	Stats     map[string]interface{}
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	if m.PingDelay > 0 {
		time.Sleep(m.PingDelay)
	}
	return m.PingError
}

func (m *MockDatabase) GetPoolStats() map[string]interface{} {
	if m.Stats == nil {
		return map[string]interface{}{
			"total_conns": 5,
			"idle_conns":  3,
		}
	}
	return m.Stats
}
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

	// exec multiple test
	for i := range 5 {
		response := health.Check()

		if response.Status != StatusHealthy {
			t.Errorf("Check %d: Expected healthy status, got %s", i, response.Status)
		}
	}
}

func TestCheckReadiness_NoDatabase(t *testing.T) {
	h := New("test-service", "1.0.0")

	response := h.CheckReadiness()

	if response.Status != StatusHealthy {
		t.Errorf("Expected healthy status when no database, got %s", response.Status)
	}

	if response.Checks != nil {
		t.Error("Expected no checks when database not configured")
	}
}

func TestCheckReadiness_DatabaseHealthy(t *testing.T) {
	h := New("test-service", "1.0.0")

	mockDB := &MockDatabase{
		PingError: nil,
		PingDelay: 50 * time.Millisecond,
	}

	h.SetDatabase(mockDB)

	response := h.CheckReadiness()

	if response.Status != StatusHealthy {
		t.Errorf("Expected healthy status, got %s", response.Status)
	}

	if len(response.Checks) != 1 {
		t.Error("Expected 1 check result")
	}

	dbCheck, ok := response.Checks["database"]
	if !ok {
		t.Fatal("Expected database check in response")
	}

	if dbCheck.Status != StatusHealthy {
		t.Errorf("Expected database healthy, got %s", dbCheck.Status)
	}
}

func TestCheckReadiness_DatabaseUnhealthy(t *testing.T) {
	h := New("test-service", "1.0.0")

	mockDB := &MockDatabase{
		PingError: errors.New("connection refused"),
	}

	h.SetDatabase(mockDB)

	response := h.CheckReadiness()

	if response.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", response.Status)
	}

	dbCheck := response.Checks["database"]
	if dbCheck.Status != StatusUnhealthy {
		t.Errorf("Expected database unhealthy, got %s", dbCheck.Status)
	}

	if dbCheck.Error == "" {
		t.Error("Expected error message in database check")
	}
}

func TestCheckReadiness_DatabaseSlow(t *testing.T) {
	h := New("test-service", "1.0.0")

	mockDB := &MockDatabase{
		PingError: nil,
		PingDelay: 1500 * time.Millisecond, // More than 1 sec
	}

	h.SetDatabase(mockDB)

	response := h.CheckReadiness()

	if response.Status != StatusDegraded {
		t.Errorf("Expected degraded status, got %s", response.Status)
	}

	dbCheck := response.Checks["database"]
	if dbCheck.Status != StatusDegraded {
		t.Errorf("Expected database degraded, got %s", dbCheck.Status)
	}
}
