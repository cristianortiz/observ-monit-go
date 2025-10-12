package health

import "time"

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

type Response struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version,omitempty"`
}

type Health struct {
	serviceName string
	version     string
}

// New creates a new instance of health
func New(serviceName, version string) *Health {
	return &Health{
		serviceName: serviceName,
		version:     version,
	}
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
// Used by K8S liveness probe
func (h *Health) CheckLiveness() Response {
	return h.Check()
}

// CheckReadiness verifica si está listo para recibir tráfico
// Usado por Kubernetes readiness probe
// TODO: Cuando tengamos DB, agregar check de conexión aquí

// CheckReadiness verifies if the service is ready to receive traffic
// Used by K8S readiness probe
// TODO: adds checks for DB, or other dependencies later
func (h *Health) CheckReadiness() Response {
	// for now, always ready

	return h.Check()
}
