package app

import (
	"database/sql"
	"fmt"

	"github.com/artemkoru/rtsprec/internal/db"
)

// InitDB initializes the database connection and runs migrations.
func InitDB(dbPath string) (*sql.DB, error) {
	database, err := db.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Test the connection
	if err := database.Conn().Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return database.Conn(), nil
}
