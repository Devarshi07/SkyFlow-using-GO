package customers

import "time"

type Customer struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	StripeID  string    `json:"stripe_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCustomerRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type PaymentHistoryItem struct {
	ID        string `json:"id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
