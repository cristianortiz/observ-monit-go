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
	MeterProvider metric.MeterProvider
	meter         metric.Meter

	// HTTP metrics
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	httpClientErrors    metric.Int64Counter
	httpServerErrors    metric.Int64Counter
	httpSlowRequests    metric.Int64Counter
	activeConnections   metric.Int64UpDownCounter

	// Auth metrics
	authAttempts      metric.Int64Counter
	authFailures      metric.Int64Counter
	tokenValidations  metric.Int64Counter
	permissionChecks  metric.Int64Counter
	permissionDenials metric.Int64Counter
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
	m := &OTELMetrics{
		MeterProvider: meterProvider,
		meter:         meter,
	}

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

	// ========================================
	// AUTH METRICS
	// ========================================

	// Counter: Total de intentos de autenticación
	m.authAttempts, err = meter.Int64Counter(
		"auth.attempts.total",
		metric.WithDescription("Total number of authentication attempts"),
		metric.WithUnit("{attempt}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create authAttempts counter: %w", err)
	}

	// Counter: Total de fallos de autenticación
	m.authFailures, err = meter.Int64Counter(
		"auth.failures.total",
		metric.WithDescription("Total number of authentication failures"),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create authFailures counter: %w", err)
	}

	// Counter: Total de validaciones de token
	m.tokenValidations, err = meter.Int64Counter(
		"auth.token.validations.total",
		metric.WithDescription("Total number of token validations"),
		metric.WithUnit("{validation}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tokenValidations counter: %w", err)
	}

	// Counter: Total de verificaciones de permisos
	m.permissionChecks, err = meter.Int64Counter(
		"auth.permission.checks.total",
		metric.WithDescription("Total number of permission checks"),
		metric.WithUnit("{check}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create permissionChecks counter: %w", err)
	}

	// Counter: Total de denegaciones de permisos
	m.permissionDenials, err = meter.Int64Counter(
		"auth.permission.denials.total",
		metric.WithDescription("Total number of permission denials"),
		metric.WithUnit("{denial}"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create permissionDenials counter: %w", err)
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

// ========================================
// AUTH METRICS METHODS
// ========================================

// RecordAuthAttempt registra un intento de autenticación
func (m *OTELMetrics) RecordAuthAttempt(ctx context.Context, success bool, reason string) {
	if m.authAttempts == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Bool("success", success),
		attribute.String("reason", reason),
	}

	m.authAttempts.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Si falló, también incrementar el counter de fallos
	if !success && m.authFailures != nil {
		m.authFailures.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordTokenValidation registra una validación de token
func (m *OTELMetrics) RecordTokenValidation(ctx context.Context, valid bool, issuer string) {
	if m.tokenValidations == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Bool("valid", valid),
		attribute.String("issuer", issuer),
	}

	m.tokenValidations.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordPermissionCheck registra una verificación de permiso
func (m *OTELMetrics) RecordPermissionCheck(ctx context.Context, permission string, granted bool, userID string) {
	if m.permissionChecks == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("permission", permission),
		attribute.Bool("granted", granted),
		attribute.String("user_id", userID),
	}

	m.permissionChecks.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Si fue denegado, también incrementar el counter de denegaciones
	if !granted && m.permissionDenials != nil {
		m.permissionDenials.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}
