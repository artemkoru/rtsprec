package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/repository"
)

// ArchiveService handles business logic for archive queries and timeline generation.
type ArchiveService struct {
	archiveRepo repository.ArchiveRepository
	logger      *slog.Logger
}

// NewArchiveService creates a new archive service.
func NewArchiveService(archiveRepo repository.ArchiveRepository, logger *slog.Logger) *ArchiveService {
	return &ArchiveService{
		archiveRepo: archiveRepo,
		logger:      logger,
	}
}

// QueryVideos returns video files matching the query criteria.
func (s *ArchiveService) QueryVideos(ctx context.Context, query *domain.ArchiveQuery) ([]domain.VideoFile, error) {
	files, err := s.archiveRepo.QueryVideos(ctx, query)
	if err != nil {
		s.logger.Error("failed to query videos", "camera_id", query.CameraID, "error", err)
		return nil, err
	}
	return files, nil
}

// QueryEvents returns motion events matching the query criteria.
func (s *ArchiveService) QueryEvents(ctx context.Context, query *domain.ArchiveQuery) ([]domain.MotionEvent, error) {
	events, err := s.archiveRepo.QueryEvents(ctx, query)
	if err != nil {
		s.logger.Error("failed to query events", "camera_id", query.CameraID, "error", err)
		return nil, err
	}
	return events, nil
}

// GenerateTimeline generates a timeline of videos and events for a specific date.
func (s *ArchiveService) GenerateTimeline(ctx context.Context, cameraID string, date time.Time) (*domain.Timeline, error) {
	timeline, err := s.archiveRepo.GenerateTimeline(ctx, cameraID, date)
	if err != nil {
		s.logger.Error("failed to generate timeline", "camera_id", cameraID, "date", date.Format("2006-01-02"), "error", err)
		return nil, err
	}
	s.logger.Info("timeline generated", "camera_id", cameraID, "date", date.Format("2006-01-02"), "items", len(timeline.Items))
	return timeline, nil
}

// Query returns videos and events within a date range.
func (s *ArchiveService) Query(ctx context.Context, cameraID string, start, end time.Time) (*domain.ArchiveQuery, error) {
	return &domain.ArchiveQuery{
		CameraID:  cameraID,
		StartTime: start,
		EndTime:   end,
	}, nil
}
