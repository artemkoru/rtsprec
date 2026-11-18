package domain

import "time"

// MotionEvent represents a detected motion event from a camera.
type MotionEvent struct {
	ID                   string
	CameraID             string
	StartTime            time.Time
	EndTime              *time.Time
	DurationSeconds      *int
	SensitivityThreshold *int
	VideoFilePath        *string
	CreatedAt            time.Time
}
