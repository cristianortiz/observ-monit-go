package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all service metrics
type Metrics struct {
	//http metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.SummaryVec
	httpResponseSize    *prometheus.SummaryVec

	// Error metrics
	httpClientErrors *prometheus.CounterVec // 4xx errors
	httpServerErrors *prometheus.CounterVec // 5xx errors
	httpSlowRequests *prometheus.CounterVec // Requests exceeding SLO threshold

	//system metrics
	activeConnections prometheus.Gauge
}

// New creates a new instance of metrics
func New(serviceName string) *Metrics {
	return &Metrics{
		// counter: Total http requests
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "path", "status"},
		),
		// Histogram: HTTP request duration
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
			},
			[]string{"service", "method", "path", "status"},
		),
		httpRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "HTTP request size in bytes",
			},
			[]string{"service", "method", "path"},
		),

		// Summary: response size
		httpResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "HTTP response size in bytes",
			},
			[]string{"service", "method", "path"},
		),

		// Gauge: active connections
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),

		// Counter: HTTP 4xx client errors
		httpClientErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_client_errors_total",
				Help: "Total number of HTTP 4xx client errors",
			},
			[]string{"service", "method", "path", "status"},
		),

		// Counter: HTTP 5xx server errors
		httpServerErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_server_errors_total",
				Help: "Total number of HTTP 5xx server errors",
			},
			[]string{"service", "method", "path", "status"},
		),

		// Counter: Slow requests (exceeding SLO threshold)
		httpSlowRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_slow_requests_total",
				Help: "Total number of HTTP requests exceeding SLO threshold",
			},
			[]string{"service", "method", "path", "threshold"},
		),
	}
}

// RecordHTTPRequest register a new HTTP request, increments total requests counter
func (m *Metrics) RecordHTTPRequest(serviceName, method, path, status string) {
	m.httpRequestsTotal.WithLabelValues(serviceName, method, path, status).Inc()
}

// RecordHTTP register HTTP request duration, watch the value in the histogram to calculate percentiles
func (m *Metrics) RecordHTTPDuration(serviceName, method, path, status string, duration float64) {
	m.httpRequestDuration.WithLabelValues(serviceName, method, path, status).Observe(duration)
}

// RecordHTTPRequestSize register Request size, watch the value in the symarry to calculate average
func (m *Metrics) RecordHTTPRequestSize(serviceName, method, path string, size float64) {
	m.httpRequestSize.WithLabelValues(serviceName, method, path).Observe(size)
}

// RecordHTTPRequestSize register Response size, watch the value in the sumary to calculate average
func (m *Metrics) RecordHTTPResponseSize(serviceName, method, path string, size float64) {
	m.httpResponseSize.WithLabelValues(serviceName, method, path).Observe(size)
}

// IncActiveConnections increments the active connections counter, increase the gauge in 1
func (m *Metrics) IncActiveConnections() {
	m.activeConnections.Inc()
}

// DecActiveConnections decrease the active connections counter, decrease the gauge in 1
func (m *Metrics) DecActiveConnections() {
	m.activeConnections.Dec()
}

// RecordHTTPClientError records a 4xx client error
func (m *Metrics) RecordHTTPClientError(serviceName, method, path, status string) {
	m.httpClientErrors.WithLabelValues(serviceName, method, path, status).Inc()
}

// RecordHTTPServerError records a 5xx server error
func (m *Metrics) RecordHTTPServerError(serviceName, method, path, status string) {
	m.httpServerErrors.WithLabelValues(serviceName, method, path, status).Inc()
}

// RecordSlowRequest records a request that exceeded the SLO threshold
func (m *Metrics) RecordSlowRequest(serviceName, method, path, threshold string) {
	m.httpSlowRequests.WithLabelValues(serviceName, method, path, threshold).Inc()
}
