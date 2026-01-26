package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewOTELMetrics_InitializesAuthMetrics verifica que las métricas de auth se inicializan junto con HTTP
func TestNewOTELMetrics_InitializesAuthMetrics(t *testing.T) {
	ctx := context.Background()

	otelMetrics, shutdown, err := NewOTELMetrics(ctx, OTELMetricsConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
		Enabled:        true,
	})

	require.NoError(t, err, "NewOTELMetrics should not fail")
	require.NotNil(t, otelMetrics, "otelMetrics should not be nil")
	defer shutdown(ctx)

	// Verificar que los counters de auth están inicializados
	assert.NotNil(t, otelMetrics.authAttempts, "authAttempts should be initialized")
	assert.NotNil(t, otelMetrics.authFailures, "authFailures should be initialized")
	assert.NotNil(t, otelMetrics.tokenValidations, "tokenValidations should be initialized")
	assert.NotNil(t, otelMetrics.permissionChecks, "permissionChecks should be initialized")
	assert.NotNil(t, otelMetrics.permissionDenials, "permissionDenials should be initialized")

	// Verificar que también los HTTP counters están inicializados
	assert.NotNil(t, otelMetrics.httpRequestsTotal, "httpRequestsTotal should be initialized")
	assert.NotNil(t, otelMetrics.httpRequestDuration, "httpRequestDuration should be initialized")
}

// TestOTELMetrics_RecordAuthAttempt verifica que se pueden registrar intentos de auth
func TestOTELMetrics_RecordAuthAttempt(t *testing.T) {
	ctx := context.Background()

	otelMetrics, shutdown, err := NewOTELMetrics(ctx, OTELMetricsConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
		Enabled:        true,
	})

	require.NoError(t, err)
	defer shutdown(ctx)

	// No debe crashear al registrar métricas
	assert.NotPanics(t, func() {
		otelMetrics.RecordAuthAttempt(ctx, true, "authenticated")
		otelMetrics.RecordAuthAttempt(ctx, false, "invalid token")
	}, "RecordAuthAttempt should not panic")
}

// TestOTELMetrics_RecordTokenValidation verifica validaciones de token
func TestOTELMetrics_RecordTokenValidation(t *testing.T) {
	ctx := context.Background()

	otelMetrics, shutdown, err := NewOTELMetrics(ctx, OTELMetricsConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
		Enabled:        true,
	})

	require.NoError(t, err)
	defer shutdown(ctx)

	assert.NotPanics(t, func() {
		otelMetrics.RecordTokenValidation(ctx, true, "https://test.auth0.com/")
		otelMetrics.RecordTokenValidation(ctx, false, "https://invalid.auth0.com/")
	}, "RecordTokenValidation should not panic")
}

// TestOTELMetrics_RecordPermissionCheck verifica verificaciones de permisos
func TestOTELMetrics_RecordPermissionCheck(t *testing.T) {
	ctx := context.Background()

	otelMetrics, shutdown, err := NewOTELMetrics(ctx, OTELMetricsConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
		Enabled:        true,
	})

	require.NoError(t, err)
	defer shutdown(ctx)

	assert.NotPanics(t, func() {
		otelMetrics.RecordPermissionCheck(ctx, "read:users", true, "user-123")
		otelMetrics.RecordPermissionCheck(ctx, "delete:users", false, "user-456")
	}, "RecordPermissionCheck should not panic")
}

// TestNewOTELMetrics_Disabled verifica que cuando está deshabilitado no falla
func TestNewOTELMetrics_Disabled(t *testing.T) {
	ctx := context.Background()

	otelMetrics, shutdown, err := NewOTELMetrics(ctx, OTELMetricsConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
		Enabled:        false, // Deshabilitado
	})

	require.NoError(t, err)
	require.NotNil(t, otelMetrics)
	require.NotNil(t, shutdown)

	// Shutdown no debe fallar
	err = shutdown(ctx)
	assert.NoError(t, err)
}
