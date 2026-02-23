package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 5 * time.Minute

// RedisCache provides get/set/delete for JSON-serializable values
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &RedisCache{client: client, ttl: ttl}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(val, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, val any) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, b, c.ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func FlightsListKey() string       { return "flights:list" }
func FlightsSearchKey(o, d, dt string) string { return fmt.Sprintf("flights:search:%s:%s:%s", o, d, dt) }
func FlightsIDKey(id string) string { return "flights:id:" + id }
func CitiesListKey() string        { return "cities:list" }
func CitiesIDKey(id string) string { return "cities:id:" + id }
func AirportsListKey() string      { return "airports:list" }
func AirportsIDKey(id string) string { return "airports:id:" + id }
