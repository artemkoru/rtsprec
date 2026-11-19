package hls

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Segment struct {
	Index     int
	Filename  string
	Duration  float64
	Path      string
	CreatedAt time.Time
}

type HLSGenerator struct {
	cameraID        string
	outputPath      string
	segmentDuration int
	segmentsPerList int
	logger          *slog.Logger

	segments     []*Segment
	segmentsMu   sync.RWMutex
	currentIndex int
}

func NewHLSGenerator(cameraID, outputPath string, segmentDuration, segmentsPerList int) *HLSGenerator {
	return &HLSGenerator{
		cameraID:        cameraID,
		outputPath:      outputPath,
		segmentDuration: segmentDuration,
		segmentsPerList: segmentsPerList,
		logger:          slog.Default().With("camera_id", cameraID),
		segments:        make([]*Segment, 0, segmentsPerList+1),
		currentIndex:    0,
	}
}

func (h *HLSGenerator) Init() error {
	cameraPath := filepath.Join(h.outputPath, h.cameraID)
	if err := os.MkdirAll(cameraPath, 0755); err != nil {
		return fmt.Errorf("failed to create HLS output directory: %w", err)
	}

	h.logger.Info("HLS generator initialized", "output_path", cameraPath)
	return nil
}

func (h *HLSGenerator) AddSegment(segmentData []byte) (string, error) {
	h.segmentsMu.Lock()
	defer h.segmentsMu.Unlock()

	// Create segment filename
	segmentFilename := fmt.Sprintf("segment-%05d.ts", h.currentIndex)
	segmentPath := filepath.Join(h.outputPath, h.cameraID, segmentFilename)

	// Write segment to disk
	if err := os.WriteFile(segmentPath, segmentData, 0644); err != nil {
		return "", fmt.Errorf("failed to write segment: %w", err)
	}

	// Create segment metadata
	segment := &Segment{
		Index:     h.currentIndex,
		Filename:  segmentFilename,
		Duration:  float64(h.segmentDuration),
		Path:      segmentPath,
		CreatedAt: time.Now(),
	}

	// Add to segments list
	h.segments = append(h.segments, segment)

	// Keep only recent segments (sliding window)
	if len(h.segments) > h.segmentsPerList+1 {
		oldSegment := h.segments[0]
		h.segments = h.segments[1:]

		// Delete old segment file
		if err := os.Remove(oldSegment.Path); err != nil {
			h.logger.Warn("failed to delete old segment", "path", oldSegment.Path)
		}
	}

	h.currentIndex++

	return segmentFilename, nil
}

func (h *HLSGenerator) GeneratePlaylist() (string, error) {
	h.segmentsMu.RLock()
	defer h.segmentsMu.RUnlock()

	if len(h.segments) == 0 {
		return h.generateEmptyPlaylist(), nil
	}

	// Build M3U8 playlist
	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", h.segmentDuration)
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"

	// Add segments
	for _, segment := range h.segments {
		playlist += fmt.Sprintf("#EXTINF:%.1f,\n", segment.Duration)
		playlist += segment.Filename + "\n"
	}

	return playlist, nil
}

func (h *HLSGenerator) generateEmptyPlaylist() string {
	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", h.segmentDuration)
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"
	return playlist
}

func (h *HLSGenerator) GetSegmentList() []*Segment {
	h.segmentsMu.RLock()
	defer h.segmentsMu.RUnlock()

	// Return copy of segment list
	segments := make([]*Segment, len(h.segments))
	copy(segments, h.segments)
	return segments
}

func (h *HLSGenerator) GetSegmentFile(filename string) ([]byte, error) {
	h.segmentsMu.RLock()

	// Verify filename is in our segments
	var found bool
	for _, seg := range h.segments {
		if seg.Filename == filename {
			found = true
			break
		}
	}
	h.segmentsMu.RUnlock()

	if !found {
		return nil, fmt.Errorf("segment not found: %s", filename)
	}

	segmentPath := filepath.Join(h.outputPath, h.cameraID, filename)
	data, err := os.ReadFile(segmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read segment file: %w", err)
	}

	return data, nil
}

func (h *HLSGenerator) CleanUp() error {
	h.segmentsMu.Lock()
	defer h.segmentsMu.Unlock()

	cameraPath := filepath.Join(h.outputPath, h.cameraID)
	if err := os.RemoveAll(cameraPath); err != nil {
		return fmt.Errorf("failed to cleanup HLS directory: %w", err)
	}

	h.segments = make([]*Segment, 0, h.segmentsPerList+1)
	h.currentIndex = 0

	return nil
}
