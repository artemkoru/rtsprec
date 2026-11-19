package s3

import (
	"context"
	"fmt"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/artemkoru/rtsprec/internal/storage"
)

// VideoRepository implements domain.VideoRepository using S3 for video storage.
type VideoRepository struct {
	uploader *storage.S3Uploader
}

// NewVideoRepository creates a new S3 video repository.
func NewVideoRepository(uploader *storage.S3Uploader) *VideoRepository {
	return &VideoRepository{uploader: uploader}
}

// FindByID is not implemented for S3 repository (use SQLite repository instead).
func (r *VideoRepository) FindByID(ctx context.Context, id string) (*domain.VideoFile, error) {
	return nil, fmt.Errorf("FindByID not supported by S3 repository; use SQLite repository")
}

// FindByCameraID is not implemented for S3 repository (use SQLite repository instead).
func (r *VideoRepository) FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.VideoFile, error) {
	return nil, fmt.Errorf("FindByCameraID not supported by S3 repository; use SQLite repository")
}

// FindByDateRange is not implemented for S3 repository (use SQLite repository instead).
func (r *VideoRepository) FindByDateRange(ctx context.Context, cameraID string, start, end interface{}) ([]domain.VideoFile, error) {
	return nil, fmt.Errorf("FindByDateRange not supported by S3 repository; use SQLite repository")
}

// Create is not implemented for S3 repository (use SQLite repository for metadata).
func (r *VideoRepository) Create(ctx context.Context, video *domain.VideoFile) (string, error) {
	return "", fmt.Errorf("Create not supported by S3 repository; use SQLite repository for metadata")
}

// UpdateMetadata is not implemented for S3 repository (use SQLite repository instead).
func (r *VideoRepository) UpdateMetadata(ctx context.Context, video *domain.VideoFile) error {
	return fmt.Errorf("UpdateMetadata not supported by S3 repository; use SQLite repository")
}

// Save uploads a video chunk to S3 and returns the S3 path.
func (r *VideoRepository) Save(ctx context.Context, chunk *domain.VideoChunk, data []byte) (string, error) {
	// Write data to temporary file
	tmpFile := chunk.FilePath

	// For now, write to local storage (actual implementation would use temp file)
	// Then upload to S3
	s3Path := fmt.Sprintf("videos/%s/%s", chunk.CameraID, chunk.ID)

	// Upload to S3 (implementation depends on S3Uploader)
	uploadedPath, err := r.uploader.UploadFile(ctx, tmpFile, s3Path)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return uploadedPath, nil
}

// Get retrieves a video chunk from S3.
func (r *VideoRepository) Get(ctx context.Context, path string) ([]byte, error) {
	// This would download from S3 and return the data
	// Implementation depends on S3Uploader capabilities
	return nil, fmt.Errorf("Get not implemented for S3 repository")
}

// Delete removes a video from S3.
func (r *VideoRepository) Delete(ctx context.Context, id string) error {
	// This would delete from S3
	// Implementation depends on S3Uploader capabilities
	return fmt.Errorf("Delete not implemented for S3 repository")
}
