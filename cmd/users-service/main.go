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
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	// load config
	cfg, err := config.Load("users-service")
	if err != nil {
		panic("failed to load config " + err.Error())
	}

	//init logger
	log, err := logger.New(cfg.Observability.LogLevel, cfg.Debug)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("starting users-service",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
	)

	//init db pool
	ctx := context.Background()
	db, err := database.NewPostgresDB(ctx, cfg, log.Logger)
	if err != nil {
		log.Fatal("failed to init database", zap.Error(err))
	}
	defer db.Close()

	log.Info("database initialized",
		zap.Any("pool_stats", db.GetPoolStats()),
	)

	//init health check system
	healthSystem := health.New(cfg.Service.Name, "1.0.0")

	//register db health checks
	healthSystem.SetDatabase(db)
	healthHandler := health.NewHandler(healthSystem, log)

	//creates fiber app
	app := fiber.New(fiber.Config{
		AppName: "Users Service v1.0.0",
	})

	//  Register health routes
	healthHandler.RegisterRoutes(
		app,
		cfg.Observability.HealthPath,
		cfg.Observability.ReadyPath,
	)

	// Start server
	go func() {
		addr := cfg.GetServiceAddress()
		log.Info("starting HTTP server", zap.String("address", addr))

		if err := app.Listen(addr); err != nil {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	//Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Info("shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped")

}
