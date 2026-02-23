package bookings

import "context"

type Store interface {
	Create(ctx context.Context, b *Booking) (*Booking, error)
	GetByID(ctx context.Context, id string) (*Booking, error)
	GetByPaymentIntent(ctx context.Context, paymentIntentID string) (*Booking, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePaymentIntent(ctx context.Context, id, paymentIntentID string) error
	UpdateBooking(ctx context.Context, b *Booking) error
	ListByUser(ctx context.Context, userID string) ([]*Booking, error)
}
