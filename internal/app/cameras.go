package app

import (
	"context"
	"log/slog"

	"github.com/artemkoru/rtsprec/internal/capture"
	"github.com/artemkoru/rtsprec/internal/encoder"
	"github.com/artemkoru/rtsprec/internal/hls"
	"github.com/artemkoru/rtsprec/internal/motion"
)

// StartCameras starts capture for every enabled camera. Each camera runs in its own
// goroutine until Shutdown cancels the application context.
func (a *App) StartCameras() {
	logger := a.logger

	rows, err := a.db.Conn().Query(`SELECT id, name, rtsp_url, username, password, motion_sensitivity, motion_timeout, video_duration_minutes, screenshot_interval_minutes FROM cameras WHERE enabled = true`)
	if err != nil {
		logger.Error("failed to fetch cameras", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, rtspURL, username, password string
		var sensitivity, timeout, videoDuration, screenshotInterval int

		if err := rows.Scan(&id, &name, &rtspURL, &username, &password, &sensitivity, &timeout, &videoDuration, &screenshotInterval); err != nil {
			logger.Error("failed to scan camera", "error", err)
			continue
		}

		go a.captureCamera(a.ctx, id, name, rtspURL, username, password, sensitivity, timeout, videoDuration, screenshotInterval)
	}
}

// captureCamera runs the RTSP reader, motion detector, encoder, HLS generator and
// screenshot manager for one camera, and uploads encoded chunks until ctx is done.
func (a *App) captureCamera(parent context.Context, cameraID, name, rtspURL, username, password string, sensitivity, timeout, videoDuration, screenshotInterval int) {
	logger := slog.Default().With("camera_id", cameraID, "camera_name", name)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	// RTSP reader
	rtspReader := capture.NewRTSPReader(cameraID, rtspURL, username, password, 100)
	rtspReader.Start(ctx)
	defer rtspReader.Stop()

	// Motion detector
	motionDetector := motion.NewMotionDetector(cameraID, a.db, sensitivity, 15, timeout)
	motionDetector.Start(ctx, rtspReader.FrameChannel())

	// Video encoder
	videoEncoder := encoder.NewVideoEncoder(cameraID, videoDuration, "/var/cache/rtsprec/video")
	videoEncoder.Start(ctx, rtspReader.FrameChannel())

	// HLS generator
	hlsGen := hls.NewHLSGenerator(cameraID, a.config.HLSOutputPath, a.config.HLSSegmentDurationSecond, a.config.HLSSegmentsPerList)
	if err := hlsGen.Init(); err != nil {
		logger.Error("failed to initialize HLS generator", "error", err)
		return
	}
	a.hlsHandler.RegisterGenerator(cameraID, hlsGen)

	// Screenshot manager (if enabled)
	if screenshotInterval > 0 {
		screenshotMgr := capture.NewScreenshotManager(cameraID, screenshotInterval, "/tmp", a.s3Uploader)
		screenshotMgr.Start(ctx, rtspReader.FrameChannel())
		defer screenshotMgr.Stop()
	}

	logger.Info("camera capture started")

	// Upload video chunks and index them in the archive
	go func() {
		for chunk := range videoEncoder.ChunkChannel() {
			if s3Key, err := a.s3Uploader.UploadChunk(ctx, chunk.FilePath, cameraID); err != nil {
				logger.Error("failed to upload chunk", "error", err)
			} else {
				if _, err := a.archiveManager.IndexVideoFile(cameraID, s3Key, chunk.FilePath, "h264", chunk.StartTime, chunk.EndTime, chunk.FileSize); err != nil {
					logger.Error("failed to index video chunk", "error", err)
				}
			}
		}
	}()

	<-ctx.Done()
	logger.Info("camera capture stopped")
}
