package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type OTELMiddlewareConfig struct {
	ServiceName string
	Metrics     *OTELMetrics
}

func OTELMiddleware(config OTELMiddlewareConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Capturar tiempo de inicio
		start := time.Now()

		// 2. IMPORTANTE: Obtener context de Fiber
		// Este context tiene trace_id, span_id si hay tracing activo
		ctx := c.UserContext()

		// 3. Incrementar conexiones activas (Gauge +1)
		config.Metrics.IncActiveConnections(ctx)

		// 4. Ejecutar el handler real (endpoint)
		err := c.Next()

		// 5. Después del handler, registrar métricas

		// Obtener información de la request
		method := c.Method()

		// IMPORTANTE: Usar c.Route().Path en lugar de c.Path()
		// c.Route().Path = "/api/v1/users/:id" (template)
		// c.Path() = "/api/v1/users/123" (actual)
		// Usar template evita cardinalidad alta
		path := c.Route().Path
		if path == "" {
			// Si no hay route (404), usar path actual
			path = c.Path()
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		// 6. Registrar métricas HTTP básicas

		// Counter: Total de requests
		config.Metrics.RecordHTTPRequest(ctx, config.ServiceName, method, path, status)

		// Histogram: Duración de la request
		config.Metrics.RecordHTTPDuration(ctx, config.ServiceName, method, path, status, duration)

		// 7. Registrar métricas de errores

		statusCode := c.Response().StatusCode()
		if statusCode >= 400 && statusCode < 500 {
			// Errores 4xx (client errors)
			config.Metrics.RecordHTTPClientError(ctx, config.ServiceName, method, path, status)
		} else if statusCode >= 500 {
			// Errores 5xx (server errors)
			config.Metrics.RecordHTTPServerError(ctx, config.ServiceName, method, path, status)
		}

		// 8. Registrar requests lentas (> 1 segundo)
		if duration > 1.0 {
			config.Metrics.RecordSlowRequest(ctx, config.ServiceName, method, path, "1s")
		}

		// 9. Decrementar conexiones activas (Gauge -1)
		config.Metrics.DecActiveConnections(ctx)

		return err
	}
}
