package airports

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/cache"
)

type CachedAirportStore struct {
	inner AirportStore
	cache *cache.RedisCache
}

func NewCachedAirportStore(inner AirportStore, c *cache.RedisCache) *CachedAirportStore {
	return &CachedAirportStore{inner: inner, cache: c}
}

func (s *CachedAirportStore) Create(ctx context.Context, a *Airport) (*Airport, error) {
	out, err := s.inner.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, cache.AirportsListKey())
	return out, nil
}

func (s *CachedAirportStore) GetByID(ctx context.Context, id string) (*Airport, bool) {
	key := cache.AirportsIDKey(id)
	var a Airport
	ok, err := s.cache.Get(ctx, key, &a)
	if err == nil && ok {
		return &a, true
	}
	out, ok := s.inner.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	_ = s.cache.Set(ctx, key, out)
	return out, true
}

func (s *CachedAirportStore) List(ctx context.Context) []*Airport {
	key := cache.AirportsListKey()
	var out []*Airport
	ok, err := s.cache.Get(ctx, key, &out)
	if err == nil && ok {
		return out
	}
	out = s.inner.List(ctx)
	_ = s.cache.Set(ctx, key, out)
	return out
}

func (s *CachedAirportStore) Update(ctx context.Context, id string, upd UpdateAirportRequest) (*Airport, bool) {
	out, ok := s.inner.Update(ctx, id, upd)
	if !ok {
		return nil, false
	}
	_ = s.cache.Delete(ctx, cache.AirportsIDKey(id), cache.AirportsListKey())
	return out, true
}

func (s *CachedAirportStore) Delete(ctx context.Context, id string) bool {
	ok := s.inner.Delete(ctx, id)
	if ok {
		_ = s.cache.Delete(ctx, cache.AirportsIDKey(id), cache.AirportsListKey())
	}
	return ok
}

var _ AirportStore = (*CachedAirportStore)(nil)
