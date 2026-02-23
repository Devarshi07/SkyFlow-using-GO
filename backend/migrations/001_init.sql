-- SkyFlow PostgreSQL schema
-- Run: psql $DATABASE_URL -f migrations/001_init.sql

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token VARCHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    code VARCHAR(10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS airports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    city_id UUID NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    code VARCHAR(10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_number VARCHAR(20) NOT NULL,
    origin_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    destination_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    departure_time TIMESTAMPTZ NOT NULL,
    arrival_time TIMESTAMPTZ NOT NULL,
    price BIGINT NOT NULL DEFAULT 0,
    seats_total INT NOT NULL DEFAULT 100,
    seats_available INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    stripe_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    payment_intent_id VARCHAR(255),
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'usd',
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_flights_origin ON flights(origin_id);
CREATE INDEX IF NOT EXISTS idx_flights_destination ON flights(destination_id);
CREATE INDEX IF NOT EXISTS idx_flights_departure ON flights(departure_time);
CREATE INDEX IF NOT EXISTS idx_customer_payments_customer ON customer_payments(customer_id);
