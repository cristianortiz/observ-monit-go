package health

import (
	"context"
	"time"
)

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

type Response struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version,omitempty"`
	Checks    map[string]CheckDetail `json:"checks,omitempty"`
}

// CheckDetail represents details of a specific check
type CheckDetail struct {
	Status       Status                 `json:"status"`
	ResponseTime int64                  `json:"response_time_ms,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// DatabaseChecker is a simple interface for database health checks
type DatabaseChecker interface {
	Ping(ctx context.Context) error
	GetPoolStats() map[string]interface{}
}

type Health struct {
	serviceName string
	version     string
	database    DatabaseChecker
}

// New creates a new instance of health
func New(serviceName, version string) *Health {
	return &Health{
		serviceName: serviceName,
		version:     version,
	}
}

// SetDatabase registers a database for health checks
func (h *Health) SetDatabase(db DatabaseChecker) {
	h.database = db
}

// Check verifyes basic system`s health
func (h *Health) Check() Response {
	return Response{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Service:   h.serviceName,
		Version:   h.version,
	}
}

// CheckLiveness verifies if the service is alive
// Used by K8S liveness probe,
func (h *Health) CheckLiveness() Response {
	return h.Check()
}

// CheckReadiness verifies if the service is ready to receive traffic
// Used by K8S readiness probe also check database status
func (h *Health) CheckReadiness() Response {
	response := h.Check()

	//id no db configured service is ready,
	if h.database == nil {
		return response
	}
	// init checks map
	response.Checks = make(map[string]CheckDetail)

	//check  db connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.database.Ping(ctx)
	duration := time.Since(start)

	dbCheck := CheckDetail{
		ResponseTime: duration.Milliseconds(),
	}

	if err != nil {
		// Database unhealthy = service not ready
		dbCheck.Status = StatusUnhealthy
		dbCheck.Error = err.Error()
		dbCheck.Message = "database connection failed"
		response.Status = StatusUnhealthy
		response.Checks["database"] = dbCheck

		return response
	}
	if duration > 1*time.Second {
		// Database slow = service degraded
		dbCheck.Status = StatusDegraded
		dbCheck.Message = "database responding slowly"
		response.Status = StatusDegraded
		response.Checks["database"] = dbCheck

		return response
	}
	// Database healthy
	dbCheck.Status = StatusHealthy
	dbCheck.Message = "database connection healthy"
	dbCheck.Details = h.database.GetPoolStats()
	response.Checks["database"] = dbCheck
	return response

}
