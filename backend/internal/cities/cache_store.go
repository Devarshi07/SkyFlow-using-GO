package cities

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/cache"
)

type CachedCityStore struct {
	inner CityStore
	cache *cache.RedisCache
}

func NewCachedCityStore(inner CityStore, c *cache.RedisCache) *CachedCityStore {
	return &CachedCityStore{inner: inner, cache: c}
}

func (s *CachedCityStore) Create(ctx context.Context, c *City) (*City, error) {
	out, err := s.inner.Create(ctx, c)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, cache.CitiesListKey())
	return out, nil
}

func (s *CachedCityStore) GetByID(ctx context.Context, id string) (*City, bool) {
	key := cache.CitiesIDKey(id)
	var c City
	ok, err := s.cache.Get(ctx, key, &c)
	if err == nil && ok {
		return &c, true
	}
	out, ok := s.inner.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	_ = s.cache.Set(ctx, key, out)
	return out, true
}

func (s *CachedCityStore) List(ctx context.Context) []*City {
	key := cache.CitiesListKey()
	var out []*City
	ok, err := s.cache.Get(ctx, key, &out)
	if err == nil && ok {
		return out
	}
	out = s.inner.List(ctx)
	_ = s.cache.Set(ctx, key, out)
	return out
}

func (s *CachedCityStore) Update(ctx context.Context, id string, upd UpdateCityRequest) (*City, bool) {
	out, ok := s.inner.Update(ctx, id, upd)
	if !ok {
		return nil, false
	}
	_ = s.cache.Delete(ctx, cache.CitiesIDKey(id), cache.CitiesListKey())
	return out, true
}

func (s *CachedCityStore) Delete(ctx context.Context, id string) bool {
	ok := s.inner.Delete(ctx, id)
	if ok {
		_ = s.cache.Delete(ctx, cache.CitiesIDKey(id), cache.CitiesListKey())
	}
	return ok
}

var _ CityStore = (*CachedCityStore)(nil)
