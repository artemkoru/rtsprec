package archive

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type VideoFile struct {
	ID        string    `json:"id"`
	CameraID  string    `json:"camera_id"`
	S3Path    string    `json:"s3_path"`
	FileName  string    `json:"file_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration_seconds"`
	FileSize  int64     `json:"file_size_bytes"`
	Codec     string    `json:"codec"`
}

type MotionEvent struct {
	ID                   string     `json:"id"`
	CameraID             string     `json:"camera_id"`
	StartTime            time.Time  `json:"start_time"`
	EndTime              *time.Time `json:"end_time"`
	Duration             int        `json:"duration_seconds"`
	SensitivityThreshold int        `json:"sensitivity_threshold"`
	VideoFilePath        string     `json:"video_file_path"`
}

type TimelineItem struct {
	Type      string      `json:"type"` // "video" or "event"
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Duration  int         `json:"duration_seconds"`
	Video     *VideoFile  `json:"video,omitempty"`
	Event     *MotionEvent `json:"event,omitempty"`
}

type ArchiveManager struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewArchiveManager(db *sql.DB) *ArchiveManager {
	return &ArchiveManager{
		db:     db,
		logger: slog.Default(),
	}
}

func (am *ArchiveManager) IndexVideoFile(cameraID, s3Path, fileName, codec string, startTime, endTime time.Time, fileSize int64) (string, error) {
	var id string
	duration := int(endTime.Sub(startTime).Seconds())

	err := am.db.QueryRow(
		`INSERT INTO video_metadata (id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec, resolution)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 RETURNING id`,
		fmt.Sprintf("video_%d", time.Now().UnixNano()),
		cameraID,
		s3Path,
		fileName,
		startTime,
		endTime,
		duration,
		fileSize,
		codec,
		"1080p",
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to index video file: %w", err)
	}

	am.logger.Info("video file indexed", "file_id", id, "s3_path", s3Path)
	return id, nil
}

func (am *ArchiveManager) GetVideosByDateRange(cameraID string, startTime, endTime time.Time) ([]VideoFile, error) {
	rows, err := am.db.Query(
		`SELECT id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec
		 FROM video_metadata
		 WHERE camera_id = ? AND start_time >= ? AND end_time <= ?
		 ORDER BY start_time ASC`,
		cameraID,
		startTime,
		endTime,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var videos []VideoFile
	for rows.Next() {
		var v VideoFile
		if err := rows.Scan(&v.ID, &v.CameraID, &v.S3Path, &v.FileName, &v.StartTime, &v.EndTime, &v.Duration, &v.FileSize, &v.Codec); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}

	return videos, rows.Err()
}

func (am *ArchiveManager) GetEventsByDateRange(cameraID string, startTime, endTime time.Time) ([]MotionEvent, error) {
	rows, err := am.db.Query(
		`SELECT id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path
		 FROM motion_events
		 WHERE camera_id = ? AND start_time >= ? AND end_time IS NOT NULL AND end_time <= ?
		 ORDER BY start_time ASC`,
		cameraID,
		startTime,
		endTime,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var events []MotionEvent
	for rows.Next() {
		var e MotionEvent
		if err := rows.Scan(&e.ID, &e.CameraID, &e.StartTime, &e.EndTime, &e.Duration, &e.SensitivityThreshold, &e.VideoFilePath); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func (am *ArchiveManager) GenerateTimeline(cameraID string, date time.Time) ([]TimelineItem, error) {
	// Set date range: start of day to end of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	// Get all videos and events for the day
	videos, err := am.GetVideosByDateRange(cameraID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	events, err := am.GetEventsByDateRange(cameraID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// Merge videos and events into timeline
	timeline := make([]TimelineItem, 0, len(videos)+len(events))

	// Add videos
	for _, v := range videos {
		timeline = append(timeline, TimelineItem{
			Type:      "video",
			StartTime: v.StartTime,
			EndTime:   v.EndTime,
			Duration:  v.Duration,
			Video:     &v,
		})
	}

	// Add events
	for _, e := range events {
		if e.EndTime != nil {
			timeline = append(timeline, TimelineItem{
				Type:      "event",
				StartTime: e.StartTime,
				EndTime:   *e.EndTime,
				Duration:  e.Duration,
				Event:     &e,
			})
		}
	}

	// Sort timeline by start time
	sortTimeline(timeline)

	return timeline, nil
}

func (am *ArchiveManager) GetVideoFile(videoID string) (*VideoFile, error) {
	var v VideoFile
	err := am.db.QueryRow(
		`SELECT id, camera_id, s3_path, file_name, start_time, end_time, duration_seconds, file_size_bytes, codec
		 FROM video_metadata
		 WHERE id = ?`,
		videoID,
	).Scan(&v.ID, &v.CameraID, &v.S3Path, &v.FileName, &v.StartTime, &v.EndTime, &v.Duration, &v.FileSize, &v.Codec)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found")
	}

	return &v, err
}

func (am *ArchiveManager) GetEventsByCamera(cameraID string, limit int, offset int) ([]MotionEvent, error) {
	rows, err := am.db.Query(
		`SELECT id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path
		 FROM motion_events
		 WHERE camera_id = ?
		 ORDER BY start_time DESC
		 LIMIT ? OFFSET ?`,
		cameraID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []MotionEvent
	for rows.Next() {
		var e MotionEvent
		if err := rows.Scan(&e.ID, &e.CameraID, &e.StartTime, &e.EndTime, &e.Duration, &e.SensitivityThreshold, &e.VideoFilePath); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func sortTimeline(timeline []TimelineItem) {
	// Simple bubble sort for small timelines
	for i := 0; i < len(timeline); i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[j].StartTime.Before(timeline[i].StartTime) {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}
}
