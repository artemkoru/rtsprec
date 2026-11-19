// Package app contains application initialization, configuration, and wiring.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// App represents the application with all its dependencies.
type App struct {
	server   *http.Server
	db       *sql.DB
	services *Services
	config   *Config
	logger   *slog.Logger
}

// Services holds all service instances for dependency injection.
type Services struct {
	Camera   interface{} // CameraService
	Motion   interface{} // MotionService
	Video    interface{} // VideoService
	Archive  interface{} // ArchiveService
	Auth     interface{} // AuthService
}

// New creates and initializes a new application.
func New() (*App, error) {
	logger := slog.Default()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := InitDB(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services and repositories
	services, err := SetupDependencies(db, logger)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to setup services: %w", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Register middleware
	handler := SetupMiddleware(mux, logger)

	// Register routes
	RegisterRoutes(mux, services, logger)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		server:   server,
		db:       db,
		services: services,
		config:   config,
		logger:   logger,
	}, nil
}

// Run starts the application and blocks until shutdown.
func (a *App) Run() error {
	a.logger.Info("starting server", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down server")

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown error", "error", err)
		return err
	}

	if err := a.db.Close(); err != nil {
		a.logger.Error("database close error", "error", err)
		return err
	}

	a.logger.Info("shutdown complete")
	return nil
}
