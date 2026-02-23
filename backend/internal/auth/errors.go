package auth

import (
	"errors"
)

var (
	ErrEmailExists   = errors.New("email already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidCreds  = errors.New("invalid credentials")
	ErrInvalidToken  = errors.New("invalid refresh token")
)
