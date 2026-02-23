package auth

import "context"

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	UpsertGoogleUser(ctx context.Context, email, fullName string) (*User, error)
	SaveRefreshToken(ctx context.Context, token, userID string)
	GetUserByRefreshToken(ctx context.Context, token string) (string, bool)
	RevokeRefreshToken(ctx context.Context, token string)
	// Password reset
	SaveResetToken(ctx context.Context, userID, token string, ttlMinutes int) error
	GetUserByResetToken(ctx context.Context, token string) (string, error)
	MarkResetTokenUsed(ctx context.Context, token string) error
}
