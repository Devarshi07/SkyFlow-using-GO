package customers

import "context"

type CustomerStore interface {
	Create(ctx context.Context, c *Customer) (*Customer, error)
	GetByID(ctx context.Context, id string) (*Customer, bool)
	AddPayment(ctx context.Context, customerID string, p PaymentHistoryItem)
	GetPayments(ctx context.Context, customerID string) []PaymentHistoryItem
}
