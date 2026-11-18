package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// VideoRepository implements domain.VideoRepository using SQLite for metadata and local storage for chunks.
type VideoRepository struct {
	db *sql.DB
}

// NewVideoRepository creates a new SQLite video repository.
func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// FindByID returns video metadata by ID.
func (r *VideoRepository) FindByID(ctx context.Context, id string) (*domain.VideoFile, error) {
	var v domain.VideoFile
	err := r.db.QueryRowContext(ctx, `
		SELECT id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec, resolution, uploaded_at
		FROM video_metadata
		WHERE id = ?
	`, id).Scan(&v.ID, &v.CameraID, &v.S3Path, &v.FileName, &v.StartTime, &v.EndTime,
		&v.DurationSeconds, &v.FileSizeBytes, &v.Codec, &v.Resolution, &v.UploadedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "VideoFile", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &v, nil
}

// FindByCameraID returns video files for a camera with pagination.
func (r *VideoRepository) FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.VideoFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec, resolution, uploaded_at
		FROM video_metadata
		WHERE camera_id = ?
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, cameraID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var videos []domain.VideoFile
	for rows.Next() {
		var v domain.VideoFile
		if err := rows.Scan(&v.ID, &v.CameraID, &v.S3Path, &v.FileName, &v.StartTime, &v.EndTime,
			&v.DurationSeconds, &v.FileSizeBytes, &v.Codec, &v.Resolution, &v.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		videos = append(videos, v)
	}

	return videos, rows.Err()
}

// FindByDateRange returns video files within a date range for a camera.
func (r *VideoRepository) FindByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.VideoFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec, resolution, uploaded_at
		FROM video_metadata
		WHERE camera_id = ? AND start_time >= ? AND end_time <= ?
		ORDER BY start_time ASC
	`, cameraID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var videos []domain.VideoFile
	for rows.Next() {
		var v domain.VideoFile
		if err := rows.Scan(&v.ID, &v.CameraID, &v.S3Path, &v.FileName, &v.StartTime, &v.EndTime,
			&v.DurationSeconds, &v.FileSizeBytes, &v.Codec, &v.Resolution, &v.UploadedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		videos = append(videos, v)
	}

	return videos, rows.Err()
}

// Create creates a new video metadata record and returns its ID.
func (r *VideoRepository) Create(ctx context.Context, video *domain.VideoFile) (string, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO video_metadata (id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec, resolution)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, video.ID, video.CameraID, video.S3Path, video.FileName, video.StartTime, video.EndTime,
		video.DurationSeconds, video.FileSizeBytes, video.Codec, video.Resolution)

	if err != nil {
		return "", fmt.Errorf("insert failed: %w", err)
	}

	return video.ID, nil
}

// UpdateMetadata updates video metadata (codec, resolution, etc).
func (r *VideoRepository) UpdateMetadata(ctx context.Context, video *domain.VideoFile) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE video_metadata
		SET duration_seconds = ?, file_size_bytes = ?, codec = ?, resolution = ?
		WHERE id = ?
	`, video.DurationSeconds, video.FileSizeBytes, video.Codec, video.Resolution, video.ID)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "VideoFile", ID: video.ID}
	}

	return nil
}

// Save stores a video chunk (implemented as local storage placeholder).
func (r *VideoRepository) Save(ctx context.Context, chunk *domain.VideoChunk, data []byte) (string, error) {
	// This is a placeholder - actual implementation would handle local disk storage
	return chunk.FilePath, nil
}

// Get retrieves a video chunk data (implemented as local storage placeholder).
func (r *VideoRepository) Get(ctx context.Context, path string) ([]byte, error) {
	// This is a placeholder - actual implementation would retrieve from local disk
	return nil, nil
}

// Delete removes a video and its associated storage.
func (r *VideoRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM video_metadata WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "VideoFile", ID: id}
	}

	return nil
}
