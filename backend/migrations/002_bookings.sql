CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    seats INT NOT NULL DEFAULT 1,
    passenger_name VARCHAR(255) NOT NULL,
    passenger_email VARCHAR(255) NOT NULL,
    passenger_phone VARCHAR(50),
    payment_intent_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_flight ON bookings(flight_id);
CREATE INDEX IF NOT EXISTS idx_bookings_payment ON bookings(payment_intent_id);
