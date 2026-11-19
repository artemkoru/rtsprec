package motion

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/artemkoru/rtsprec/internal/capture"
	"github.com/artemkoru/rtsprec/internal/db"
	"github.com/google/uuid"
)

type MotionEvent struct {
	ID                   string
	CameraID             string
	StartTime            time.Time
	EndTime              *time.Time
	DurationSeconds      int
	SensitivityThreshold int
	VideoFilePath        string
}

type MotionDetector struct {
	cameraID        string
	db              *db.DB
	sensitivity     int     // 0-100 scale
	frameSkip       int     // analyze every Nth frame
	motionThreshold float64 // percentage of pixels that changed
	logger          *slog.Logger

	currentEvent      *MotionEvent
	eventMutex        sync.RWMutex
	lastMotionTime    time.Time
	motionTimeoutSecs int

	frameCount int
}

func NewMotionDetector(cameraID string, database *db.DB, sensitivity int, frameSkip int, motionTimeout int) *MotionDetector {
	if sensitivity < 0 {
		sensitivity = 0
	}
	if sensitivity > 100 {
		sensitivity = 100
	}
	if frameSkip < 1 {
		frameSkip = 10
	}

	return &MotionDetector{
		cameraID:          cameraID,
		db:                database,
		sensitivity:       sensitivity,
		frameSkip:         frameSkip,
		motionThreshold:   float64(100-sensitivity) / 100.0, // Higher sensitivity = lower threshold
		logger:            slog.Default().With("camera_id", cameraID),
		motionTimeoutSecs: motionTimeout,
	}
}

func (md *MotionDetector) Start(ctx context.Context, frameChan <-chan capture.FrameData) {
	go md.run(ctx, frameChan)
}

func (md *MotionDetector) run(ctx context.Context, frameChan <-chan capture.FrameData) {
	var previousFrame []byte

	for {
		select {
		case <-ctx.Done():
			md.endEvent()
			return
		case frame, ok := <-frameChan:
			if !ok {
				md.endEvent()
				return
			}

			md.frameCount++

			// Skip frames for efficiency
			if md.frameCount%md.frameSkip != 0 {
				continue
			}

			// Decode frame (simplified - actual implementation would use proper video decoder)
			if previousFrame == nil {
				previousFrame = frame.Data
				continue
			}

			// Calculate motion score
			motionScore := md.calculateMotionScore(previousFrame, frame.Data)

			if motionScore > md.motionThreshold {
				md.handleMotionDetected(frame.Timestamp, md.sensitivity)
			} else {
				md.checkMotionTimeout()
			}

			previousFrame = frame.Data
		}
	}
}

func (md *MotionDetector) calculateMotionScore(prev, curr []byte) float64 {
	if len(prev) != len(curr) {
		return 0
	}

	changedPixels := 0
	threshold := 30 // Pixel difference threshold (0-255)

	// Compare pixel data (simplified - compares every 4 bytes as a pixel)
	for i := 0; i < len(prev)-3; i += 4 {
		diff := 0
		for j := 0; j < 4; j++ {
			d := int(prev[i+j]) - int(curr[i+j])
			if d < 0 {
				d = -d
			}
			diff += d
		}

		if diff > threshold*4 {
			changedPixels++
		}
	}

	totalPixels := len(prev) / 4
	if totalPixels == 0 {
		return 0
	}

	return float64(changedPixels) / float64(totalPixels)
}

func (md *MotionDetector) handleMotionDetected(timestamp time.Time, sensitivityThreshold int) {
	md.eventMutex.Lock()
	defer md.eventMutex.Unlock()

	now := time.Now()
	md.lastMotionTime = now

	if md.currentEvent == nil {
		// Start new motion event
		md.currentEvent = &MotionEvent{
			ID:                   uuid.New().String(),
			CameraID:             md.cameraID,
			StartTime:            now,
			SensitivityThreshold: sensitivityThreshold,
		}
		md.logger.Info("motion detected", "event_id", md.currentEvent.ID)
	}
}

func (md *MotionDetector) checkMotionTimeout() {
	md.eventMutex.Lock()
	defer md.eventMutex.Unlock()

	if md.currentEvent == nil {
		return
	}

	elapsedSecs := int(time.Since(md.lastMotionTime).Seconds())
	if elapsedSecs >= md.motionTimeoutSecs {
		md.endEventLocked()
	}
}

func (md *MotionDetector) endEvent() {
	md.eventMutex.Lock()
	defer md.eventMutex.Unlock()
	md.endEventLocked()
}

func (md *MotionDetector) endEventLocked() {
	if md.currentEvent == nil {
		return
	}

	now := time.Now()
	md.currentEvent.EndTime = &now
	md.currentEvent.DurationSeconds = int(md.currentEvent.EndTime.Sub(md.currentEvent.StartTime).Seconds())

	// Save to database
	if err := md.saveEvent(md.currentEvent); err != nil {
		md.logger.Error("failed to save motion event", "error", err)
	} else {
		md.logger.Info("motion event ended", "event_id", md.currentEvent.ID, "duration_secs", md.currentEvent.DurationSeconds)
	}

	md.currentEvent = nil
}

func (md *MotionDetector) saveEvent(event *MotionEvent) error {
	_, err := md.db.Conn().Exec(
		`INSERT INTO motion_events (id, camera_id, start_time, end_time, duration_seconds, sensitivity_threshold, video_file_path)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.CameraID,
		event.StartTime,
		event.EndTime,
		event.DurationSeconds,
		event.SensitivityThreshold,
		event.VideoFilePath,
	)
	return err
}

func (md *MotionDetector) GetCurrentEvent() *MotionEvent {
	md.eventMutex.RLock()
	defer md.eventMutex.RUnlock()

	if md.currentEvent == nil {
		return nil
	}

	// Return a copy
	event := *md.currentEvent
	return &event
}

func (md *MotionDetector) SetSensitivity(sensitivity int) {
	if sensitivity < 0 {
		sensitivity = 0
	}
	if sensitivity > 100 {
		sensitivity = 100
	}

	md.eventMutex.Lock()
	defer md.eventMutex.Unlock()

	md.sensitivity = sensitivity
	md.motionThreshold = float64(100-sensitivity) / 100.0

	md.logger.Info("motion sensitivity updated", "sensitivity", sensitivity)
}

func (md *MotionDetector) SetFrameSkip(frameSkip int) {
	if frameSkip < 1 {
		frameSkip = 1
	}

	md.eventMutex.Lock()
	defer md.eventMutex.Unlock()

	md.frameSkip = frameSkip
	md.logger.Info("frame skip updated", "frame_skip", frameSkip)
}
