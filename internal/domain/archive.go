package domain

import "time"

// Timeline represents a ordered collection of timeline items for a specific camera on a specific date.
type Timeline struct {
	CameraID  string
	Date      time.Time
	Items     []TimelineItem
	CreatedAt time.Time
}

// TimelineItem represents either a video segment or a motion event in the timeline.
type TimelineItem struct {
	Type      string        // "video" or "event"
	StartTime time.Time
	EndTime   time.Time
	Duration  int
	Video     *VideoFile
	Event     *MotionEvent
}

// ArchiveQuery represents a query for archive data with filtering criteria.
type ArchiveQuery struct {
	CameraID  string
	StartTime time.Time
	EndTime   time.Time
	Limit     *int
	Offset    *int
}
