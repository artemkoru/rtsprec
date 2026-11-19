package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// VideoService interface defines video service methods.
type VideoService interface {
	GetFile(ctx context.Context, id string) (*domain.VideoFile, error)
	ListFiles(ctx context.Context, cameraID string, limit, offset int) ([]domain.VideoFile, error)
	GetFilesByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.VideoFile, error)
	SaveChunk(ctx context.Context, chunk *domain.VideoChunk, data []byte) (string, error)
	IndexFile(ctx context.Context, file *domain.VideoFile) (string, error)
	UpdateFileMetadata(ctx context.Context, file *domain.VideoFile) error
	DeleteFile(ctx context.Context, id string) error
}

// VideoHandler handles video file HTTP requests.
type VideoHandler struct {
	videoService VideoService
	logger       *slog.Logger
}

// NewVideoHandler creates a new video handler.
func NewVideoHandler(videoService VideoService, logger *slog.Logger) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		logger:       logger,
	}
}

// HandleGetFile returns a video file by ID.
func (h *VideoHandler) HandleGetFile(w http.ResponseWriter, r *http.Request) {
	videoID := r.URL.Query().Get("id")
	if videoID == "" {
		h.writeError(w, http.StatusBadRequest, "missing id parameter")
		return
	}

	file, err := h.videoService.GetFile(r.Context(), videoID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			h.writeError(w, http.StatusNotFound, "video file not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get video file: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, file)
}

// HandleListFiles returns video files for a camera.
func (h *VideoHandler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	if cameraID == "" {
		h.writeError(w, http.StatusBadRequest, "missing camera_id parameter")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	files, err := h.videoService.ListFiles(r.Context(), cameraID, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list video files: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, files)
}

// HandleGetFilesByDateRange returns video files within a date range.
func (h *VideoHandler) HandleGetFilesByDateRange(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if cameraID == "" || startStr == "" || endStr == "" {
		h.writeError(w, http.StatusBadRequest, "missing required parameters: camera_id, start, end")
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid start timestamp")
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid end timestamp")
		return
	}

	files, err := h.videoService.GetFilesByDateRange(r.Context(), cameraID, start, end)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get video files: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, files)
}

// Helper methods
func (h *VideoHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

func (h *VideoHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}
