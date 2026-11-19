package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// AuthService interface defines auth service methods.
type AuthService interface {
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (string, error)
	Authenticate(ctx context.Context, session *domain.Session) (string, error)
	ValidateSession(ctx context.Context, token string) (bool, error)
	GetSession(ctx context.Context, token string) (*domain.Session, error)
	Logout(ctx context.Context, token string) error
}

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService AuthService
	logger      *slog.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// HandleLogin authenticates a user and returns a session token.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Username == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "missing username or password")
		return
	}

	session := &domain.Session{
		UserID:    req.Username,
		Method:    "basic",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	token, err := h.authService.Authenticate(r.Context(), session)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("authentication failed: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// HandleLogout logs out a user by deleting their session.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "missing session token")
		return
	}

	if err := h.authService.Logout(r.Context(), token); err != nil {
		h.logger.Debug("logout session not found (already expired or invalid)", "error", err)
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// HandleValidateSession validates a session token.
func (h *AuthHandler) HandleValidateSession(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "missing session token")
		return
	}

	valid, err := h.authService.ValidateSession(r.Context(), token)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("validation failed: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]bool{"valid": valid})
}

// extractToken extracts the session token from the request.
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

	// Try query parameter
	return r.URL.Query().Get("token")
}

// Helper methods
func (h *AuthHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

func (h *AuthHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}
