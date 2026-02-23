package customers

import (
	"context"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
)

type Service struct {
	store CustomerStore
}

func NewService(store CustomerStore) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, req CreateCustomerRequest) (*Customer, *apperrors.AppError) {
	if req.Email == "" {
		return nil, apperrors.BadRequest("email required")
	}
	c := &Customer{Email: req.Email, Name: req.Name}
	out, err := s.store.Create(ctx, c)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return out, nil
}

func (s *Service) GetByID(ctx context.Context, customerID string) (*Customer, *apperrors.AppError) {
	c, ok := s.store.GetByID(ctx, customerID)
	if !ok {
		return nil, apperrors.New(apperrors.CodeCustomerNotFound, "customer not found", 404)
	}
	return c, nil
}

func (s *Service) GetPaymentHistory(ctx context.Context, customerID string) ([]PaymentHistoryItem, *apperrors.AppError) {
	_, ok := s.store.GetByID(ctx, customerID)
	if !ok {
		return nil, apperrors.New(apperrors.CodeCustomerNotFound, "customer not found", 404)
	}
	return s.store.GetPayments(ctx, customerID), nil
}
