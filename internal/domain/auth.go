package domain

import "time"

// User represents an authenticated user of the system.
type User struct {
	ID        string
	Username  string
	Email     string
	Method    string    // "basic" or "oidc"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Session represents an authenticated session for a user.
type Session struct {
	ID        string
	UserID    string
	Token     string
	Method    string    // "basic" or "oidc"
	ExpiresAt time.Time
	CreatedAt time.Time
}
