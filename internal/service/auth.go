package service

import (
	"context"
	"log/slog"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/repository"
)

// AuthService handles business logic for authentication and sessions.
type AuthService struct {
	authRepo repository.AuthRepository
	logger   *slog.Logger
}

// NewAuthService creates a new auth service.
func NewAuthService(authRepo repository.AuthRepository, logger *slog.Logger) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		logger:   logger,
	}
}

// GetUser returns a user by ID.
func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.authRepo.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user", "user_id", userID, "error", err)
		return nil, err
	}
	return user, nil
}

// CreateUser creates a new user.
func (s *AuthService) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	userID, err := s.authRepo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("failed to create user", "username", user.Username, "error", err)
		return "", err
	}
	s.logger.Info("user created", "user_id", userID, "username", user.Username)
	return userID, nil
}

// Authenticate authenticates a user and creates a session.
func (s *AuthService) Authenticate(ctx context.Context, session *domain.Session) (string, error) {
	token, err := s.authRepo.CreateSession(ctx, session)
	if err != nil {
		s.logger.Error("failed to create session", "user_id", session.UserID, "error", err)
		return "", err
	}
	s.logger.Info("session created", "user_id", session.UserID, "method", session.Method)
	return token, nil
}

// ValidateSession validates a session token.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (bool, error) {
	valid, err := s.authRepo.ValidateSession(ctx, token)
	if err != nil {
		s.logger.Error("failed to validate session", "error", err)
		return false, err
	}
	return valid, nil
}

// GetSession returns a session by token.
func (s *AuthService) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	session, err := s.authRepo.GetSession(ctx, token)
	if err != nil {
		s.logger.Error("failed to get session", "error", err)
		return nil, err
	}
	return session, nil
}

// Logout logs out a user by deleting their session.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	err := s.authRepo.DeleteSession(ctx, token)
	if err != nil {
		s.logger.Error("failed to delete session", "error", err)
		return err
	}
	s.logger.Info("session deleted")
	return nil
}

// CreateSession creates a new authenticated session.
func (s *AuthService) CreateSession(ctx context.Context, session *domain.Session) (string, error) {
	return s.Authenticate(ctx, session)
}
