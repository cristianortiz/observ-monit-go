package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler handles the metrics EP for prometheus
type Handler struct {
	metrics  *Metrics
	gatherer prometheus.Gatherer
}

func NewHandler(metrics *Metrics) *Handler {
	return &Handler{
		metrics:  metrics,
		gatherer: prometheus.DefaultGatherer, //production use
	}
}

// HandlerMetrics expoes ep /metrics for prometheus using prometheus standard handle
// (promhttp.Handler) and uses and adaptor to make it works with fiber
func (h *Handler) HandleMetrics(c *fiber.Ctx) error {
	//promhttp.Handler() returns http.Handler (std library)
	// adaptor.HTTPHandler converts http.Handler to fiber.Handler
	//  h.gatherer (default in prod, custom for testing)
	handler := adaptor.HTTPHandler(promhttp.HandlerFor(h.gatherer, promhttp.HandlerOpts{}))
	return handler(c)

}

// RegisteRoutes register routes for metrics EP in fiber app
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/metrics", h.HandleMetrics)
}

// newHandlerWithGatherer only for TESTING purposes
// private, to use only
func newHandlerWithGatherer(metrics *Metrics, gatherer prometheus.Gatherer) *Handler {
	return &Handler{
		metrics:  metrics,
		gatherer: gatherer,
	}
}
