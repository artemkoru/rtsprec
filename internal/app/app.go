// Package app contains application initialization, configuration, and wiring.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/artemkoru/rtsprec/internal/api"
	"github.com/artemkoru/rtsprec/internal/archive"
	"github.com/artemkoru/rtsprec/internal/auth"
	"github.com/artemkoru/rtsprec/internal/config"
	"github.com/artemkoru/rtsprec/internal/db"
	"github.com/artemkoru/rtsprec/internal/storage"
)

// App represents the application with all its dependencies.
type App struct {
	server         *http.Server
	db             *db.DB
	config         *Config
	logger         *slog.Logger
	newS3Uploader  func(region, bucket, pathPrefix, accessKey, secretKey string) (*storage.S3Uploader, error)
	archiveManager *archive.ArchiveManager
	s3Uploader     *storage.S3Uploader
	authManager    *auth.AuthManager
	apiServer      *api.APIServer
	hlsHandler     *api.HLSHandler

	// ctx is cancelled by Shutdown and stops camera capture goroutines
	ctx    context.Context
	cancel context.CancelFunc
}

// New loads configuration. It does not open the database or any other resource.
func New() (*App, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		config:        config,
		logger:        slog.Default(),
		newS3Uploader: storage.NewS3Uploader,
	}, nil
}

// Init opens the database and builds all services, routes and the HTTP server.
// ctx bounds the lifetime of background work started by the application.
func (a *App) Init(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	database, err := db.New(a.config.DatabasePath)
	if err != nil {
		a.cancel()
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = database

	configStore, err := config.NewConfigStore(database.Conn(), a.config.EncryptionKeyPath)
	if err != nil {
		a.closeOnInitError()
		return fmt.Errorf("failed to initialize config store: %w", err)
	}

	s3Uploader, err := a.newS3Uploader(
		a.config.S3Region,
		a.config.S3Bucket,
		a.config.S3PathPrefix,
		a.config.S3AccessKeyID,
		a.config.S3SecretKey,
	)
	if err != nil {
		a.closeOnInitError()
		return fmt.Errorf("failed to initialize S3 uploader: %w", err)
	}
	a.s3Uploader = s3Uploader

	a.archiveManager = archive.NewArchiveManager(database.Conn())

	a.authManager = auth.NewAuthManager(24 * time.Hour)
	switch a.config.AuthMethod {
	case "basic":
		if err := a.authManager.EnableBasicAuth(a.config.BasicUsername, a.config.BasicPassword); err != nil {
			a.logger.Error("failed to enable basic auth", "error", err)
		}
	case "oidc":
		a.authManager.EnableOIDC(a.config.OIDCProvider, a.config.OIDCClientID, a.config.OIDCSecret)
	}

	a.apiServer = api.NewAPIServer(database, configStore, a.archiveManager, s3Uploader)
	a.hlsHandler = api.NewHLSHandler()

	// Public routes (login, logout) bypass authentication; everything else requires a session
	outer := http.NewServeMux()
	registerPublicRoutes(outer, a)

	mux := http.NewServeMux()
	registerRoutes(mux, a)
	outer.Handle("/", a.authManager.AuthMiddleware(mux))

	// Logging, recovery and CORS wrap everything
	handler := SetupMiddleware(outer, a.logger)

	a.server = &http.Server{
		Addr:         ":" + a.config.HTTPPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// closeOnInitError releases resources opened so far when Init fails.
func (a *App) closeOnInitError() {
	a.cancel()
	if a.db != nil {
		a.db.Close()
	}
}

// Run starts camera capture and the HTTP server. It blocks until Shutdown is called
// or the server fails.
func (a *App) Run() error {
	if a.server == nil {
		return fmt.Errorf("Run called before Init")
	}

	a.StartCameras()

	a.logger.Info("starting HTTP server", "addr", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown stops camera capture, gracefully shuts down the HTTP server and closes the database.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down")

	if a.cancel != nil {
		a.cancel()
	}

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Error("shutdown error", "error", err)
			return err
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("database close error", "error", err)
			return err
		}
	}

	a.logger.Info("shutdown complete")
	return nil
}
