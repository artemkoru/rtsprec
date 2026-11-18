package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
)

// ArchiveService interface defines archive service methods.
type ArchiveService interface {
	QueryVideos(ctx context.Context, query *domain.ArchiveQuery) ([]domain.VideoFile, error)
	QueryEvents(ctx context.Context, query *domain.ArchiveQuery) ([]domain.MotionEvent, error)
	GenerateTimeline(ctx context.Context, cameraID string, date time.Time) (*domain.Timeline, error)
}

// ArchiveHandler handles archive query HTTP requests.
type ArchiveHandler struct {
	archiveService ArchiveService
	logger         *slog.Logger
}

// NewArchiveHandler creates a new archive handler.
func NewArchiveHandler(archiveService ArchiveService, logger *slog.Logger) *ArchiveHandler {
	return &ArchiveHandler{
		archiveService: archiveService,
		logger:         logger,
	}
}

// HandleGetTimeline returns a timeline for a camera on a specific date.
func (h *ArchiveHandler) HandleGetTimeline(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	dateStr := r.URL.Query().Get("date")

	if cameraID == "" || dateStr == "" {
		h.writeError(w, http.StatusBadRequest, "missing camera_id or date parameter")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid date format: %v", err))
		return
	}

	timeline, err := h.archiveService.GenerateTimeline(r.Context(), cameraID, date)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate timeline: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, timeline)
}

// HandleQueryArchive returns videos and events within a date range.
func (h *ArchiveHandler) HandleQueryArchive(w http.ResponseWriter, r *http.Request) {
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

	query := &domain.ArchiveQuery{
		CameraID:  cameraID,
		StartTime: start,
		EndTime:   end,
	}

	// Query videos and events in parallel
	videoChan := make(chan []domain.VideoFile, 1)
	eventChan := make(chan []domain.MotionEvent, 1)
	errChan := make(chan error, 2)

	go func() {
		videos, err := h.archiveService.QueryVideos(r.Context(), query)
		if err != nil {
			errChan <- err
			return
		}
		videoChan <- videos
	}()

	go func() {
		events, err := h.archiveService.QueryEvents(r.Context(), query)
		if err != nil {
			errChan <- err
			return
		}
		eventChan <- events
	}()

	var videos []domain.VideoFile
	var events []domain.MotionEvent

	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
			return
		case v := <-videoChan:
			videos = v
		case e := <-eventChan:
			events = e
		}
	}

	result := map[string]interface{}{
		"videos": videos,
		"events": events,
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Helper methods
func (h *ArchiveHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ArchiveHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}
