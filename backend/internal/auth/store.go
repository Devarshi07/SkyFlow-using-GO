package auth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MemStore struct {
	mu      sync.RWMutex
	users   map[string]*User
	byEmail map[string]*User
	tokens  map[string]string
}

func NewStore() *MemStore {
	return &MemStore{
		users:   make(map[string]*User),
		byEmail: make(map[string]*User),
		tokens:  make(map[string]string),
	}
}

func (s *MemStore) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byEmail[email]; ok {
		return nil, ErrEmailExists
	}
	u := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.users[u.ID] = u
	s.byEmail[email] = u
	return u, nil
}

func (s *MemStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *MemStore) GetByID(ctx context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *MemStore) UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if req.FullName != nil {
		u.FullName = *req.FullName
	}
	if req.Phone != nil {
		u.Phone = *req.Phone
	}
	if req.DateOfBirth != nil {
		u.DateOfBirth = *req.DateOfBirth
	}
	if req.Gender != nil {
		u.Gender = *req.Gender
	}
	if req.Address != nil {
		u.Address = *req.Address
	}
	return u, nil
}

func (s *MemStore) UpsertGoogleUser(ctx context.Context, email, fullName string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.byEmail[email]; ok {
		if u.FullName == "" {
			u.FullName = fullName
		}
		return u, nil
	}
	u := &User{
		ID:        uuid.New().String(),
		Email:     email,
		FullName:  fullName,
		CreatedAt: time.Now(),
	}
	s.users[u.ID] = u
	s.byEmail[email] = u
	return u, nil
}

func (s *MemStore) SaveRefreshToken(ctx context.Context, token, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = userID
}

func (s *MemStore) GetUserByRefreshToken(ctx context.Context, token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.tokens[token]
	return id, ok
}

func (s *MemStore) RevokeRefreshToken(ctx context.Context, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
}

func (s *MemStore) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return ErrUserNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

func (s *MemStore) SaveResetToken(ctx context.Context, userID, token string, ttlMinutes int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens["reset:"+token] = userID
	return nil
}

func (s *MemStore) GetUserByResetToken(ctx context.Context, token string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	uid, ok := s.tokens["reset:"+token]
	if !ok {
		return "", ErrUserNotFound
	}
	return uid, nil
}

func (s *MemStore) MarkResetTokenUsed(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, "reset:"+token)
	return nil
}

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
