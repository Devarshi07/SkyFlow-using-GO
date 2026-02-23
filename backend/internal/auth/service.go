package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
	jwtSecret  = "skyflow-demo-secret-change-in-production"
)

type claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type PasswordResetPublisher interface {
	PublishPasswordReset(ctx context.Context, email, resetLink string)
}

type Service struct {
	store          Store
	resetPublisher PasswordResetPublisher
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) SetResetPublisher(p PasswordResetPublisher) {
	s.resetPublisher = p
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, *apperrors.AppError) {
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.BadRequest("email and password required")
	}
	u, err := s.store.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, apperrors.Unauthorized("invalid credentials")
		}
		return nil, apperrors.Internal(err)
	}
	if !CheckPassword(u.PasswordHash, req.Password) {
		return nil, apperrors.Unauthorized("invalid credentials")
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*TokenResponse, *apperrors.AppError) {
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.BadRequest("email and password required")
	}
	if len(req.Password) < 6 {
		return nil, apperrors.Validation(map[string]string{"password": "must be at least 6 characters"})
	}
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	u, err := s.store.CreateUser(ctx, req.Email, hash)
	if err != nil {
		if err == ErrEmailExists {
			return nil, apperrors.New(apperrors.CodeConflict, "email already registered", 409)
		}
		return nil, apperrors.Internal(err)
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, *apperrors.AppError) {
	if refreshToken == "" {
		return nil, apperrors.BadRequest("refresh_token required")
	}
	userID, ok := s.store.GetUserByRefreshToken(ctx, refreshToken)
	if !ok {
		return nil, apperrors.Unauthorized("invalid or expired refresh token")
	}
	u, err := s.store.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid refresh token")
	}
	s.store.RevokeRefreshToken(ctx, refreshToken)
	return s.issueTokens(ctx, u)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) *apperrors.AppError {
	if refreshToken != "" {
		s.store.RevokeRefreshToken(ctx, refreshToken)
	}
	return nil
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*User, *apperrors.AppError) {
	u, err := s.store.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NotFound("user")
	}
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*User, *apperrors.AppError) {
	u, err := s.store.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return u, nil
}

func (s *Service) GoogleLogin(ctx context.Context, code, redirectURI string) (*TokenResponse, *apperrors.AppError) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, apperrors.BadRequest("Google OAuth not configured")
	}

	// Exchange code for tokens
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		appErr := apperrors.Internal(fmt.Errorf("google token exchange: %w", err))
		appErr.Details = map[string]string{"cause": "Failed to contact Google OAuth server: " + err.Error()}
		return nil, appErr
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		appErr := apperrors.Unauthorized("google auth: invalid response")
		appErr.Details = map[string]string{"cause": "Could not parse Google token response"}
		return nil, appErr
	}
	if tokenResp.AccessToken == "" {
		msg := tokenResp.Error
		if tokenResp.ErrorDescription != "" {
			msg = tokenResp.ErrorDescription
		}
		if msg == "" {
			msg = "token exchange failed - check redirect_uri matches Google Console"
		}
		appErr := apperrors.Unauthorized("Google auth: " + msg)
		appErr.Details = map[string]string{"cause": msg}
		return nil, appErr
	}

	// Fetch user info
	infoReq, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	infoReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	infoResp, err := http.DefaultClient.Do(infoReq)
	if err != nil {
		appErr := apperrors.Internal(fmt.Errorf("google userinfo: %w", err))
		appErr.Details = map[string]string{"cause": "Failed to fetch Google user info: " + err.Error()}
		return nil, appErr
	}
	defer infoResp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(infoResp.Body).Decode(&userInfo); err != nil || userInfo.Email == "" {
		appErr := apperrors.Unauthorized("failed to get google user info")
		appErr.Details = map[string]string{"cause": "Google returned empty or invalid user info"}
		return nil, appErr
	}

	u, err2 := s.store.UpsertGoogleUser(ctx, userInfo.Email, userInfo.Name)
	if err2 != nil {
		appErr := apperrors.Internal(err2)
		appErr.Details = map[string]string{
			"cause": "Database error during user upsert: " + err2.Error(),
			"hint":  "Ensure migrations are applied (006_google_oauth_schema.sql)",
		}
		return nil, appErr
	}
	return s.issueTokens(ctx, u)
}

func (s *Service) ForgotPassword(ctx context.Context, email string) *apperrors.AppError {
	if email == "" {
		return apperrors.BadRequest("email required")
	}
	u, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate a unique token
	token := uuid.New().String()
	if err := s.store.SaveResetToken(ctx, u.ID, token, 30); err != nil {
		return apperrors.Internal(err)
	}

	// Build reset link
	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:5173"
	}
	resetLink := frontend + "/reset-password?token=" + token

	// Publish reset email via RabbitMQ
	if s.resetPublisher != nil {
		s.resetPublisher.PublishPasswordReset(ctx, u.Email, resetLink)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) *apperrors.AppError {
	if token == "" || newPassword == "" {
		return apperrors.BadRequest("token and new_password required")
	}
	if len(newPassword) < 6 {
		return apperrors.Validation(map[string]string{"password": "must be at least 6 characters"})
	}

	userID, err := s.store.GetUserByResetToken(ctx, token)
	if err != nil {
		return apperrors.BadRequest("invalid or expired reset token")
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return apperrors.Internal(err)
	}

	if err := s.store.UpdatePassword(ctx, userID, hash); err != nil {
		return apperrors.Internal(err)
	}

	_ = s.store.MarkResetTokenUsed(ctx, token)
	return nil
}

// ParseToken validates a JWT and returns claims
func ParseToken(tokenStr string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}

func (s *Service) issueTokens(ctx context.Context, u *User) (*TokenResponse, *apperrors.AppError) {
	now := time.Now()
	accessClaims := &claims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	refreshToken := uuid.New().String()
	s.store.SaveRefreshToken(ctx, refreshToken, u.ID)
	return &TokenResponse{
		AccessToken:  at,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}
