package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artemkoru/rtsprec/internal/api"
	"github.com/artemkoru/rtsprec/internal/archive"
	"github.com/artemkoru/rtsprec/internal/auth"
	"github.com/artemkoru/rtsprec/internal/capture"
	"github.com/artemkoru/rtsprec/internal/config"
	"github.com/artemkoru/rtsprec/internal/db"
	"github.com/artemkoru/rtsprec/internal/encoder"
	"github.com/artemkoru/rtsprec/internal/hls"
	"github.com/artemkoru/rtsprec/internal/motion"
	"github.com/artemkoru/rtsprec/internal/storage"
)

func main() {
	configPath := flag.String("config", "/opt/rtsprec/config.yaml", "Path to config file")
	flag.Parse()

	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting RTSP surveillance system")

	// Load configuration
	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("database initialized")

	// Initialize config store
	configStore, err := config.NewConfigStore(database.Conn(), "/opt/rtsprec/.encryption.key")
	if err != nil {
		logger.Error("failed to initialize config store", "error", err)
		os.Exit(1)
	}

	// Initialize S3 uploader
	s3Uploader, err := storage.NewS3Uploader(
		cfg.AWS.Region,
		cfg.AWS.S3Bucket,
		cfg.AWS.S3PathPrefix,
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
	)
	if err != nil {
		logger.Error("failed to initialize S3 uploader", "error", err)
		os.Exit(1)
	}

	logger.Info("S3 uploader initialized")

	// Initialize archive manager
	archiveManager := archive.NewArchiveManager(database.Conn())

	// Initialize authentication
	authManager := auth.NewAuthManager(24 * time.Hour)
	if cfg.Auth.Method == "basic" {
		if err := authManager.EnableBasicAuth(cfg.Auth.BasicUsername, cfg.Auth.BasicPassword); err != nil {
			logger.Error("failed to enable basic auth", "error", err)
		}
	} else if cfg.Auth.Method == "oidc" {
		authManager.EnableOIDC(cfg.Auth.OIDCProvider, cfg.Auth.OIDCClientID, cfg.Auth.OIDCSecret)
	}

	// Initialize API server
	apiServer := api.NewAPIServer(database, configStore, archiveManager, s3Uploader)

	// Initialize HLS handler
	hlsHandler := api.NewHLSHandler()

	// Setup HTTP server
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/cameras", apiServer.HandleGetCameras)
	mux.HandleFunc("POST /api/cameras", apiServer.HandleAddCamera)
	mux.HandleFunc("/api/cameras/*/status", apiServer.HandleCameraStatus)
	mux.HandleFunc("/api/config/*", apiServer.HandleGetConfig)
	mux.HandleFunc("POST /api/config/*", apiServer.HandleUpdateConfig)
	mux.HandleFunc("/api/archive", apiServer.HandleTimeline)
	mux.HandleFunc("/api/events", apiServer.HandleGetEvents)
	mux.HandleFunc("/api/diagnostics", apiServer.HandleDiagnostics)

	// HLS routes
	mux.HandleFunc("/hls/*/playlist.m3u8", hlsHandler.HandlePlaylist)
	mux.HandleFunc("/hls/*/*", hlsHandler.HandleSegment)

	// Auth routes
	mux.HandleFunc("POST /api/auth/login", authManager.HandleBasicLogin)
	mux.HandleFunc("POST /api/auth/logout", handleLogout)

	// Protected API with auth middleware
	protectedMux := authManager.AuthMiddleware(mux)

	// Serve static files (React app)
	staticHandler := http.FileServer(http.Dir("./static"))
	mux.Handle("/", staticHandler)

	// Setup HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      protectedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start camera capture goroutines
	go startCameraCapture(cfg, database, archiveManager, s3Uploader, hlsHandler)

	// Start server
	logger.Info("starting HTTP server", "port", cfg.Server.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("shutdown complete")
}

func startCameraCapture(cfg *config.Config, database *db.DB, archiveManager *archive.ArchiveManager, s3Uploader *storage.S3Uploader, hlsHandler *api.HLSHandler) {
	logger := slog.Default()

	rows, err := database.Conn().Query(`SELECT id, name, rtsp_url, username, password, motion_sensitivity, motion_timeout, video_duration_minutes, screenshot_interval_minutes FROM cameras WHERE enabled = true`)
	if err != nil {
		logger.Error("failed to fetch cameras", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, rtspURL, username, password string
		var sensitivity, timeout, videoDuration, screenshotInterval int

		if err := rows.Scan(&id, &name, &rtspURL, &username, &password, &sensitivity, &timeout, &videoDuration, &screenshotInterval); err != nil {
			logger.Error("failed to scan camera", "error", err)
			continue
		}

		go captureCamera(id, name, rtspURL, username, password, sensitivity, timeout, videoDuration, screenshotInterval, database, archiveManager, s3Uploader, hlsHandler)
	}
}

func captureCamera(cameraID, name, rtspURL, username, password string, sensitivity, timeout, videoDuration, screenshotInterval int, database *db.DB, archiveManager *archive.ArchiveManager, s3Uploader *storage.S3Uploader, hlsHandler *api.HLSHandler) {
	logger := slog.Default().With("camera_id", cameraID, "camera_name", name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// RTSP reader
	rtspReader := capture.NewRTSPReader(cameraID, rtspURL, username, password, 100)
	rtspReader.Start(ctx)
	defer rtspReader.Stop()

	// Motion detector
	motionDetector := motion.NewMotionDetector(cameraID, database, sensitivity, 15, timeout)
	motionDetector.Start(ctx, rtspReader.FrameChannel())

	// Video encoder
	videoEncoder := encoder.NewVideoEncoder(cameraID, videoDuration, "/var/cache/rtsprec/video")
	videoEncoder.Start(ctx, rtspReader.FrameChannel())

	// HLS generator
	hlsGen := hls.NewHLSGenerator(cameraID, cfg.HLS.OutputPath, cfg.HLS.SegmentDuration, cfg.HLS.SegmentsPerList)
	if err := hlsGen.Init(); err != nil {
		logger.Error("failed to initialize HLS generator", "error", err)
		return
	}
	hlsHandler.RegisterGenerator(cameraID, hlsGen)

	// Screenshot manager (if enabled)
	var screenshotMgr *capture.ScreenshotManager
	if screenshotInterval > 0 {
		screenshotMgr = capture.NewScreenshotManager(cameraID, screenshotInterval, "/tmp", s3Uploader)
		screenshotMgr.Start(ctx, rtspReader.FrameChannel())
		defer screenshotMgr.Stop()
	}

	logger.Info("camera capture started")

	// Handle video chunks
	go func() {
		for chunk := range videoEncoder.ChunkChannel() {
			if s3Key, err := s3Uploader.UploadChunk(ctx, chunk.FilePath, cameraID); err != nil {
				logger.Error("failed to upload chunk", "error", err)
			} else {
				archiveManager.IndexVideoFile(cameraID, s3Key, chunk.FilePath, "h264", chunk.StartTime, chunk.EndTime, chunk.FileSize)
			}
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("camera capture stopped")
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "logged out"}`)
}
