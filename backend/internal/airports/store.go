package airports

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemStore struct {
	mu       sync.RWMutex
	airports map[string]*Airport
}

func NewStore() *MemStore {
	return &MemStore{airports: make(map[string]*Airport)}
}

func (s *MemStore) Create(ctx context.Context, a *Airport) (*Airport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.ID = uuid.New().String()
	a.CreatedAt = time.Now()
	s.airports[a.ID] = a
	return a, nil
}

func (s *MemStore) GetByID(ctx context.Context, id string) (*Airport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.airports[id]
	return a, ok
}

func (s *MemStore) List(ctx context.Context) []*Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Airport, 0, len(s.airports))
	for _, a := range s.airports {
		out = append(out, a)
	}
	return out
}

func (s *MemStore) Update(ctx context.Context, id string, upd UpdateAirportRequest) (*Airport, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.airports[id]
	if !ok {
		return nil, false
	}
	if upd.Name != nil {
		a.Name = *upd.Name
	}
	if upd.CityID != nil {
		a.CityID = *upd.CityID
	}
	if upd.Code != nil {
		a.Code = *upd.Code
	}
	return a, true
}

func (s *MemStore) Delete(ctx context.Context, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.airports[id]; !ok {
		return false
	}
	delete(s.airports, id)
	return true
}
