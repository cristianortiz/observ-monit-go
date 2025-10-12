package health

import (
	"github.com/cristianortiz/observ-monit-go/pkg/observability/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Handler handles health checks EPs
type Handler struct {
	health *Health
	logger *logger.Logger
}

// NewHandler creates a new health checks Handler
func NewHandler(health *Health, log *logger.Logger) *Handler {
	return &Handler{
		health: health,
		logger: log.WithComponent("health-handler"),
	}
}

// HandleHealth manages EP /health (liveness probe)
func (h *Handler) HandleHealth(c *fiber.Ctx) error {
	response := h.health.CheckLiveness()
	h.logger.Debug("Health check performed",
		zap.String("status", string(response.Status)),
		zap.String("EP", "/health"),
	)

	if response.Status == StatusHealthy {
		return c.Status(fiber.StatusOK).JSON(response)
	}
	return c.Status(fiber.StatusServiceUnavailable).JSON(response)
}

func (h *Handler) HandleReady(c *fiber.Ctx) error {
	response := h.health.CheckReadiness()
	h.logger.Debug("Readiness check performed",
		zap.String("status", string(response.Status)),
		zap.String("EP", "/ready"),
	)
	if response.Status == StatusHealthy {
		return c.Status(fiber.StatusOK).JSON(response)
	}

	return c.Status(fiber.StatusServiceUnavailable).JSON(response)
}

// RegisterRoutes register health routes in Fiber app
func (h *Handler) RegisterRoutes(app *fiber.App, healthPath, readyPath string) {
	h.logger.Info("Registering health check routes",
		zap.String("health_path", healthPath),
		zap.String("ready_path", readyPath),
	)

	app.Get(healthPath, h.HandleHealth)
	app.Get(readyPath, h.HandleReady)
}
