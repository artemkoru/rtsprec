// Package domain defines the core business entities and value objects for the RTSP surveillance system.
package domain

import "time"

// VideoFile represents a complete video file stored in S3 with metadata tracked in the database.
type VideoFile struct {
	ID              string
	CameraID        string
	S3Path          string
	FileName        string
	StartTime       time.Time
	EndTime         time.Time
	DurationSeconds *int
	FileSizeBytes   *int
	Codec           *string
	Resolution      *string
	UploadedAt      time.Time
}

// VideoChunk represents a segment of video being captured before it's combined into a VideoFile.
type VideoChunk struct {
	ID            string
	CameraID      string
	StartTime     time.Time
	EndTime       time.Time
	FilePath      string
	FileSizeBytes int64
	CreatedAt     time.Time
}
