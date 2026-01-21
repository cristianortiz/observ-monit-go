package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OTELMetrics contiene todas las métricas del servicio usando OTEL SDK
type OTELMetrics struct {
	meter metric.Meter

	// HTTP metrics
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	httpClientErrors    metric.Int64Counter
	httpServerErrors    metric.Int64Counter
	httpSlowRequests    metric.Int64Counter
	activeConnections   metric.Int64UpDownCounter
}

// OTELMetricsConfig configuración para inicializar OTEL Metrics
type OTELMetricsConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
}

// NewOTELMetrics crea una nueva instancia de métricas OTEL
// Retorna: metrics, shutdownFunc, error
func NewOTELMetrics(ctx context.Context, config OTELMetricsConfig) (*OTELMetrics, func(context.Context) error, error) {
	if !config.Enabled {
		// Si está deshabilitado, retornar noop
		return &OTELMetrics{}, func(context.Context) error { return nil }, nil
	}

	// 1. Crear resource (metadata del servicio)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 2. Configurar OTLP gRPC exporter (envía al Collector)
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(config.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(), // Desactiva TLS (desarrollo)
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// 3. Crear MeterProvider (equivalente a prometheus.Registry)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			// PeriodicReader envía métricas cada X tiempo
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(10*time.Second), // Enviar cada 10 segundos
			),
		),
	)

	// 4. Registrar MeterProvider globalmente
	otel.SetMeterProvider(meterProvider)

	// 5. Obtener un Meter (factory de métricas)
	meter := meterProvider.Meter(
		config.ServiceName,
		metric.WithInstrumentationVersion(config.ServiceVersion),
	)

	// 6. Crear instrumentos de métricas
	m := &OTELMetrics{meter: meter}

	// Counter: Total HTTP requests
	m.httpRequestsTotal, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create httpRequestsTotal counter: %w", err)
	}

	// Histogram: HTTP request duration
	m.httpRequestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create httpRequestDuration histogram: %w", err)
	}

	// Counter: HTTP 4xx errors
	m.httpClientErrors, err = meter.Int64Counter(
		"http.server.client_errors.count",
		metric.WithDescription("Total number of HTTP 4xx client errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create httpClientErrors counter: %w", err)
	}

	// Counter: HTTP 5xx errors
	m.httpServerErrors, err = meter.Int64Counter(
		"http.server.server_errors.count",
		metric.WithDescription("Total number of HTTP 5xx server errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create httpServerErrors counter: %w", err)
	}

	// Counter: Slow requests
	m.httpSlowRequests, err = meter.Int64Counter(
		"http.server.slow_requests.count",
		metric.WithDescription("Total number of slow HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create httpSlowRequests counter: %w", err)
	}

	// UpDownCounter: Active connections (equivalente a Gauge)
	m.activeConnections, err = meter.Int64UpDownCounter(
		"http.server.active_connections",
		metric.WithDescription("Number of active HTTP connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create activeConnections gauge: %w", err)
	}

	return m, meterProvider.Shutdown, nil
}

// RecordHTTPRequest registra una request HTTP
func (m *OTELMetrics) RecordHTTPRequest(ctx context.Context, service, method, path, status string) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.status_code", status),
	}
	m.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordHTTPDuration registra la duración de una request HTTP
func (m *OTELMetrics) RecordHTTPDuration(ctx context.Context, service, method, path, status string, duration float64) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.status_code", status),
	}
	m.httpRequestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
}

// RecordHTTPClientError registra un error 4xx
func (m *OTELMetrics) RecordHTTPClientError(ctx context.Context, service, method, path, status string) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.status_code", status),
	}
	m.httpClientErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordHTTPServerError registra un error 5xx
func (m *OTELMetrics) RecordHTTPServerError(ctx context.Context, service, method, path, status string) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.status_code", status),
	}
	m.httpServerErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordSlowRequest registra una request lenta
func (m *OTELMetrics) RecordSlowRequest(ctx context.Context, service, method, path, threshold string) {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", service),
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("threshold", threshold),
	}
	m.httpSlowRequests.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// IncActiveConnections incrementa el contador de conexiones activas
func (m *OTELMetrics) IncActiveConnections(ctx context.Context) {
	m.activeConnections.Add(ctx, 1)
}

// DecActiveConnections decrementa el contador de conexiones activas
func (m *OTELMetrics) DecActiveConnections(ctx context.Context) {
	m.activeConnections.Add(ctx, -1)
}
