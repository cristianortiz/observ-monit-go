package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cristianortiz/observ-monit-go/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresDB struct {
	Pool   *pgxpool.Pool
	Config *config.Config
	Logger *zap.Logger
}

// NewPostgresDB creates and configure a new db pool with pgxpool
func NewPostgresDB(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*PostgresDB, error) {
	logger.Info("initializing PostgreSQL connection pool",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
		zap.Int32("max_conns", cfg.Database.MaxConns),
		zap.Int32("min_conns", cfg.Database.MinConns),
	)

	connString := cfg.GetDatabaseURL()

	logger.Debug("database connection string generated",
		zap.String("url", maskPassword(connString)),
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.Database.HealthCheckInterval

	logger.Info("pgxpool configuration set",
		zap.Int32("max_conns", poolConfig.MaxConns),
		zap.Int32("min_conns", poolConfig.MinConns),
		zap.Duration("max_conn_lifetime", poolConfig.MaxConnLifetime),
		zap.Duration("max_conn_idle_time", poolConfig.MaxConnIdleTime),
		zap.Duration("health_check_period", poolConfig.HealthCheckPeriod),
	)

	// Create pool with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	//check initial conn  with ping
	if err := pool.Ping(ctxWithTimeout); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection pool created and verified successfully")

	return &PostgresDB{
		Pool:   pool,
		Config: cfg,
		Logger: logger,
	}, nil
}

// Safetly closing for connections pool
func (p *PostgresDB) Close() {
	if p.Pool == nil {
		p.Logger.Warn("database pool already closed")
		return
	}

	p.Logger.Info("closing database connection pool")
	p.Pool.Close()
	p.Logger.Info("database connection pool closed successfully")
}

func (p *PostgresDB) Ping(ctx context.Context) error {
	if p.Pool == nil {
		return fmt.Errorf("pool is not initialized")
	}
	return p.Pool.Ping(ctx)
}

// Acquire get a connection from the pool
// IMPORTANT: always call conn.Release() when their work is finish
func (p *PostgresDB) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.Pool.Acquire(ctx)
}

func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := p.Pool.Exec(ctx, query, args...)
	return err
}

func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return p.Pool.Query(ctx, query, args...)
}

func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return p.Pool.QueryRow(ctx, query, args...)
}

func (p *PostgresDB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return p.Pool.Begin(ctx)
}

// Stats return statistics about db pool
func (p *PostgresDB) Stats() *pgxpool.Stat {
	if p.Pool == nil {
		return nil
	}
	return p.Pool.Stat()
}

// GetPoolStats return stats in map format (useful for metrics)
func (p *PostgresDB) GetPoolStats() map[string]interface{} {
	stats := p.Pool.Stat()

	return map[string]interface{}{
		"max_conns": p.Pool.Config().MaxConns,
		"min_conns": p.Pool.Config().MinConns,

		"acquire_count":  stats.AcquireCount(),
		"acquired_conns": stats.AcquiredConns(),
		"idle_conns":     stats.IdleConns(),
		"total_conns":    stats.TotalConns(),

		"new_conns_count":            stats.NewConnsCount(),
		"max_lifetime_destroy_count": stats.MaxLifetimeDestroyCount(),
		"max_idle_destroy_count":     stats.MaxIdleDestroyCount(),

		// wait times
		"acquire_duration_ms":    stats.AcquireDuration().Milliseconds(),
		"empty_acquire_count":    stats.EmptyAcquireCount(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
	}
}

func maskPassword(connString string) string {
	// postgres://user:password@host:port/db?sslmode=disable
	// Converts to: postgres://user:****@host:port/db?sslmode=disable

	var masked string
	if idx := strings.Index(connString, "://"); idx != -1 {
		prefix := connString[:idx+3]
		rest := connString[idx+3:]

		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			userPass := rest[:atIdx]
			hostAndRest := rest[atIdx:]

			if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
				user := userPass[:colonIdx]
				masked = prefix + user + ":****" + hostAndRest
			} else {
				masked = connString
			}
		} else {
			masked = connString
		}
	} else {
		masked = connString
	}

	return masked
}
