package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Helper function para crear métricas de test con registry custom
func setupTestMetrics(_ *testing.T) (*Metrics, *prometheus.Registry) {
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

	return metrics, registry
}

func TestNew(t *testing.T) {
	serviceName := "test-service"

	// Crear métricas usando New()
	metrics := New(serviceName)

	if metrics == nil {
		t.Fatal("New() returned nil")
	}

	// Verificar que todas las métricas fueron inicializadas
	if metrics.httpRequestsTotal == nil {
		t.Error("httpRequestsTotal was not initialized")
	}

	if metrics.httpRequestDuration == nil {
		t.Error("httpRequestDuration was not initialized")
	}

	if metrics.httpRequestSize == nil {
		t.Error("httpRequestSize was not initialized")
	}

	if metrics.httpResponseSize == nil {
		t.Error("httpResponseSize was not initialized")
	}

	if metrics.activeConnections == nil {
		t.Error("activeConnections was not initialized")
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Registrar una request
	metrics.RecordHTTPRequest("test-service", "GET", "/api/users", "200")

	// Verificar que el counter se incrementó
	value := testutil.ToFloat64(
		metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/api/users", "200"),
	)

	if value != 1 {
		t.Errorf("Expected counter value 1, got %f", value)
	}

	// Registrar otra request al mismo endpoint
	metrics.RecordHTTPRequest("test-service", "GET", "/api/users", "200")

	// Verificar que el counter ahora es 2
	value = testutil.ToFloat64(
		metrics.httpRequestsTotal.WithLabelValues("test-service", "GET", "/api/users", "200"),
	)

	if value != 2 {
		t.Errorf("Expected counter value 2, got %f", value)
	}
}

func TestRecordHTTPRequestWithDifferentLabels(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Registrar requests con diferentes combinaciones de labels
	metrics.RecordHTTPRequest("test-service", "GET", "/api/users", "200")
	metrics.RecordHTTPRequest("test-service", "POST", "/api/users", "201")
	metrics.RecordHTTPRequest("test-service", "GET", "/api/products", "200")
	metrics.RecordHTTPRequest("test-service", "GET", "/api/users", "404")

	// Verificar cada combinación de labels
	tests := []struct {
		method   string
		path     string
		status   string
		expected float64
	}{
		{"GET", "/api/users", "200", 1},
		{"POST", "/api/users", "201", 1},
		{"GET", "/api/products", "200", 1},
		{"GET", "/api/users", "404", 1},
	}

	for _, tt := range tests {
		value := testutil.ToFloat64(
			metrics.httpRequestsTotal.WithLabelValues("test-service", tt.method, tt.path, tt.status),
		)

		if value != tt.expected {
			t.Errorf("Expected %f for %s %s %s, got %f",
				tt.expected, tt.method, tt.path, tt.status, value)
		}
	}
}

func TestRecordHTTPDuration(t *testing.T) {
	metrics, registry := setupTestMetrics(t)

	// Registrar varias duraciones
	durations := []float64{0.001, 0.015, 0.120, 0.350, 1.500}

	for _, duration := range durations {
		metrics.RecordHTTPDuration("test-service", "GET", "/api/users", "200", duration)
	}

	// Para histograms, verificamos usando el registry completo
	// Buscamos la métrica en el output de texto
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Buscar la métrica http_request_duration_seconds
	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "http_request_duration_seconds" {
			found = true
			// Verificar que hay al menos una métrica registrada
			if len(mf.GetMetric()) == 0 {
				t.Error("Expected histogram to have recorded observations")
			}

			// Verificar el count (número de observaciones)
			for _, m := range mf.GetMetric() {
				if m.GetHistogram() != nil {
					count := m.GetHistogram().GetSampleCount()
					if count != uint64(len(durations)) {
						t.Errorf("Expected %d observations, got %d", len(durations), count)
					}
				}
			}
		}
	}

	if !found {
		t.Error("http_request_duration_seconds metric not found")
	}
}

func TestRecordHTTPRequestSize(t *testing.T) {
	metrics, registry := setupTestMetrics(t)

	// Registrar tamaños de request
	sizes := []float64{100, 200, 1500, 5000}

	for _, size := range sizes {
		metrics.RecordHTTPRequestSize("test-service", "POST", "/api/users", size)
	}

	// Para summaries, verificamos usando el registry completo
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Buscar la métrica http_request_size_bytes
	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "http_request_size_bytes" {
			found = true
			// Verificar que hay al menos una métrica registrada
			if len(mf.GetMetric()) == 0 {
				t.Error("Expected summary to have recorded observations")
			}

			// Verificar el count y sum
			for _, m := range mf.GetMetric() {
				if m.GetSummary() != nil {
					count := m.GetSummary().GetSampleCount()
					if count != uint64(len(sizes)) {
						t.Errorf("Expected %d observations, got %d", len(sizes), count)
					}

					sum := m.GetSummary().GetSampleSum()
					expectedSum := 0.0
					for _, s := range sizes {
						expectedSum += s
					}
					if sum != expectedSum {
						t.Errorf("Expected sum %f, got %f", expectedSum, sum)
					}
				}
			}
		}
	}

	if !found {
		t.Error("http_request_size_bytes metric not found")
	}
}

func TestRecordHTTPResponseSize(t *testing.T) {
	metrics, registry := setupTestMetrics(t)

	// Registrar tamaños de response
	sizes := []float64{250, 1000, 3500, 10000}

	for _, size := range sizes {
		metrics.RecordHTTPResponseSize("test-service", "GET", "/api/products", size)
	}

	// Para summaries, verificamos usando el registry completo
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Buscar la métrica http_response_size_bytes
	found := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "http_response_size_bytes" {
			found = true
			// Verificar que hay al menos una métrica registrada
			if len(mf.GetMetric()) == 0 {
				t.Error("Expected summary to have recorded observations")
			}

			// Verificar el count y sum
			for _, m := range mf.GetMetric() {
				if m.GetSummary() != nil {
					count := m.GetSummary().GetSampleCount()
					if count != uint64(len(sizes)) {
						t.Errorf("Expected %d observations, got %d", len(sizes), count)
					}

					sum := m.GetSummary().GetSampleSum()
					expectedSum := 0.0
					for _, s := range sizes {
						expectedSum += s
					}
					if sum != expectedSum {
						t.Errorf("Expected sum %f, got %f", expectedSum, sum)
					}
				}
			}
		}
	}

	if !found {
		t.Error("http_response_size_bytes metric not found")
	}
}

func TestIncActiveConnections(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Estado inicial debe ser 0
	value := testutil.ToFloat64(metrics.activeConnections)
	if value != 0 {
		t.Errorf("Expected initial value 0, got %f", value)
	}

	// Incrementar conexiones activas
	metrics.IncActiveConnections()

	value = testutil.ToFloat64(metrics.activeConnections)
	if value != 1 {
		t.Errorf("Expected value 1 after increment, got %f", value)
	}

	// Incrementar de nuevo
	metrics.IncActiveConnections()

	value = testutil.ToFloat64(metrics.activeConnections)
	if value != 2 {
		t.Errorf("Expected value 2 after second increment, got %f", value)
	}
}

func TestDecActiveConnections(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Incrementar primero
	metrics.IncActiveConnections()
	metrics.IncActiveConnections()
	metrics.IncActiveConnections()

	value := testutil.ToFloat64(metrics.activeConnections)
	if value != 3 {
		t.Errorf("Expected value 3 after increments, got %f", value)
	}

	// Decrementar
	metrics.DecActiveConnections()

	value = testutil.ToFloat64(metrics.activeConnections)
	if value != 2 {
		t.Errorf("Expected value 2 after decrement, got %f", value)
	}

	// Decrementar de nuevo
	metrics.DecActiveConnections()

	value = testutil.ToFloat64(metrics.activeConnections)
	if value != 1 {
		t.Errorf("Expected value 1 after second decrement, got %f", value)
	}

	// Decrementar hasta 0
	metrics.DecActiveConnections()

	value = testutil.ToFloat64(metrics.activeConnections)
	if value != 0 {
		t.Errorf("Expected value 0 after third decrement, got %f", value)
	}
}

func TestActiveConnectionsLifecycle(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Simular ciclo de vida de múltiples conexiones
	// Conexión 1 abre
	metrics.IncActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 1 {
		t.Error("Expected 1 active connection")
	}

	// Conexión 2 abre
	metrics.IncActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 2 {
		t.Error("Expected 2 active connections")
	}

	// Conexión 1 cierra
	metrics.DecActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 1 {
		t.Error("Expected 1 active connection")
	}

	// Conexión 3 abre
	metrics.IncActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 2 {
		t.Error("Expected 2 active connections")
	}

	// Conexión 2 cierra
	metrics.DecActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 1 {
		t.Error("Expected 1 active connection")
	}

	// Conexión 3 cierra
	metrics.DecActiveConnections()
	if testutil.ToFloat64(metrics.activeConnections) != 0 {
		t.Error("Expected 0 active connections")
	}
}

func TestMultipleMetricsRecording(t *testing.T) {
	metrics, _ := setupTestMetrics(t)

	// Simular el registro completo de una request
	serviceName := "test-service"
	method := "POST"
	path := "/api/orders"
	status := "201"

	// 1. Incrementar conexiones activas
	metrics.IncActiveConnections()

	// 2. Registrar tamaño de request
	metrics.RecordHTTPRequestSize(serviceName, method, path, 1024)

	// 3. Registrar la request
	metrics.RecordHTTPRequest(serviceName, method, path, status)

	// 4. Registrar duración
	metrics.RecordHTTPDuration(serviceName, method, path, status, 0.125)

	// 5. Registrar tamaño de response
	metrics.RecordHTTPResponseSize(serviceName, method, path, 2048)

	// 6. Decrementar conexiones activas
	metrics.DecActiveConnections()

	// Verificar que todas las métricas se registraron
	counterValue := testutil.ToFloat64(
		metrics.httpRequestsTotal.WithLabelValues(serviceName, method, path, status),
	)
	if counterValue != 1 {
		t.Errorf("Expected counter value 1, got %f", counterValue)
	}

	gaugeValue := testutil.ToFloat64(metrics.activeConnections)
	if gaugeValue != 0 {
		t.Errorf("Expected gauge value 0, got %f", gaugeValue)
	}
}
