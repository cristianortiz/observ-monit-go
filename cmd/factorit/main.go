package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cristianortiz/observ-monit-go/pkg/config"
	"github.com/cristianortiz/observ-monit-go/pkg/database"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/health"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/logger"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/metrics"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	// ========================================
	// 1. LOAD CONFIGURATION
	// ========================================
	cfg, err := config.Load("factorit")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// ========================================
	// 2. INITIALIZE LOGGER
	// ========================================
	log, err := logger.New(cfg.Observability.LogLevel, cfg.Debug)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting Factorit Platform",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
		zap.String("architecture", "modular-monolith"),
	)

	// ========================================
	// 3. INITIALIZE DATABASE
	// ========================================
	ctx := context.Background()
	db, err := database.NewPostgresDB(ctx, cfg, log.Logger)
	if err != nil {
		log.Fatal("failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connection established",
		zap.Any("pool_stats", db.GetPoolStats()),
	)

	// ========================================
	// 4. INITIALIZE OBSERVABILITY
	// ========================================

	// Health Check System
	healthSystem := health.New(cfg.Service.Name, "1.0.0")
	healthSystem.SetDatabase(db)
	healthHandler := health.NewHandler(healthSystem, log)

	// Metrics System
	metricsSystem := metrics.New(cfg.Service.Name)
	metricsHandler := metrics.NewHandler(metricsSystem)

	log.Info("Observability systems initialized")

	// ========================================
	// 5. CREATE FIBER APP
	// ========================================
	app := fiber.New(fiber.Config{
		AppName:      "Factorit Platform v1.0.0",
		ErrorHandler: customErrorHandler(log),
	})

	// Global Middlewares
	app.Use(recover.New())
	app.Use(metrics.Middleware(metrics.MetricsConfig{
		ServiceName: cfg.Service.Name,
		Metrics:     metricsSystem,
	}))

	// ========================================
	// 6. REGISTER OBSERVABILITY ROUTES
	// ========================================
	healthHandler.RegisterRoutes(
		app,
		cfg.Observability.HealthPath,
		cfg.Observability.ReadyPath,
	)
	metricsHandler.RegisterRoutes(app)

	log.Info("Observability routes registered",
		zap.String("health", cfg.Observability.HealthPath),
		zap.String("ready", cfg.Observability.ReadyPath),
		zap.String("metrics", "/metrics"),
	)

	// ========================================
	// 7. REGISTER MODULE ROUTES
	// ========================================
	_ = app.Group("/api") // Will be used when modules are ready

	// TODO: Users Module Routes (FASE 6 - Handlers)
	// usersGroup := api.Group("/users")
	// usersHandlers.RegisterRoutes(usersGroup)

	// TODO: Products Module Routes (Future)
	// productsGroup := api.Group("/products")
	// productsHandlers.RegisterRoutes(productsGroup)

	// TODO: Orders Module Routes (Future)
	// ordersGroup := api.Group("/orders")
	// ordersHandlers.RegisterRoutes(ordersGroup)

	log.Info("Module routes will be registered here",
		zap.String("users_prefix", "/api/users"),
		zap.String("products_prefix", "/api/products"),
		zap.String("orders_prefix", "/api/orders"),
	)

	// ========================================
	// 8. START HTTP SERVER
	// ========================================
	go func() {
		addr := cfg.GetServiceAddress()
		log.Info("Starting HTTP server",
			zap.String("address", addr),
			zap.String("host", cfg.Service.Host),
			zap.Int("port", cfg.Service.Port),
		)

		if err := app.Listen(addr); err != nil {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// ========================================
	// 9. GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("❌ Server forced to shutdown", zap.Error(err))
	}

	log.Info("✅ Server stopped successfully")
}

// customErrorHandler handles errors globally
func customErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("HTTP error",
			zap.Error(err),
			zap.Int("status_code", code),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)

		return c.Status(code).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "An error occurred processing your request",
		})
	}
}
