package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skyflow/skyflow/internal/shared/config"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

// NewPool creates a PostgreSQL connection pool
func NewPool(ctx context.Context, log *logger.Logger) (*pgxpool.Pool, error) {
	cfg := config.DB()
	pool, err := pgxpool.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	log.Info("postgres connected")
	return pool, nil
}
