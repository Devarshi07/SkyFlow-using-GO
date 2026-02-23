package flights

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemStore struct {
	mu      sync.RWMutex
	flights map[string]*Flight
}

func NewStore() *MemStore {
	return &MemStore{flights: make(map[string]*Flight)}
}

func (s *MemStore) Create(ctx context.Context, f *Flight) (*Flight, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f.ID = uuid.New().String()
	f.CreatedAt = time.Now()
	f.SeatsAvailable = f.SeatsTotal
	s.flights[f.ID] = f
	return f, nil
}

func (s *MemStore) GetByID(ctx context.Context, id string) (*Flight, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flights[id]
	return f, ok
}

func (s *MemStore) List(ctx context.Context) []*Flight {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Flight, 0, len(s.flights))
	for _, f := range s.flights {
		out = append(out, f)
	}
	return out
}

func (s *MemStore) Update(ctx context.Context, id string, upd UpdateFlightRequest) (*Flight, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.flights[id]
	if !ok {
		return nil, false
	}
	if upd.FlightNumber != nil {
		f.FlightNumber = *upd.FlightNumber
	}
	if upd.OriginID != nil {
		f.OriginID = *upd.OriginID
	}
	if upd.DestinationID != nil {
		f.DestinationID = *upd.DestinationID
	}
	if upd.DepartureTime != nil {
		f.DepartureTime = *upd.DepartureTime
	}
	if upd.ArrivalTime != nil {
		f.ArrivalTime = *upd.ArrivalTime
	}
	if upd.Price != nil {
		f.Price = *upd.Price
	}
	if upd.SeatsTotal != nil {
		f.SeatsTotal = *upd.SeatsTotal
	}
	if upd.SeatsAvailable != nil {
		f.SeatsAvailable = *upd.SeatsAvailable
	}
	return f, true
}

func (s *MemStore) Delete(ctx context.Context, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flights[id]; !ok {
		return false
	}
	delete(s.flights, id)
	return true
}

func (s *MemStore) Search(ctx context.Context, origin, dest, date string) []*Flight {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Flight
	for _, f := range s.flights {
		match := true
		if origin != "" && f.OriginID != origin {
			match = false
		}
		if dest != "" && f.DestinationID != dest {
			match = false
		}
		if date != "" {
			d := f.DepartureTime.Format("2006-01-02")
			if d != date {
				match = false
			}
		}
		if match {
			out = append(out, f)
		}
	}
	return out
}

func (s *MemStore) SearchConnecting(ctx context.Context, origin, dest, date string) []*ConnectingFlight {
	return nil
}
