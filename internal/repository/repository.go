// Package repository defines interfaces for data access abstraction across all storage backends (SQLite, S3, etc).
package repository

import (
	"context"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// CameraRepository provides access to camera entity data.
type CameraRepository interface {
	// FindAll returns all cameras.
	FindAll(ctx context.Context) ([]domain.Camera, error)
	// FindByID returns a single camera by ID.
	FindByID(ctx context.Context, id string) (*domain.Camera, error)
	// Create creates a new camera and returns its ID.
	Create(ctx context.Context, camera *domain.Camera) (string, error)
	// Update updates an existing camera.
	Update(ctx context.Context, camera *domain.Camera) error
	// Delete deletes a camera by ID.
	Delete(ctx context.Context, id string) error
}

// MotionRepository provides access to motion event data.
type MotionRepository interface {
	// FindByID returns a motion event by ID.
	FindByID(ctx context.Context, id string) (*domain.MotionEvent, error)
	// FindByCameraID returns all motion events for a camera.
	FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error)
	// FindByDateRange returns motion events within a date range for a camera.
	FindByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error)
	// Create creates a new motion event and returns its ID.
	Create(ctx context.Context, event *domain.MotionEvent) (string, error)
	// Update updates an existing motion event.
	Update(ctx context.Context, event *domain.MotionEvent) error
}

// VideoRepository provides access to video metadata and storage.
type VideoRepository interface {
	// FindByID returns video metadata by ID.
	FindByID(ctx context.Context, id string) (*domain.VideoFile, error)
	// FindByCameraID returns video files for a camera.
	FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.VideoFile, error)
	// FindByDateRange returns video files within a date range for a camera.
	FindByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.VideoFile, error)
	// Create creates a new video metadata record and returns its ID.
	Create(ctx context.Context, video *domain.VideoFile) (string, error)
	// UpdateMetadata updates video metadata (codec, resolution, etc).
	UpdateMetadata(ctx context.Context, video *domain.VideoFile) error
	// Save stores a video chunk and returns its path.
	Save(ctx context.Context, chunk *domain.VideoChunk, data []byte) (string, error)
	// Get retrieves a video chunk data.
	Get(ctx context.Context, path string) ([]byte, error)
	// Delete removes a video and its associated storage.
	Delete(ctx context.Context, id string) error
}

// ArchiveRepository provides access to archive queries and timeline generation.
type ArchiveRepository interface {
	// QueryVideos returns video files matching the query criteria.
	QueryVideos(ctx context.Context, query *domain.ArchiveQuery) ([]domain.VideoFile, error)
	// QueryEvents returns motion events matching the query criteria.
	QueryEvents(ctx context.Context, query *domain.ArchiveQuery) ([]domain.MotionEvent, error)
	// GenerateTimeline generates a timeline of videos and events for a specific date.
	GenerateTimeline(ctx context.Context, cameraID string, date time.Time) (*domain.Timeline, error)
}

// AuthRepository provides access to user and session data.
type AuthRepository interface {
	// GetUser returns a user by ID.
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	// CreateUser creates a new user and returns its ID.
	CreateUser(ctx context.Context, user *domain.User) (string, error)
	// GetSession returns a session by token.
	GetSession(ctx context.Context, token string) (*domain.Session, error)
	// CreateSession creates a new session and returns the token.
	CreateSession(ctx context.Context, session *domain.Session) (string, error)
	// DeleteSession deletes a session by token.
	DeleteSession(ctx context.Context, token string) error
	// ValidateSession checks if a session is valid (exists and not expired).
	ValidateSession(ctx context.Context, token string) (bool, error)
}
