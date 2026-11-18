package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/google/uuid"
)

// CameraService interface defines camera service methods.
type CameraService interface {
	GetAll(ctx context.Context) ([]domain.Camera, error)
	GetByID(ctx context.Context, id string) (*domain.Camera, error)
	Create(ctx context.Context, camera *domain.Camera) (string, error)
	Update(ctx context.Context, camera *domain.Camera) error
	Delete(ctx context.Context, id string) error
}

// CameraHandler handles camera-related HTTP requests.
type CameraHandler struct {
	cameraService CameraService
	logger        *slog.Logger
}

// NewCameraHandler creates a new camera handler.
func NewCameraHandler(cameraService CameraService, logger *slog.Logger) *CameraHandler {
	return &CameraHandler{
		cameraService: cameraService,
		logger:        logger,
	}
}

// HandleGetCameras returns all cameras.
func (h *CameraHandler) HandleGetCameras(w http.ResponseWriter, r *http.Request) {
	cameras, err := h.cameraService.GetAll(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get cameras: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, cameras)
}

// HandleAddCamera creates a new camera.
func (h *CameraHandler) HandleAddCamera(w http.ResponseWriter, r *http.Request) {
	type AddCameraRequest struct {
		Name                      string `json:"name"`
		RtspURL                   string `json:"rtsp_url"`
		Username                  string `json:"username"`
		Password                  string `json:"password"`
		MotionSensitivity         int    `json:"motion_sensitivity"`
		MotionTimeout             int    `json:"motion_timeout"`
		VideoDurationMinutes      int    `json:"video_duration_minutes"`
		ScreenshotIntervalMinutes int    `json:"screenshot_interval_minutes"`
	}

	var req AddCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	camera := &domain.Camera{
		ID:                        uuid.New().String(),
		Name:                      req.Name,
		RtspURL:                   req.RtspURL,
		Username:                  req.Username,
		Password:                  req.Password,
		MotionSensitivity:         req.MotionSensitivity,
		MotionTimeout:             req.MotionTimeout,
		VideoDurationMinutes:      req.VideoDurationMinutes,
		ScreenshotIntervalMinutes: req.ScreenshotIntervalMinutes,
		Enabled:                   true,
		Status:                    "offline",
	}

	cameraID, err := h.cameraService.Create(r.Context(), camera)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create camera: %v", err))
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{"id": cameraID})
}

// HandleCameraStatus returns the status of a camera.
func (h *CameraHandler) HandleCameraStatus(w http.ResponseWriter, r *http.Request) {
	cameraID := strings.TrimPrefix(r.URL.Path, "/api/cameras/")
	cameraID = strings.TrimSuffix(cameraID, "/status")

	camera, err := h.cameraService.GetByID(r.Context(), cameraID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			h.writeError(w, http.StatusNotFound, "camera not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get camera: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"camera_id": cameraID,
		"status":    camera.Status,
		"enabled":   camera.Enabled,
	})
}

// HandleUpdateCamera updates a camera.
func (h *CameraHandler) HandleUpdateCamera(w http.ResponseWriter, r *http.Request) {
	cameraID := strings.TrimPrefix(r.URL.Path, "/api/cameras/")

	var camera domain.Camera
	if err := json.NewDecoder(r.Body).Decode(&camera); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	camera.ID = cameraID
	if err := h.cameraService.Update(r.Context(), &camera); err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			h.writeError(w, http.StatusNotFound, "camera not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update camera: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// HandleDeleteCamera deletes a camera.
func (h *CameraHandler) HandleDeleteCamera(w http.ResponseWriter, r *http.Request) {
	cameraID := strings.TrimPrefix(r.URL.Path, "/api/cameras/")

	if err := h.cameraService.Delete(r.Context(), cameraID); err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			h.writeError(w, http.StatusNotFound, "camera not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete camera: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods
func (h *CameraHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *CameraHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}
