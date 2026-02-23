-- Add amount column to bookings table for storing payment amount
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS amount BIGINT NOT NULL DEFAULT 0;
