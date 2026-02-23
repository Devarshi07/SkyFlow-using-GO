package payments

type CreateIntentRequest struct {
	Amount     int64  `json:"amount"` // in cents
	Currency   string `json:"currency"`
	CustomerID string `json:"customer_id,omitempty"`
}

type IntentResponse struct {
	ClientSecret    string `json:"client_secret"`
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
}

type PaymentDetails struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type RefundRequest struct {
	Amount *int64 `json:"amount,omitempty"` // partial refund in cents
}

type PaymentMethod struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type CheckoutSessionStatus struct {
	Status          string `json:"status"`
	PaymentStatus   string `json:"payment_status"`
	PaymentIntentID string `json:"payment_intent_id,omitempty"`
}

type CheckoutSessionResponse struct {
	SessionID       string `json:"session_id"`
	CheckoutURL     string `json:"checkout_url"`
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
}
