-- Ensure users table supports Google OAuth (run if 003 wasn't applied)
-- Safe to run multiple times (idempotent)

ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS gender VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT;

-- Allow NULL password_hash for Google-only users
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
