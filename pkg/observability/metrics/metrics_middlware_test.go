package metrics

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMiddleware(t *testing.T) {

	// 1. creates test metrics, NOTE: use registry custom to avoid conflicts between tests
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "HTTP request size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		httpResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
	}

	// Register metrics in the registry custom
	registry.MustRegister(
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
		metrics.httpRequestSize,
		metrics.httpResponseSize,
		metrics.activeConnections,
	)

	//2. creates app Fiber with the middleware
	app := fiber.New()
	app.Use(Middleware(MetricsConfig{
		ServiceName: "test-service",
		Metrics:     metrics,
	}))

	// 3. Create test EP
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 4. make test request
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// 5. check response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 6. check the counter already increments
	counterValue := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/test", "200"))
	if counterValue != 1 {
		t.Errorf("Expected counter value 1, got %f", counterValue)
	}

	// 7. check the gauge return to 0 (conn closed)
	gaugeValue := testutil.ToFloat64(metrics.activeConnections)
	if gaugeValue != 0 {
		t.Errorf("Expected active connections 0, got %f", gaugeValue)
	}
}

func TestMiddlewareMultipleRequests(t *testing.T) {
	// Test to check  multiples requests
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "HTTP request size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		httpResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
	}

	registry.MustRegister(
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
		metrics.httpRequestSize,
		metrics.httpResponseSize,
		metrics.activeConnections,
	)

	app := fiber.New()
	app.Use(Middleware(MetricsConfig{
		ServiceName: "test-service",
		Metrics:     metrics,
	}))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// make 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}

	// check the counter at 5
	counterValue := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/test", "200"))
	if counterValue != 5 {
		t.Errorf("Expected counter value 5, got %f", counterValue)
	}

	// check there is not active connections
	gaugeValue := testutil.ToFloat64(metrics.activeConnections)
	if gaugeValue != 0 {
		t.Errorf("Expected active connections 0, got %f", gaugeValue)
	}
}

func TestMiddlewareWithDifferentStatusCodes(t *testing.T) {
	// Test to check differents status codes
	registry := prometheus.NewRegistry()
	metrics := &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "HTTP request size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		httpResponseSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response size in bytes",
			},
			[]string{"service", "method", "path"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
		// Agregar las nuevas métricas de errores
		httpClientErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_client_errors_total",
				Help: "Total number of HTTP 4xx client errors",
			},
			[]string{"service", "method", "path", "status"},
		),
		httpServerErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_server_errors_total",
				Help: "Total number of HTTP 5xx server errors",
			},
			[]string{"service", "method", "path", "status"},
		),
		httpSlowRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_slow_requests_total",
				Help: "Total number of HTTP requests exceeding SLO threshold",
			},
			[]string{"service", "method", "path", "threshold"},
		),
	}

	registry.MustRegister(
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
		metrics.httpRequestSize,
		metrics.httpResponseSize,
		metrics.activeConnections,
		metrics.httpClientErrors,
		metrics.httpServerErrors,
		metrics.httpSlowRequests,
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(Middleware(MetricsConfig{
		ServiceName: "test-service",
		Metrics:     metrics,
	}))

	// Endpoint that returns 200
	app.Get("/success", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Endpoint that returns 404
	app.Get("/not-found", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("Not Found")
	})

	// Endpoint that returns 500
	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("Internal Error")
	})

	// Make requests to each endpoint
	tests := []struct {
		path           string
		expectedStatus int
	}{
		{"/success", 200},
		{"/not-found", 404},
		{"/error", 500},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		resp, err := app.Test(req, -1) // -1 disables timeout and forces fresh context
		if err != nil {
			t.Fatalf("Request to %s failed: %v", tt.path, err)
		}

		if resp.StatusCode != tt.expectedStatus {
			t.Errorf("Expected status %d for %s, got %d", tt.expectedStatus, tt.path, resp.StatusCode)
		}

		resp.Body.Close()
	}

	// Verify metrics by status code
	success := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/success", "200"))
	notFound := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/not-found", "404"))
	errorCount := testutil.ToFloat64(metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/error", "500"))

	// Debug: Print all metrics to see what labels are actually registered
	if notFound != 1 {
		t.Logf("Debugging: checking all registered metrics")
		metricFamilies, err := registry.Gather()
		if err == nil {
			for _, mf := range metricFamilies {
				if mf.GetName() == "http_requests_total" {
					t.Logf("Found http_requests_total metrics:")
					for _, m := range mf.GetMetric() {
						labels := m.GetLabel()
						labelStr := ""
						for _, l := range labels {
							labelStr += l.GetName() + "=" + l.GetValue() + " "
						}
						t.Logf("  Labels: %s, Value: %f", labelStr, m.GetCounter().GetValue())
					}
				}
			}
		}
	}

	if success != 1 {
		t.Errorf("Expected 1 success request, got %f", success)
	}
	if notFound != 1 {
		t.Errorf("Expected 1 not found request, got %f", notFound)
	}
	if errorCount != 1 {
		t.Errorf("Expected 1 error request, got %f", errorCount)
	}
}
