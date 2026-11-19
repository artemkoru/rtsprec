package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/repository"
)

// VideoService handles business logic for video file operations.
type VideoService struct {
	videoRepo repository.VideoRepository
	logger    *slog.Logger
}

// NewVideoService creates a new video service.
func NewVideoService(videoRepo repository.VideoRepository, logger *slog.Logger) *VideoService {
	return &VideoService{
		videoRepo: videoRepo,
		logger:    logger,
	}
}

// GetFile returns a video file by ID.
func (s *VideoService) GetFile(ctx context.Context, id string) (*domain.VideoFile, error) {
	file, err := s.videoRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get video file", "video_id", id, "error", err)
		return nil, err
	}
	return file, nil
}

// ListFiles returns video files for a camera with pagination.
func (s *VideoService) ListFiles(ctx context.Context, cameraID string, limit, offset int) ([]domain.VideoFile, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	files, err := s.videoRepo.FindByCameraID(ctx, cameraID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list video files", "camera_id", cameraID, "error", err)
		return nil, err
	}
	return files, nil
}

// GetFilesByDateRange returns video files within a date range.
func (s *VideoService) GetFilesByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.VideoFile, error) {
	files, err := s.videoRepo.FindByDateRange(ctx, cameraID, start, end)
	if err != nil {
		s.logger.Error("failed to get video files by date range", "camera_id", cameraID, "error", err)
		return nil, err
	}
	return files, nil
}

// SaveChunk saves a video chunk and indexes it.
func (s *VideoService) SaveChunk(ctx context.Context, chunk *domain.VideoChunk, data []byte) (string, error) {
	path, err := s.videoRepo.Save(ctx, chunk, data)
	if err != nil {
		s.logger.Error("failed to save video chunk", "chunk_id", chunk.ID, "error", err)
		return "", err
	}
	s.logger.Info("video chunk saved", "chunk_id", chunk.ID, "path", path)
	return path, nil
}

// IndexFile indexes a video file in the database.
func (s *VideoService) IndexFile(ctx context.Context, file *domain.VideoFile) (string, error) {
	fileID, err := s.videoRepo.Create(ctx, file)
	if err != nil {
		s.logger.Error("failed to index video file", "s3_path", file.S3Path, "error", err)
		return "", err
	}
	s.logger.Info("video file indexed", "file_id", fileID, "s3_path", file.S3Path)
	return fileID, nil
}

// UpdateFileMetadata updates video file metadata.
func (s *VideoService) UpdateFileMetadata(ctx context.Context, file *domain.VideoFile) error {
	err := s.videoRepo.UpdateMetadata(ctx, file)
	if err != nil {
		s.logger.Error("failed to update video file metadata", "file_id", file.ID, "error", err)
		return err
	}
	s.logger.Info("video file metadata updated", "file_id", file.ID)
	return nil
}

// DeleteFile deletes a video file.
func (s *VideoService) DeleteFile(ctx context.Context, id string) error {
	err := s.videoRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete video file", "file_id", id, "error", err)
		return err
	}
	s.logger.Info("video file deleted", "file_id", id)
	return nil
}
