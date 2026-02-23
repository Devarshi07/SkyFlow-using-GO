package flights

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/cache"
)

// CachedFlightStore wraps a FlightStore and only caches search operations.
// All other operations (CRUD) pass through directly to the inner store.
type CachedFlightStore struct {
	inner FlightStore
	cache *cache.RedisCache
}

func NewCachedFlightStore(inner FlightStore, c *cache.RedisCache) *CachedFlightStore {
	return &CachedFlightStore{inner: inner, cache: c}
}

// ── Pass-through (no caching) ──────────────────────────────

func (s *CachedFlightStore) Create(ctx context.Context, f *Flight) (*Flight, error) {
	out, err := s.inner.Create(ctx, f)
	if err != nil {
		return nil, err
	}
	// Invalidate search cache since a new flight may affect results
	_ = s.cache.DeletePattern(ctx, "flights:search:*")
	return out, nil
}

func (s *CachedFlightStore) GetByID(ctx context.Context, id string) (*Flight, bool) {
	return s.inner.GetByID(ctx, id)
}

func (s *CachedFlightStore) List(ctx context.Context) []*Flight {
	return s.inner.List(ctx)
}

func (s *CachedFlightStore) Update(ctx context.Context, id string, upd UpdateFlightRequest) (*Flight, bool) {
	out, ok := s.inner.Update(ctx, id, upd)
	if ok {
		// Invalidate search cache since updated flight may affect results
		_ = s.cache.DeletePattern(ctx, "flights:search:*")
	}
	return out, ok
}

func (s *CachedFlightStore) Delete(ctx context.Context, id string) bool {
	ok := s.inner.Delete(ctx, id)
	if ok {
		_ = s.cache.DeletePattern(ctx, "flights:search:*")
	}
	return ok
}

// ── Cached: search operations only ─────────────────────────

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
	key := "flights:search:connecting:" + origin + ":" + dest + ":" + date
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
