package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/skyflow/skyflow/internal/shared/config"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

// NewClient creates a Redis client
func NewClient(ctx context.Context, log *logger.Logger) (*redis.Client, error) {
	url := config.DB().Redis
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	log.Info("redis connected", "addr", opts.Addr)
	return client, nil
}
