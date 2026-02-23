package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skyflow/skyflow/internal/shared/postgres"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	var u User
	var fullName, phone, dob, gender, address sql.NullString
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id::text, email, password_hash, COALESCE(full_name,''), COALESCE(phone,''), COALESCE(date_of_birth::text,''), COALESCE(gender,''), COALESCE(address,''), created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &fullName, &phone, &dob, &gender, &address, &u.CreatedAt)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	u.FullName = nullStr(fullName)
	u.Phone = nullStr(phone)
	u.DateOfBirth = nullStr(dob)
	u.Gender = nullStr(gender)
	u.Address = nullStr(address)
	return &u, nil
}

func (s *PostgresStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.scanUser(ctx, `SELECT id::text, email, COALESCE(password_hash,''), COALESCE(full_name,''), COALESCE(phone,''), COALESCE(date_of_birth::text,''), COALESCE(gender,''), COALESCE(address,''), created_at FROM users WHERE email = $1`, email)
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*User, error) {
	return s.scanUser(ctx, `SELECT id::text, email, COALESCE(password_hash,''), COALESCE(full_name,''), COALESCE(phone,''), COALESCE(date_of_birth::text,''), COALESCE(gender,''), COALESCE(address,''), created_at FROM users WHERE id::text = $1`, id)
}

func (s *PostgresStore) scanUser(ctx context.Context, query string, args ...interface{}) (*User, error) {
	var u User
	var fullName, phone, dob, gender, address sql.NullString
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&fullName, &phone, &dob, &gender, &address,
		&u.CreatedAt,
	)
	if err != nil {
		if postgres.IsNoRows(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.FullName = nullStr(fullName)
	u.Phone = nullStr(phone)
	u.DateOfBirth = nullStr(dob)
	u.Gender = nullStr(gender)
	u.Address = nullStr(address)
	return &u, nil
}

func (s *PostgresStore) UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*User, error) {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET
			full_name = COALESCE($2, full_name),
			phone = COALESCE($3, phone),
			date_of_birth = CASE WHEN $4 = '' THEN date_of_birth ELSE $4::date END,
			gender = COALESCE($5, gender),
			address = COALESCE($6, address)
		 WHERE id::text = $1`,
		id, req.FullName, req.Phone, ptrOrEmpty(req.DateOfBirth), req.Gender, req.Address,
	)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *PostgresStore) UpsertGoogleUser(ctx context.Context, email, fullName string) (*User, error) {
	var u User
	var fullNameVal, phone, dob, gender, address sql.NullString
	// fullName can be empty from Google; use email prefix as fallback for display
	displayName := fullName
	if displayName == "" {
		displayName = email
	}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, full_name, password_hash) VALUES ($1, $2, '')
		 ON CONFLICT (email) DO UPDATE SET full_name = COALESCE(NULLIF(TRIM(users.full_name),''), EXCLUDED.full_name)
		 RETURNING id::text, email, COALESCE(password_hash,''), COALESCE(full_name,''), COALESCE(phone,''), COALESCE(date_of_birth::text,''), COALESCE(gender,''), COALESCE(address,''), created_at`,
		email, displayName,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &fullNameVal, &phone, &dob, &gender, &address, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.FullName = nullStr(fullNameVal)
	if u.FullName == "" {
		u.FullName = displayName
	}
	u.Phone = nullStr(phone)
	u.DateOfBirth = nullStr(dob)
	u.Gender = nullStr(gender)
	u.Address = nullStr(address)
	return &u, nil
}

func (s *PostgresStore) SaveRefreshToken(ctx context.Context, token, userID string) {
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (token, user_id) VALUES ($1, $2::uuid)`,
		token, userID,
	)
}

func (s *PostgresStore) GetUserByRefreshToken(ctx context.Context, token string) (string, bool) {
	var userID string
	err := s.pool.QueryRow(ctx,
		`SELECT user_id::text FROM refresh_tokens WHERE token = $1`,
		token,
	).Scan(&userID)
	return userID, err == nil
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, token string) {
	_, _ = s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
}

func (s *PostgresStore) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id::text = $1`, id, passwordHash)
	return err
}

func (s *PostgresStore) SaveResetToken(ctx context.Context, userID, token string, ttlMinutes int) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1::uuid, $2, NOW() + ($3 || ' minutes')::interval)`,
		userID, token, fmt.Sprintf("%d", ttlMinutes),
	)
	return err
}

func (s *PostgresStore) GetUserByResetToken(ctx context.Context, token string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx,
		`SELECT user_id::text FROM password_reset_tokens WHERE token = $1 AND used = FALSE AND expires_at > NOW()`,
		token,
	).Scan(&userID)
	if err != nil {
		return "", ErrUserNotFound
	}
	return userID, nil
}

func (s *PostgresStore) MarkResetTokenUsed(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `UPDATE password_reset_tokens SET used = TRUE WHERE token = $1`, token)
	return err
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func ptrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
