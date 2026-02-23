package cities

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemStore struct {
	mu     sync.RWMutex
	cities map[string]*City
}

func NewStore() *MemStore {
	return &MemStore{cities: make(map[string]*City)}
}

func (s *MemStore) Create(ctx context.Context, c *City) (*City, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	s.cities[c.ID] = c
	return c, nil
}

func (s *MemStore) GetByID(ctx context.Context, id string) (*City, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.cities[id]
	return c, ok
}

func (s *MemStore) List(ctx context.Context) []*City {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*City, 0, len(s.cities))
	for _, c := range s.cities {
		out = append(out, c)
	}
	return out
}

func (s *MemStore) Update(ctx context.Context, id string, upd UpdateCityRequest) (*City, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cities[id]
	if !ok {
		return nil, false
	}
	if upd.Name != nil {
		c.Name = *upd.Name
	}
	if upd.Country != nil {
		c.Country = *upd.Country
	}
	if upd.Code != nil {
		c.Code = *upd.Code
	}
	return c, true
}

func (s *MemStore) Delete(ctx context.Context, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cities[id]; !ok {
		return false
	}
	delete(s.cities, id)
	return true
}
