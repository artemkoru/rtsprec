package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// CameraRepository implements domain.CameraRepository using SQLite.
type CameraRepository struct {
	db *sql.DB
}

// NewCameraRepository creates a new SQLite camera repository.
func NewCameraRepository(db *sql.DB) *CameraRepository {
	return &CameraRepository{db: db}
}

// FindAll returns all cameras.
func (r *CameraRepository) FindAll(ctx context.Context) ([]domain.Camera, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, rtsp_url, username, password, status, motion_sensitivity, motion_timeout,
		       video_duration_minutes, screenshot_interval_minutes, enabled, created_at, updated_at
		FROM cameras
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var cameras []domain.Camera
	for rows.Next() {
		var c domain.Camera
		if err := rows.Scan(&c.ID, &c.Name, &c.RtspURL, &c.Username, &c.Password, &c.Status,
			&c.MotionSensitivity, &c.MotionTimeout, &c.VideoDurationMinutes,
			&c.ScreenshotIntervalMinutes, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		cameras = append(cameras, c)
	}

	return cameras, rows.Err()
}

// FindByID returns a camera by ID.
func (r *CameraRepository) FindByID(ctx context.Context, id string) (*domain.Camera, error) {
	var c domain.Camera
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, rtsp_url, username, password, status, motion_sensitivity, motion_timeout,
		       video_duration_minutes, screenshot_interval_minutes, enabled, created_at, updated_at
		FROM cameras
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.RtspURL, &c.Username, &c.Password, &c.Status,
		&c.MotionSensitivity, &c.MotionTimeout, &c.VideoDurationMinutes,
		&c.ScreenshotIntervalMinutes, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "Camera", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &c, nil
}

// Create creates a new camera and returns its ID.
func (r *CameraRepository) Create(ctx context.Context, camera *domain.Camera) (string, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cameras (id, name, rtsp_url, username, password, motion_sensitivity,
		                      motion_timeout, video_duration_minutes, screenshot_interval_minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, camera.ID, camera.Name, camera.RtspURL, camera.Username, camera.Password,
		camera.MotionSensitivity, camera.MotionTimeout, camera.VideoDurationMinutes,
		camera.ScreenshotIntervalMinutes)

	if err != nil {
		return "", fmt.Errorf("insert failed: %w", err)
	}

	return camera.ID, nil
}

// Update updates an existing camera.
func (r *CameraRepository) Update(ctx context.Context, camera *domain.Camera) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE cameras
		SET name = ?, rtsp_url = ?, username = ?, password = ?, status = ?,
		    motion_sensitivity = ?, motion_timeout = ?, video_duration_minutes = ?,
		    screenshot_interval_minutes = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, camera.Name, camera.RtspURL, camera.Username, camera.Password, camera.Status,
		camera.MotionSensitivity, camera.MotionTimeout, camera.VideoDurationMinutes,
		camera.ScreenshotIntervalMinutes, camera.Enabled, camera.ID)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "Camera", ID: camera.ID}
	}

	return nil
}

// Delete deletes a camera by ID.
func (r *CameraRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM cameras WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "Camera", ID: id}
	}

	return nil
}
