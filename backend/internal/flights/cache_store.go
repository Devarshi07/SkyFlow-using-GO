package flights

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/cache"
)

type CachedFlightStore struct {
	inner FlightStore
	cache *cache.RedisCache
}

func NewCachedFlightStore(inner FlightStore, c *cache.RedisCache) *CachedFlightStore {
	return &CachedFlightStore{inner: inner, cache: c}
}

func (s *CachedFlightStore) Create(ctx context.Context, f *Flight) (*Flight, error) {
	out, err := s.inner.Create(ctx, f)
	if err != nil {
		return nil, err
	}
	_ = s.cache.DeletePattern(ctx, "flights:*")
	return out, nil
}

func (s *CachedFlightStore) GetByID(ctx context.Context, id string) (*Flight, bool) {
	key := cache.FlightsIDKey(id)
	var f Flight
	ok, err := s.cache.Get(ctx, key, &f)
	if err == nil && ok {
		return &f, true
	}
	out, ok := s.inner.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	_ = s.cache.Set(ctx, key, out)
	return out, true
}

func (s *CachedFlightStore) List(ctx context.Context) []*Flight {
	key := cache.FlightsListKey()
	var out []*Flight
	ok, err := s.cache.Get(ctx, key, &out)
	if err == nil && ok {
		return out
	}
	out = s.inner.List(ctx)
	_ = s.cache.Set(ctx, key, out)
	return out
}

func (s *CachedFlightStore) Update(ctx context.Context, id string, upd UpdateFlightRequest) (*Flight, bool) {
	out, ok := s.inner.Update(ctx, id, upd)
	if !ok {
		return nil, false
	}
	_ = s.cache.Delete(ctx, cache.FlightsIDKey(id))
	_ = s.cache.DeletePattern(ctx, "flights:*")
	return out, true
}

func (s *CachedFlightStore) Delete(ctx context.Context, id string) bool {
	ok := s.inner.Delete(ctx, id)
	if ok {
		_ = s.cache.Delete(ctx, cache.FlightsIDKey(id))
		_ = s.cache.DeletePattern(ctx, "flights:*")
	}
	return ok
}

func (s *CachedFlightStore) Search(ctx context.Context, origin, dest, date string) []*Flight {
	key := cache.FlightsSearchKey(origin, dest, date)
	var out []*Flight
	ok, err := s.cache.Get(ctx, key, &out)
	if err == nil && ok {
		return out
	}
	out = s.inner.Search(ctx, origin, dest, date)
	_ = s.cache.Set(ctx, key, out)
	return out
}

func (s *CachedFlightStore) SearchConnecting(ctx context.Context, origin, dest, date string) []*ConnectingFlight {
	key := "flights:connecting:" + origin + ":" + dest + ":" + date
	var out []*ConnectingFlight
	ok, err := s.cache.Get(ctx, key, &out)
	if err == nil && ok {
		return out
	}
	out = s.inner.SearchConnecting(ctx, origin, dest, date)
	_ = s.cache.Set(ctx, key, out)
	return out
}

var _ FlightStore = (*CachedFlightStore)(nil)
