package customers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, c *Customer) (*Customer, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO customers (email, name, stripe_id) VALUES ($1, $2, $3)
		 RETURNING id::text, email, name, stripe_id, created_at`,
		c.Email, c.Name, c.StripeID,
	).Scan(&c.ID, &c.Email, &c.Name, &c.StripeID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Customer, bool) {
	var c Customer
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, email, name, stripe_id, created_at FROM customers WHERE id::text = $1`, id,
	).Scan(&c.ID, &c.Email, &c.Name, &c.StripeID, &c.CreatedAt)
	return &c, err == nil
}

func (s *PostgresStore) AddPayment(ctx context.Context, customerID string, p PaymentHistoryItem) {
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO customer_payments (customer_id, payment_intent_id, amount, currency, status) VALUES ($1::uuid, $2, $3, $4, $5)`,
		customerID, p.ID, p.Amount, p.Currency, p.Status)
}

func (s *PostgresStore) GetPayments(ctx context.Context, customerID string) []PaymentHistoryItem {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, amount, currency, status, created_at::text FROM customer_payments WHERE customer_id::text = $1 ORDER BY created_at DESC`,
		customerID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []PaymentHistoryItem
	for rows.Next() {
		var p PaymentHistoryItem
		var createdAt string
		if err := rows.Scan(&p.ID, &p.Amount, &p.Currency, &p.Status, &createdAt); err != nil {
			continue
		}
		p.CreatedAt = createdAt
		out = append(out, p)
	}
	return out
}
