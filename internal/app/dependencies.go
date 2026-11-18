package app

import (
	"database/sql"
	"log/slog"

	"github.com/artemkoru/rtsprec/internal/repository/sqlite"
	"github.com/artemkoru/rtsprec/internal/service"
)

// SetupDependencies wires all repositories and services.
func SetupDependencies(db *sql.DB, logger *slog.Logger) (*Services, error) {
	// Initialize repositories
	cameraRepo := sqlite.NewCameraRepository(db)
	motionRepo := sqlite.NewMotionRepository(db)
	videoRepo := sqlite.NewVideoRepository(db)
	archiveRepo := sqlite.NewArchiveRepository(videoRepo, motionRepo)
	authRepo := sqlite.NewAuthRepository(db)

	// Initialize services
	cameraService := service.NewCameraService(cameraRepo, logger)
	motionService := service.NewMotionService(motionRepo, logger)
	videoService := service.NewVideoService(videoRepo, logger)
	archiveService := service.NewArchiveService(archiveRepo, logger)
	authService := service.NewAuthService(authRepo, logger)

	return &Services{
		Camera:   cameraService,
		Motion:   motionService,
		Video:    videoService,
		Archive:  archiveService,
		Auth:     authService,
	}, nil
}
