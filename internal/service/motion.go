package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/repository"
)

// MotionService handles business logic for motion event operations.
type MotionService struct {
	motionRepo repository.MotionRepository
	logger     *slog.Logger
}

// NewMotionService creates a new motion service.
func NewMotionService(motionRepo repository.MotionRepository, logger *slog.Logger) *MotionService {
	return &MotionService{
		motionRepo: motionRepo,
		logger:     logger,
	}
}

// GetEvents returns motion events for a camera with optional pagination.
func (s *MotionService) GetEvents(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	events, err := s.motionRepo.FindByCameraID(ctx, cameraID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get motion events", "camera_id", cameraID, "error", err)
		return nil, err
	}
	return events, nil
}

// GetEventsByDateRange returns motion events within a date range.
func (s *MotionService) GetEventsByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error) {
	events, err := s.motionRepo.FindByDateRange(ctx, cameraID, start, end)
	if err != nil {
		s.logger.Error("failed to get motion events by date range", "camera_id", cameraID, "error", err)
		return nil, err
	}
	return events, nil
}

// CreateEvent creates a new motion event.
func (s *MotionService) CreateEvent(ctx context.Context, event *domain.MotionEvent) (string, error) {
	eventID, err := s.motionRepo.Create(ctx, event)
	if err != nil {
		s.logger.Error("failed to create motion event", "camera_id", event.CameraID, "error", err)
		return "", err
	}
	s.logger.Info("motion event created", "event_id", eventID, "camera_id", event.CameraID)
	return eventID, nil
}

// UpdateEvent updates an existing motion event.
func (s *MotionService) UpdateEvent(ctx context.Context, event *domain.MotionEvent) error {
	err := s.motionRepo.Update(ctx, event)
	if err != nil {
		s.logger.Error("failed to update motion event", "event_id", event.ID, "error", err)
		return err
	}
	s.logger.Info("motion event updated", "event_id", event.ID)
	return nil
}
