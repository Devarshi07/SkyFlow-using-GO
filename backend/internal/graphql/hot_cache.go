package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/skyflow/skyflow/internal/flights"
)

const (
	hotThreshold = 5
	freqWindow   = 30 * time.Minute
	cacheTTL     = 15 * time.Minute
)

type HotRouteCache struct {
	rdb *redis.Client
}

func NewHotRouteCache(rdb *redis.Client) *HotRouteCache {
	if rdb == nil {
		return nil
	}
	return &HotRouteCache{rdb: rdb}
}

func routeKey(origin, dest, date string) string {
	return fmt.Sprintf("hotroute:%s:%s:%s", origin, dest, date)
}

func freqKey(origin, dest, date string) string {
	return fmt.Sprintf("hotfreq:%s:%s:%s", origin, dest, date)
}

func (c *HotRouteCache) RecordSearch(origin, dest, date string) {
	if c == nil {
		return
	}
	ctx := context.Background()
	key := freqKey(origin, dest, date)
	c.rdb.Incr(ctx, key)
	c.rdb.Expire(ctx, key, freqWindow)
}

func (c *HotRouteCache) isHot(origin, dest, date string) bool {
	if c == nil {
		return false
	}
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, freqKey(origin, dest, date)).Int()
	if err != nil {
		return false
	}
	return val >= hotThreshold
}

func (c *HotRouteCache) Get(origin, dest, date string) (*flights.SearchResult, bool) {
	if c == nil {
		return nil, false
	}
	ctx := context.Background()
	data, err := c.rdb.Get(ctx, routeKey(origin, dest, date)).Bytes()
	if err != nil {
		return nil, false
	}
	var result flights.SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}
	return &result, true
}

func (c *HotRouteCache) MaybeCache(origin, dest, date string, result *flights.SearchResult) {
	if c == nil || result == nil {
		return
	}
	if !c.isHot(origin, dest, date) {
		return
	}
	ctx := context.Background()
	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	c.rdb.Set(ctx, routeKey(origin, dest, date), data, cacheTTL)
}
