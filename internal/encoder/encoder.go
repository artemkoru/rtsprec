package encoder

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/artemkoru/rtsprec/internal/capture"
	"github.com/google/uuid"
)

type EncodedChunk struct {
	ID        string
	CameraID  string
	StartTime time.Time
	EndTime   time.Time
	FilePath  string
	FileSize  int64
}

type VideoEncoder struct {
	cameraID           string
	chunkDuration      time.Duration
	outputPath         string
	logger             *slog.Logger

	chunkChan          chan EncodedChunk
	frameBuffer        []capture.FrameData
	bufferMutex        sync.Mutex
	lastChunkStartTime time.Time
}

func NewVideoEncoder(cameraID string, chunkDurationMinutes int, outputPath string) *VideoEncoder {
	return &VideoEncoder{
		cameraID:      cameraID,
		chunkDuration: time.Duration(chunkDurationMinutes) * time.Minute,
		outputPath:    outputPath,
		logger:        slog.Default().With("camera_id", cameraID),
		chunkChan:     make(chan EncodedChunk, 1),
		frameBuffer:   make([]capture.FrameData, 0, 1000),
	}
}

func (ve *VideoEncoder) Start(ctx context.Context, frameChan <-chan capture.FrameData) {
	go ve.run(ctx, frameChan)
}

func (ve *VideoEncoder) ChunkChannel() <-chan EncodedChunk {
	return ve.chunkChan
}

func (ve *VideoEncoder) run(ctx context.Context, frameChan <-chan capture.FrameData) {
	defer close(ve.chunkChan)

	ve.lastChunkStartTime = time.Now()
	chunkTicker := time.NewTicker(ve.chunkDuration)
	defer chunkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			ve.flushChunk()
			return

		case <-chunkTicker.C:
			// Time to create a chunk
			ve.flushChunk()
			ve.lastChunkStartTime = time.Now()

		case frame, ok := <-frameChan:
			if !ok {
				ve.flushChunk()
				return
			}

			ve.bufferMutex.Lock()
			ve.frameBuffer = append(ve.frameBuffer, frame)
			ve.bufferMutex.Unlock()
		}
	}
}

func (ve *VideoEncoder) flushChunk() {
	ve.bufferMutex.Lock()
	if len(ve.frameBuffer) == 0 {
		ve.bufferMutex.Unlock()
		return
	}

	// Copy frames and clear buffer
	frames := make([]capture.FrameData, len(ve.frameBuffer))
	copy(frames, ve.frameBuffer)
	ve.frameBuffer = ve.frameBuffer[:0]
	ve.bufferMutex.Unlock()

	// Create chunk
	chunk := ve.createChunk(frames)
	if chunk != nil {
		select {
		case ve.chunkChan <- *chunk:
			ve.logger.Info("video chunk created and queued", "chunk_id", chunk.ID, "file_size", chunk.FileSize)
		default:
			ve.logger.Warn("chunk channel full, dropping chunk")
			os.Remove(chunk.FilePath)
		}
	}
}

func (ve *VideoEncoder) createChunk(frames []capture.FrameData) *EncodedChunk {
	if len(frames) == 0 {
		return nil
	}

	// Create output directory
	cameraPath := filepath.Join(ve.outputPath, ve.cameraID)
	if err := os.MkdirAll(cameraPath, 0755); err != nil {
		ve.logger.Error("failed to create output directory", "error", err)
		return nil
	}

	// Generate filename with timestamp
	startTime := frames[0].Timestamp
	filename := fmt.Sprintf("%s-%s.mp4", startTime.Format("2006-01-02T15-04-05"), uuid.New().String()[:8])
	filePath := filepath.Join(cameraPath, filename)

	// Simplified: concatenate frame data (real implementation would use ffmpeg or libav)
	// For now, we'll just write frame data as-is
	var totalData []byte
	for _, frame := range frames {
		totalData = append(totalData, frame.Data...)
	}

	if err := os.WriteFile(filePath, totalData, 0644); err != nil {
		ve.logger.Error("failed to write chunk file", "error", err)
		return nil
	}

	fileInfo, _ := os.Stat(filePath)
	fileSize := fileInfo.Size()

	endTime := frames[len(frames)-1].Timestamp

	return &EncodedChunk{
		ID:        uuid.New().String(),
		CameraID:  ve.cameraID,
		StartTime: startTime,
		EndTime:   endTime,
		FilePath:  filePath,
		FileSize:  fileSize,
	}
}
