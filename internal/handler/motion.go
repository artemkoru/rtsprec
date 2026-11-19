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

// MotionService interface defines motion service methods.
type MotionService interface {
	GetEvents(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error)
	GetEventsByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error)
	CreateEvent(ctx context.Context, event *domain.MotionEvent) (string, error)
	UpdateEvent(ctx context.Context, event *domain.MotionEvent) error
}

// MotionHandler handles motion event HTTP requests.
type MotionHandler struct {
	motionService MotionService
	logger        *slog.Logger
}

// NewMotionHandler creates a new motion handler.
func NewMotionHandler(motionService MotionService, logger *slog.Logger) *MotionHandler {
	return &MotionHandler{
		motionService: motionService,
		logger:        logger,
	}
}

// HandleGetEvents returns motion events for a camera.
func (h *MotionHandler) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	if cameraID == "" {
		h.writeError(w, http.StatusBadRequest, "missing camera_id parameter")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
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

	events, err := h.motionService.GetEvents(r.Context(), cameraID, limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get events: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, events)
}

// HandleGetEventsByDateRange returns motion events within a date range.
func (h *MotionHandler) HandleGetEventsByDateRange(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if cameraID == "" {
		h.writeError(w, http.StatusBadRequest, "missing camera_id parameter")
		return
	}

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid from timestamp")
			return
		}
	} else {
		from = time.Now().AddDate(0, 0, -7)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid to timestamp")
			return
		}
	} else {
		to = time.Now()
	}

	events, err := h.motionService.GetEventsByDateRange(r.Context(), cameraID, from, to)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get events: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, events)
}

// Helper methods
func (h *MotionHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

func (h *MotionHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}
