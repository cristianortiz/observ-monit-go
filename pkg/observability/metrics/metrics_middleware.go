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
		// IMPORTANT: Make defensive copies of strings to avoid Fiber's internal pooling issues
		// Fiber reuses byte slices which can cause corruption when strings point to them
		method := string([]byte(c.Method()))
		path := string([]byte(c.Path()))
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
		//7. decrease active connections gauge
		config.Metrics.DecActiveConnections()

		return err
	}
}
