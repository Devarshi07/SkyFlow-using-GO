package auth

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	DateOfBirth  string    `json:"date_of_birth,omitempty"`
	Gender       string    `json:"gender,omitempty"`
	Address      string    `json:"address,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UpdateProfileRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	DateOfBirth *string `json:"date_of_birth,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	Address     *string `json:"address,omitempty"`
}

type GoogleAuthRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
