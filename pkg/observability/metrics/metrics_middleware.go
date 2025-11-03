package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MetricsConfig struct {
	ServiceName string
	Metrics     *Metrics
}

func Middleware(config MetricsConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		//1. captures start information
		start := time.Now()

		//2. increase active conns (gauge +1)
		config.Metrics.IncActiveConnections()

		//4. executes next handler (real EP)
		err := c.Next()

		//5. captures info after the handler
		// IMPORTANT: Use c.Route().Path instead of c.Path() for better cardinality
		// c.Route().Path gives the route template (e.g., "/api/users/:id")
		// instead of the actual path (e.g., "/api/users/123")
		method := string([]byte(c.Method()))

		// Get route path, fallback to actual path if route is not found (404 cases)
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		//3. register size request (summary)
		requestSize := float64(len(c.Request().Body()))
		config.Metrics.RecordHTTPRequestSize(config.ServiceName, method, path, requestSize)

		//6. register final metrics
		//counter: Total of requests
		config.Metrics.RecordHTTPRequest(config.ServiceName, method, path, status)

		//histogram : request duration
		config.Metrics.RecordHTTPDuration(config.ServiceName, method, path, status, duration)

		//summary: response size
		responseSize := float64(len(c.Response().Body()))
		config.Metrics.RecordHTTPResponseSize(config.ServiceName, method, path, responseSize)

		// Record error metrics based on status code
		statusCode := c.Response().StatusCode()
		if statusCode >= 400 && statusCode < 500 {
			config.Metrics.RecordHTTPClientError(config.ServiceName, method, path, status)
		} else if statusCode >= 500 {
			config.Metrics.RecordHTTPServerError(config.ServiceName, method, path, status)
		}

		// Record slow requests (threshold: 1 second for demo, adjust based on your SLO)
		if duration > 1.0 {
			config.Metrics.RecordSlowRequest(config.ServiceName, method, path, "1s")
		}

		//7. decrease active connections gauge
		config.Metrics.DecActiveConnections()

		return err
	}
}
