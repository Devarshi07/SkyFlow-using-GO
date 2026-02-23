package bookings

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skyflow/skyflow/internal/shared/postgres"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, b *Booking) (*Booking, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO bookings (user_id, flight_id, seats, passenger_name, passenger_email, passenger_phone, payment_intent_id, status, amount)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id::text, user_id::text, flight_id::text, seats, COALESCE(passenger_name,''), COALESCE(passenger_email,''), COALESCE(passenger_phone,''), COALESCE(payment_intent_id,''), status, COALESCE(amount,0), created_at`,
		b.UserID, b.FlightID, b.Seats, b.PassengerName, b.PassengerEmail, nullIfEmpty(b.PassengerPhone), nullIfEmpty(b.PaymentIntentID), b.Status, b.Amount,
	).Scan(&b.ID, &b.UserID, &b.FlightID, &b.Seats, &b.PassengerName, &b.PassengerEmail, &b.PassengerPhone, &b.PaymentIntentID, &b.Status, &b.Amount, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

const selectBookingCols = `id::text, user_id::text, flight_id::text, seats, passenger_name, passenger_email, COALESCE(passenger_phone,''), COALESCE(payment_intent_id,''), status, COALESCE(amount,0), created_at`

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Booking, error) {
	return s.scanBooking(ctx,
		`SELECT `+selectBookingCols+` FROM bookings WHERE id::text = $1`, id)
}

func (s *PostgresStore) GetByPaymentIntent(ctx context.Context, paymentIntentID string) (*Booking, error) {
	return s.scanBooking(ctx,
		`SELECT `+selectBookingCols+` FROM bookings WHERE payment_intent_id = $1`, paymentIntentID)
}

func (s *PostgresStore) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE bookings SET status = $2 WHERE id::text = $1`, id, status)
	return err
}

func (s *PostgresStore) UpdatePaymentIntent(ctx context.Context, id, paymentIntentID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE bookings SET payment_intent_id = $2 WHERE id::text = $1`, id, paymentIntentID)
	return err
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID string) ([]*Booking, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+selectBookingCols+` FROM bookings WHERE user_id::text = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.FlightID, &b.Seats, &b.PassengerName, &b.PassengerEmail, &b.PassengerPhone, &b.PaymentIntentID, &b.Status, &b.Amount, &b.CreatedAt); err != nil {
			continue
		}
		out = append(out, &b)
	}
	return out, nil
}

func (s *PostgresStore) scanBooking(ctx context.Context, query string, args ...interface{}) (*Booking, error) {
	var b Booking
	err := s.pool.QueryRow(ctx, query, args...).Scan(&b.ID, &b.UserID, &b.FlightID, &b.Seats, &b.PassengerName, &b.PassengerEmail, &b.PassengerPhone, &b.PaymentIntentID, &b.Status, &b.Amount, &b.CreatedAt)
	if err != nil {
		if postgres.IsNoRows(err) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	return &b, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
