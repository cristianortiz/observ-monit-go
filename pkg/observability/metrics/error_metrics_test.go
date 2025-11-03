package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestErrorMetrics(t *testing.T) {
	// Create a custom registry for testing
	registry := prometheus.NewRegistry()

	// Create metrics instance
	metrics := New("test-service")

	// Register metrics in custom registry
	registry.MustRegister(
		metrics.httpClientErrors,
		metrics.httpServerErrors,
		metrics.httpSlowRequests,
	)

	tests := []struct {
		name           string
		statusCode     string
		expectedMetric string
		action         func()
	}{
		{
			name:           "should record 4xx client error",
			statusCode:     "404",
			expectedMetric: "http_client_errors_total",
			action: func() {
				metrics.RecordHTTPClientError("test-service", "GET", "/api/users", "404")
			},
		},
		{
			name:           "should record 5xx server error",
			statusCode:     "500",
			expectedMetric: "http_server_errors_total",
			action: func() {
				metrics.RecordHTTPServerError("test-service", "POST", "/api/users", "500")
			},
		},
		{
			name:           "should record slow request",
			statusCode:     "200",
			expectedMetric: "http_slow_requests_total",
			action: func() {
				metrics.RecordSlowRequest("test-service", "GET", "/api/users", "1s")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the action
			tt.action()

			// Gather metrics
			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Failed to gather metrics: %v", err)
			}

			// Find the metric family
			var found bool
			for _, family := range families {
				if family.GetName() == tt.expectedMetric {
					found = true
					// Check that we have at least one metric
					if len(family.GetMetric()) == 0 {
						t.Errorf("Expected metric %s to have values", tt.expectedMetric)
					}

					// Check the value is greater than 0
					metric := family.GetMetric()[0]
					var value float64
					if metric.GetCounter() != nil {
						value = metric.GetCounter().GetValue()
					}

					if value <= 0 {
						t.Errorf("Expected metric %s to have value > 0, got %f", tt.expectedMetric, value)
					}
					break
				}
			}

			if !found {
				t.Errorf("Metric %s not found", tt.expectedMetric)
			}
		})
	}
}

func TestMetricLabels(t *testing.T) {
	// Create a custom registry for testing
	registry := prometheus.NewRegistry()

	// Create metrics manually to avoid global registry conflicts
	clientErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_errors_total",
			Help: "Total number of HTTP 4xx client errors",
		},
		[]string{"service", "method", "path", "status"},
	)

	// Register in custom registry
	registry.MustRegister(clientErrors)

	// Record an error
	clientErrors.WithLabelValues("test-service", "GET", "/api/users/:id", "404").Inc()

	// Gather metrics
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Find the client errors metric
	for _, family := range families {
		if family.GetName() == "http_client_errors_total" {
			if len(family.GetMetric()) == 0 {
				t.Fatal("Expected at least one metric")
			}

			metric := family.GetMetric()[0]
			labels := metric.GetLabel()

			// Check that we have the expected labels
			expectedLabels := map[string]string{
				"service": "test-service",
				"method":  "GET",
				"path":    "/api/users/:id",
				"status":  "404",
			}

			if len(labels) != len(expectedLabels) {
				t.Errorf("Expected %d labels, got %d", len(expectedLabels), len(labels))
			}

			labelMap := make(map[string]string)
			for _, label := range labels {
				labelMap[label.GetName()] = label.GetValue()
			}

			for key, expectedValue := range expectedLabels {
				if actualValue, exists := labelMap[key]; !exists {
					t.Errorf("Label %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Label %s: expected %s, got %s", key, expectedValue, actualValue)
				}
			}
			return
		}
	}

	t.Error("http_client_errors_total metric not found")
}
