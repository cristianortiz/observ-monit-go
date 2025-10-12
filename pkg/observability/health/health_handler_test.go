package health

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/cristianortiz/observ-monit-go/pkg/observability/logger"
	"github.com/gofiber/fiber/v2"
)

func setupTestApp(t *testing.T) (*fiber.App, *Handler) {
	// tests logger
	log, err := logger.New("debug", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// health checker
	health := New("test-service", "1.0.0")

	// handler
	handler := NewHandler(health, log)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	handler.RegisterRoutes(app, "/health", "/ready")

	return app, handler
}

func TestHandleHealth(t *testing.T) {
	app, _ := setupTestApp(t)

	// Creates request for test
	req := httptest.NewRequest("GET", "/health", nil)

	// make request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// check Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
	//read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// check response fields from health pkg
	if response.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, response.Status)
	}

	if response.Service != "test-service" {
		t.Errorf("Expected service test-service, got %s", response.Service)
	}
}

func TestHandleReady(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest("GET", "/ready", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// check response fields from health
	if response.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, response.Status)
	}
}

func TestMultipleHealthRequests(t *testing.T) {
	app, _ := setupTestApp(t)

	// make multiple requests
	for i := range 10 {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i, resp.StatusCode)
		}

		resp.Body.Close()
	}
}

func TestHealthAndReadyEndpoints(t *testing.T) {
	app, _ := setupTestApp(t)

	tests := []struct {
		name     string
		endpoint string
		wantCode int
	}{
		{
			name:     "health endpoint",
			endpoint: "/health",
			wantCode: fiber.StatusOK,
		},
		{
			name:     "ready endpoint",
			endpoint: "/ready",
			wantCode: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, resp.StatusCode)
			}
		})
	}
}
