package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// AuthRepository implements domain.AuthRepository using SQLite for users and in-memory storage for sessions.
type AuthRepository struct {
	db       *sql.DB
	sessions map[string]*domain.Session
	mu       sync.RWMutex
}

// NewAuthRepository creates a new SQLite auth repository.
func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		db:       db,
		sessions: make(map[string]*domain.Session),
	}
}

// GetUser returns a user by ID.
func (r *AuthRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, method, created_at, updated_at
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Method, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "User", ID: userID}
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user and returns its ID.
func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, method)
		VALUES (?, ?, ?, ?)
	`, user.ID, user.Username, user.Email, user.Method)

	if err != nil {
		return "", fmt.Errorf("insert failed: %w", err)
	}

	return user.ID, nil
}

// GetSession returns a session by token.
func (r *AuthRepository) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[token]
	if !exists {
		return nil, &domain.NotFoundError{Resource: "Session", ID: token}
	}

	return session, nil
}

// CreateSession creates a new session and returns the token.
func (r *AuthRepository) CreateSession(ctx context.Context, session *domain.Session) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.Token] = session
	return session.Token, nil
}

// DeleteSession deletes a session by token.
func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[token]; !exists {
		return &domain.NotFoundError{Resource: "Session", ID: token}
	}

	delete(r.sessions, token)
	return nil
}

// ValidateSession checks if a session is valid (exists and not expired).
func (r *AuthRepository) ValidateSession(ctx context.Context, token string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[token]
	if !exists {
		return false, nil
	}

	if time.Now().After(session.ExpiresAt) {
		// Session is expired - remove it
		delete(r.sessions, token)
		return false, nil
	}

	return true, nil
}

// CleanupExpiredSessions removes expired sessions (call periodically from a background goroutine).
func (r *AuthRepository) CleanupExpiredSessions() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for token, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, token)
		}
	}
}
