package capture

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/artemkoru/rtsprec/internal/storage"
)

type ScreenshotManager struct {
	cameraID        string
	intervalMinutes int
	tempDir         string
	uploader        *storage.S3Uploader
	logger          *slog.Logger

	lastScreenshot time.Time
	mu             sync.Mutex
	stopChan       chan struct{}
}

func NewScreenshotManager(cameraID string, intervalMinutes int, tempDir string, uploader *storage.S3Uploader) *ScreenshotManager {
	return &ScreenshotManager{
		cameraID:        cameraID,
		intervalMinutes: intervalMinutes,
		tempDir:         tempDir,
		uploader:        uploader,
		logger:          slog.Default().With("camera_id", cameraID),
		stopChan:        make(chan struct{}),
	}
}

func (sm *ScreenshotManager) Start(ctx context.Context, frameChan <-chan FrameData) {
	if sm.intervalMinutes <= 0 {
		sm.logger.Info("screenshot scheduling disabled")
		return
	}

	go sm.run(ctx, frameChan)
}

func (sm *ScreenshotManager) Stop() {
	close(sm.stopChan)
}

func (sm *ScreenshotManager) run(ctx context.Context, frameChan <-chan FrameData) {
	sm.logger.Info("screenshot manager started", "interval_minutes", sm.intervalMinutes)

	ticker := time.NewTicker(time.Duration(sm.intervalMinutes) * time.Minute)
	defer ticker.Stop()

	var lastFrame FrameData
	var mu sync.Mutex

	// Goroutine to capture latest frame
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sm.stopChan:
				return
			case frame, ok := <-frameChan:
				if !ok {
					return
				}
				mu.Lock()
				lastFrame = frame
				mu.Unlock()
			}
		}
	}()

	// Initial screenshot
	if err := sm.captureAndUpload(ctx, lastFrame, &mu); err != nil {
		sm.logger.Error("failed to capture initial screenshot", "error", err)
	}

	// Schedule subsequent screenshots
	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			mu.Lock()
			frame := lastFrame
			mu.Unlock()

			if err := sm.captureAndUpload(ctx, frame, &mu); err != nil {
				sm.logger.Error("failed to capture screenshot", "error", err)
			}
		}
	}
}

func (sm *ScreenshotManager) captureAndUpload(ctx context.Context, frame FrameData, mu *sync.Mutex) error {
	if len(frame.Data) == 0 {
		return fmt.Errorf("no frame data available")
	}

	// Create temp JPEG file (simplified - real implementation would encode frame to JPEG)
	timestamp := time.Now().UTC()
	tempFile := filepath.Join(sm.tempDir, fmt.Sprintf("screenshot-%s-%d.jpg", sm.cameraID, timestamp.Unix()))

	if err := os.WriteFile(tempFile, frame.Data, 0644); err != nil {
		return fmt.Errorf("failed to write temp screenshot file: %w", err)
	}

	defer os.Remove(tempFile)

	// Upload to S3
	_, err := sm.uploader.UploadScreenshot(ctx, tempFile, sm.cameraID)
	if err != nil {
		return fmt.Errorf("failed to upload screenshot: %w", err)
	}

	sm.mu.Lock()
	sm.lastScreenshot = timestamp
	sm.mu.Unlock()

	sm.logger.Info("screenshot captured and uploaded", "timestamp", timestamp)
	return nil
}

func (sm *ScreenshotManager) GetLastScreenshotTime() time.Time {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.lastScreenshot
}
