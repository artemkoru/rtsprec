// Package domain defines the core business entities and value objects for the RTSP surveillance system.
package domain

import "time"

// Camera represents a video camera source that captures RTSP streams.
type Camera struct {
	ID                        string
	Name                      string
	RtspURL                   string
	Username                  string
	Password                  string
	Status                    string
	MotionSensitivity         int
	MotionTimeout             int
	VideoDurationMinutes      int
	ScreenshotIntervalMinutes int
	Enabled                   bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
