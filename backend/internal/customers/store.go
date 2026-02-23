package customers

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemStore struct {
	mu        sync.RWMutex
	customers map[string]*Customer
	payments  map[string][]PaymentHistoryItem
}

func NewStore() *MemStore {
	return &MemStore{
		customers: make(map[string]*Customer),
		payments:  make(map[string][]PaymentHistoryItem),
	}
}

func (s *MemStore) Create(ctx context.Context, c *Customer) (*Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	s.customers[c.ID] = c
	s.payments[c.ID] = nil
	return c, nil
}

func (s *MemStore) GetByID(ctx context.Context, id string) (*Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.customers[id]
	return c, ok
}

func (s *MemStore) AddPayment(ctx context.Context, customerID string, p PaymentHistoryItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[customerID] = append(s.payments[customerID], p)
}

func (s *MemStore) GetPayments(ctx context.Context, customerID string) []PaymentHistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]PaymentHistoryItem(nil), s.payments[customerID]...)
}
