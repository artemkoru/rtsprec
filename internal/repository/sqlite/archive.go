package sqlite

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// ArchiveRepository implements domain.ArchiveRepository using SQLite.
type ArchiveRepository struct {
	videoRepo  *VideoRepository
	motionRepo *MotionRepository
}

// NewArchiveRepository creates a new SQLite archive repository.
func NewArchiveRepository(videoRepo *VideoRepository, motionRepo *MotionRepository) *ArchiveRepository {
	return &ArchiveRepository{
		videoRepo:  videoRepo,
		motionRepo: motionRepo,
	}
}

// QueryVideos returns video files matching the query criteria.
func (r *ArchiveRepository) QueryVideos(ctx context.Context, query *domain.ArchiveQuery) ([]domain.VideoFile, error) {
	return r.videoRepo.FindByDateRange(ctx, query.CameraID, query.StartTime, query.EndTime)
}

// QueryEvents returns motion events matching the query criteria.
func (r *ArchiveRepository) QueryEvents(ctx context.Context, query *domain.ArchiveQuery) ([]domain.MotionEvent, error) {
	return r.motionRepo.FindByDateRange(ctx, query.CameraID, query.StartTime, query.EndTime)
}

// GenerateTimeline generates a timeline of videos and events for a specific date.
func (r *ArchiveRepository) GenerateTimeline(ctx context.Context, cameraID string, date time.Time) (*domain.Timeline, error) {
	// Set date range: start of day to end of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	// Get all videos for the day
	videos, err := r.videoRepo.FindByDateRange(ctx, cameraID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query videos: %w", err)
	}

	// Get all events for the day
	events, err := r.motionRepo.FindByDateRange(ctx, cameraID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	// Create timeline items
	items := make([]domain.TimelineItem, 0, len(videos)+len(events))

	// Add videos
	for _, v := range videos {
		duration := 0
		if v.DurationSeconds != nil {
			duration = *v.DurationSeconds
		}
		items = append(items, domain.TimelineItem{
			Type:      "video",
			StartTime: v.StartTime,
			EndTime:   v.EndTime,
			Duration:  duration,
			Video:     &v,
		})
	}

	// Add events
	for _, e := range events {
		if e.EndTime != nil {
			duration := 0
			if e.DurationSeconds != nil {
				duration = *e.DurationSeconds
			}
			items = append(items, domain.TimelineItem{
				Type:      "event",
				StartTime: e.StartTime,
				EndTime:   *e.EndTime,
				Duration:  duration,
				Event:     &e,
			})
		}
	}

	// Sort timeline by start time
	sort.Slice(items, func(i, j int) bool {
		return items[i].StartTime.Before(items[j].StartTime)
	})

	return &domain.Timeline{
		CameraID:  cameraID,
		Date:      date,
		Items:     items,
		CreatedAt: time.Now(),
	}, nil
}
