package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cristianortiz/observ-monit-go/internal/users/adapters/postgres"
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http"
	"github.com/cristianortiz/observ-monit-go/internal/users/usecase"
	"github.com/cristianortiz/observ-monit-go/pkg/config"
	"github.com/cristianortiz/observ-monit-go/pkg/database"
	"github.com/cristianortiz/observ-monit-go/pkg/http-utils/middleware"
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
	//Init Validator
	middleware.InitValidator()
	log.Info("Validator initialized")

	log.Info("Starting Factorit Platform",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
		zap.Int("port", cfg.Service.Port),
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
	// 5. INITIALIZE USERS MODULE (NUEVO)
	// ========================================

	// Dependency Injection: Repository → Service → Handler
	userRepository := postgres.NewUserRepository(db.Pool)
	userService := usecase.NewUserService(userRepository)
	userHandler := http.NewUserHandler(userService)

	log.Info("Users module initialized",
		zap.String("repository", "postgres"),
		zap.String("service", "user_service"),
		zap.String("handler", "user_handler"),
	)

	// ========================================
	// 6. CREATE FIBER APP
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
	// 7. REGISTER OBSERVABILITY ROUTES
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
	// 8. REGISTER MODULE ROUTES
	// ========================================
	app.Group("/api") // Will be used when modules are ready

	// ✅ USERS MODULE ROUTES (Ya no es TODO)
	http.RegisterRoutes(app, userHandler)

	log.Info("Users module routes registered",
		zap.String("prefix", "/api/users"),
		zap.Strings("endpoints", []string{
			"POST /api/users",
			"GET /api/users",
			"GET /api/users/:id",
			"PUT /api/users/:id",
			"DELETE /api/users/:id",
		}),
	)

	// TODO: Products Module Routes (Future)
	// productsGroup := api.Group("/products")
	// productsHandlers.RegisterRoutes(productsGroup)

	// TODO: Orders Module Routes (Future)
	// ordersGroup := api.Group("/orders")
	// ordersHandlers.RegisterRoutes(ordersGroup)

	log.Info("API ready",
		zap.String("base_path", "/api"),
		zap.String("users_module", "✅ active"),
		zap.String("products_module", "⏳ pending"),
		zap.String("orders_module", "⏳ pending"),
	)

	// ========================================
	// 9. START HTTP SERVER
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
	// 10. GRACEFUL SHUTDOWN
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
