package metrics

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHandleMetrics(t *testing.T) {
	// 1. creates metrics for testing purpuses only
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

	// 2. Register some metrics for testing
	metrics.RecordHTTPRequest("test-service", "GET", "/api/users", "200")
	metrics.RecordHTTPDuration("test-service", "GET", "/api/users", "200", 0.125)
	metrics.IncActiveConnections()

	// 3. Creates handler
	handler := newHandlerWithGatherer(metrics, registry)
	// 4. fiber app and register routes
	app := fiber.New()
	handler.RegisterRoutes(app)

	// 5. request metrics ep
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// 6. check status code
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 7. check Content-Type
	contentType := resp.Header.Get("Content-Type")
	//prometheus requires test/plain
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to contain 'text/plain', got '%s'", contentType)
	}

	// 8. body read
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// 9. check for expected metrics
	expectedMetrics := []string{
		"# HELP http_requests_total",
		"# TYPE http_requests_total counter",
		"http_requests_total",
		"# HELP http_request_duration_seconds",
		"# TYPE http_request_duration_seconds histogram",
		"http_request_duration_seconds",
		"# HELP http_active_connections",
		"# TYPE http_active_connections gauge",
		"http_active_connections",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Expected response to contain '%s', but it didn't", expected)
		}
	}
}

func TestHandleMetricsFormat(t *testing.T) {
	// Test más específico del formato
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

	// Registrar métricas específicas
	metrics.RecordHTTPRequest("users-service", "POST", "/api/users", "201")
	metrics.RecordHTTPRequest("users-service", "GET", "/api/users/:id", "200")
	metrics.RecordHTTPRequest("users-service", "GET", "/api/users/:id", "404")

	handler := newHandlerWithGatherer(metrics, registry)
	app := fiber.New()
	handler.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Verificar formato específico de Prometheus
	tests := []struct {
		description string
		contains    string
	}{
		{
			description: "Counter con labels específicos",
			contains:    `http_requests_total{method="POST",path="/api/users",service="users-service",status="201"}`,
		},
		{
			description: "Counter con status 200",
			contains:    `http_requests_total{method="GET",path="/api/users/:id",service="users-service",status="200"}`,
		},
		{
			description: "Counter con status 404",
			contains:    `http_requests_total{method="GET",path="/api/users/:id",service="users-service",status="404"}`,
		},
	}

	for _, tt := range tests {
		if !strings.Contains(bodyStr, tt.contains) {
			t.Errorf("%s: Expected to find '%s' in response", tt.description, tt.contains)
		}
	}
}

func TestRegisterRoutes(t *testing.T) {
	// Test que verifica que RegisterRoutes configura la ruta correctamente
	// Usar un registro personalizado para evitar conflictos
	registry := prometheus.NewRegistry()

	// Crear métricas manualmente con el registro personalizado
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)
	registry.MustRegister(counter)

	// Crear un struct Metrics vacío (solo para cumplir la firma)
	metrics := &Metrics{}

	// Crear un handler con gatherer personalizado
	handler := newHandlerWithGatherer(metrics, registry)

	app := fiber.New()
	handler.RegisterRoutes(app)

	// Verificar que la ruta /metrics existe
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Si la ruta no estuviera registrada, obtendríamos 404
	if resp.StatusCode == 404 {
		t.Error("Expected /metrics route to be registered, got 404")
	}
}

func TestHandleMetricsWithNoMetrics(t *testing.T) {
	// Test del endpoint cuando no hay métricas registradas
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

	// NO registrar ninguna métrica

	handler := newHandlerWithGatherer(metrics, registry)
	app := fiber.New()
	handler.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Debe responder 200 aunque no haya métricas
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 even with no metrics, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Debe contener los HELP y TYPE aunque no haya valores
	if !strings.Contains(bodyStr, "# HELP") {
		t.Error("Expected response to contain metric help text")
	}
}
