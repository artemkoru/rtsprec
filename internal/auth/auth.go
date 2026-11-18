package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SessionToken struct {
	Token     string
	UserID    string
	Method    string // "oidc" or "basic"
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AuthManager struct {
	sessions          map[string]*SessionToken
	sessionsMu        sync.RWMutex
	sessionExpiration time.Duration
	logger            *slog.Logger

	// Basic Auth
	basicUsername string
	basicPassword string
	basicEnabled  bool

	// OIDC
	oidcEnabled  bool
	oidcProvider string
	oidcClientID string
	oidcSecret   string
}

func NewAuthManager(sessionExpiration time.Duration) *AuthManager {
	am := &AuthManager{
		sessions:          make(map[string]*SessionToken),
		sessionExpiration: sessionExpiration,
		logger:            slog.Default(),
	}

	// Start background goroutine to clean expired sessions
	go am.cleanupExpiredSessions()

	return am
}

// Basic Auth configuration
func (am *AuthManager) EnableBasicAuth(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	am.basicUsername = username
	am.basicPassword = string(hashedPassword)
	am.basicEnabled = true
	am.logger.Info("basic auth enabled")
	return nil
}

// OIDC configuration
func (am *AuthManager) EnableOIDC(provider, clientID, secret string) {
	am.oidcEnabled = true
	am.oidcProvider = provider
	am.oidcClientID = clientID
	am.oidcSecret = secret
	am.logger.Info("OIDC enabled", "provider", provider)
}

// Authenticate with basic auth
func (am *AuthManager) AuthenticateBasic(username, password string) (string, error) {
	if !am.basicEnabled {
		return "", fmt.Errorf("basic auth not enabled")
	}

	if username != am.basicUsername {
		return "", fmt.Errorf("invalid username")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(am.basicPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	return am.createSession(username, "basic"), nil
}

// Authenticate with OIDC (simplified - real implementation would use oauth2 library)
func (am *AuthManager) AuthenticateOIDC(userID, email string) (string, error) {
	if !am.oidcEnabled {
		return "", fmt.Errorf("OIDC not enabled")
	}

	return am.createSession(email, "oidc"), nil
}

// Create session token
func (am *AuthManager) createSession(userID, method string) string {
	token := generateToken()
	session := &SessionToken{
		Token:     token,
		UserID:    userID,
		Method:    method,
		ExpiresAt: time.Now().Add(am.sessionExpiration),
		CreatedAt: time.Now(),
	}

	am.sessionsMu.Lock()
	am.sessions[token] = session
	am.sessionsMu.Unlock()

	am.logger.Info("session created", "user_id", userID, "method", method)
	return token
}

// Validate session token
func (am *AuthManager) ValidateSession(token string) (*SessionToken, error) {
	am.sessionsMu.RLock()
	session, exists := am.sessions[token]
	am.sessionsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		am.sessionsMu.Lock()
		delete(am.sessions, token)
		am.sessionsMu.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// Logout
func (am *AuthManager) Logout(token string) {
	am.sessionsMu.Lock()
	delete(am.sessions, token)
	am.sessionsMu.Unlock()

	am.logger.Info("session terminated", "token", token[:8])
}

// Middleware for HTTP handlers
func (am *AuthManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, "unauthorized: missing session token", http.StatusUnauthorized)
			return
		}

		session, err := am.ValidateSession(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		ctx = context.WithValue(ctx, "auth_method", session.Method)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Extract token from request
func extractToken(r *http.Request) string {
	// Try Authorization header first
	auth := r.Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	// Try cookie
	if cookie, err := r.Cookie("session_token"); err == nil {
		return cookie.Value
	}

	// Try query parameter (for HLS compatibility)
	return r.URL.Query().Get("token")
}

// Generate random token
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Cleanup expired sessions periodically
func (am *AuthManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		am.sessionsMu.Lock()
		now := time.Now()
		for token, session := range am.sessions {
			if now.After(session.ExpiresAt) {
				delete(am.sessions, token)
			}
		}
		am.sessionsMu.Unlock()
	}
}

// Handler for basic auth login
func (am *AuthManager) HandleBasicLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := am.AuthenticateBasic(req.Username, req.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(am.sessionExpiration),
		HttpOnly: true,
		Secure:   true, // Set to false for dev, true for production with HTTPS
		SameSite: http.SameSiteLax,
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"token": "%s"}`, token)
}

// Placeholder for json decoder
func NewDecoder(r interface{}) interface{} {
	return r
}
