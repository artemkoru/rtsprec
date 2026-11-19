package service

import (
	"context"
	"log/slog"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/repository"
)

// CameraService handles business logic for camera operations.
type CameraService struct {
	cameraRepo repository.CameraRepository
	logger     *slog.Logger
}

// NewCameraService creates a new camera service.
func NewCameraService(cameraRepo repository.CameraRepository, logger *slog.Logger) *CameraService {
	return &CameraService{
		cameraRepo: cameraRepo,
		logger:     logger,
	}
}

// GetAll returns all cameras.
func (s *CameraService) GetAll(ctx context.Context) ([]domain.Camera, error) {
	cameras, err := s.cameraRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("failed to get all cameras", "error", err)
		return nil, err
	}
	return cameras, nil
}

// GetByID returns a camera by ID.
func (s *CameraService) GetByID(ctx context.Context, id string) (*domain.Camera, error) {
	camera, err := s.cameraRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get camera", "camera_id", id, "error", err)
		return nil, err
	}
	return camera, nil
}

// Create creates a new camera.
func (s *CameraService) Create(ctx context.Context, camera *domain.Camera) (string, error) {
	cameraID, err := s.cameraRepo.Create(ctx, camera)
	if err != nil {
		s.logger.Error("failed to create camera", "name", camera.Name, "error", err)
		return "", err
	}
	s.logger.Info("camera created", "camera_id", cameraID, "name", camera.Name)
	return cameraID, nil
}

// Update updates an existing camera.
func (s *CameraService) Update(ctx context.Context, camera *domain.Camera) error {
	err := s.cameraRepo.Update(ctx, camera)
	if err != nil {
		s.logger.Error("failed to update camera", "camera_id", camera.ID, "error", err)
		return err
	}
	s.logger.Info("camera updated", "camera_id", camera.ID)
	return nil
}

// Delete deletes a camera.
func (s *CameraService) Delete(ctx context.Context, id string) error {
	err := s.cameraRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete camera", "camera_id", id, "error", err)
		return err
	}
	s.logger.Info("camera deleted", "camera_id", id)
	return nil
}
