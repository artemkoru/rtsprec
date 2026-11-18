package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// MotionRepository implements domain.MotionRepository using SQLite.
type MotionRepository struct {
	db *sql.DB
}

// NewMotionRepository creates a new SQLite motion repository.
func NewMotionRepository(db *sql.DB) *MotionRepository {
	return &MotionRepository{db: db}
}

// FindByID returns a motion event by ID.
func (r *MotionRepository) FindByID(ctx context.Context, id string) (*domain.MotionEvent, error) {
	var e domain.MotionEvent
	err := r.db.QueryRowContext(ctx, `
		SELECT id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path, created_at
		FROM motion_events
		WHERE id = ?
	`, id).Scan(&e.ID, &e.CameraID, &e.StartTime, &e.EndTime, &e.DurationSeconds,
		&e.SensitivityThreshold, &e.VideoFilePath, &e.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "MotionEvent", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &e, nil
}

// FindByCameraID returns motion events for a camera with pagination.
func (r *MotionRepository) FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path, created_at
		FROM motion_events
		WHERE camera_id = ?
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, cameraID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var events []domain.MotionEvent
	for rows.Next() {
		var e domain.MotionEvent
		if err := rows.Scan(&e.ID, &e.CameraID, &e.StartTime, &e.EndTime, &e.DurationSeconds,
			&e.SensitivityThreshold, &e.VideoFilePath, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// FindByDateRange returns motion events within a date range for a camera.
func (r *MotionRepository) FindByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path, created_at
		FROM motion_events
		WHERE camera_id = ? AND start_time >= ? AND end_time IS NOT NULL AND end_time <= ?
		ORDER BY start_time ASC
	`, cameraID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var events []domain.MotionEvent
	for rows.Next() {
		var e domain.MotionEvent
		if err := rows.Scan(&e.ID, &e.CameraID, &e.StartTime, &e.EndTime, &e.DurationSeconds,
			&e.SensitivityThreshold, &e.VideoFilePath, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// Create creates a new motion event and returns its ID.
func (r *MotionRepository) Create(ctx context.Context, event *domain.MotionEvent) (string, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO motion_events (id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.CameraID, event.StartTime, event.EndTime, event.DurationSeconds,
		event.SensitivityThreshold, event.VideoFilePath)

	if err != nil {
		return "", fmt.Errorf("insert failed: %w", err)
	}

	return event.ID, nil
}

// Update updates an existing motion event.
func (r *MotionRepository) Update(ctx context.Context, event *domain.MotionEvent) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE motion_events
		SET end_time = ?, duration_seconds = ?, sensitivity_threshold = ?, video_file_path = ?
		WHERE id = ?
	`, event.EndTime, event.DurationSeconds, event.SensitivityThreshold, event.VideoFilePath, event.ID)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "MotionEvent", ID: event.ID}
	}

	return nil
}
