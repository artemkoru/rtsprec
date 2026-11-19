package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/artemkoru/rtsprec/internal/archive"
	"github.com/artemkoru/rtsprec/internal/capture"
	"github.com/artemkoru/rtsprec/internal/config"
	"github.com/artemkoru/rtsprec/internal/db"
	"github.com/artemkoru/rtsprec/internal/motion"
	"github.com/artemkoru/rtsprec/internal/storage"
	"github.com/google/uuid"
)

type APIServer struct {
	db              *db.DB
	configStore     *config.ConfigStore
	archiveManager  *archive.ArchiveManager
	s3Uploader      *storage.S3Uploader
	motionDetectors map[string]*motion.MotionDetector
	logger          *slog.Logger
}

func NewAPIServer(database *db.DB, cfg *config.ConfigStore, arch *archive.ArchiveManager, uploader *storage.S3Uploader) *APIServer {
	return &APIServer{
		db:              database,
		configStore:     cfg,
		archiveManager:  arch,
		s3Uploader:      uploader,
		motionDetectors: make(map[string]*motion.MotionDetector),
		logger:          slog.Default(),
	}
}

// Camera Management Endpoints
func (api *APIServer) HandleGetCameras(w http.ResponseWriter, r *http.Request) {
	rows, err := api.db.Conn().Query(`
		SELECT id, name, rtsp_url, status, motion_sensitivity, motion_timeout, video_duration_minutes, screenshot_interval_minutes, enabled, created_at
		FROM cameras
		ORDER BY created_at ASC
	`)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
		return
	}
	defer rows.Close()

	type Camera struct {
		ID                        string    `json:"id"`
		Name                      string    `json:"name"`
		RtspURL                   string    `json:"rtsp_url"`
		Status                    string    `json:"status"`
		MotionSensitivity         int       `json:"motion_sensitivity"`
		MotionTimeout             int       `json:"motion_timeout"`
		VideoDurationMinutes      int       `json:"video_duration_minutes"`
		ScreenshotIntervalMinutes int       `json:"screenshot_interval_minutes"`
		Enabled                   bool      `json:"enabled"`
		CreatedAt                 time.Time `json:"created_at"`
	}

	cameras := []Camera{}
	for rows.Next() {
		var c Camera
		if err := rows.Scan(&c.ID, &c.Name, &c.RtspURL, &c.Status, &c.MotionSensitivity, &c.MotionTimeout, &c.VideoDurationMinutes, &c.ScreenshotIntervalMinutes, &c.Enabled, &c.CreatedAt); err != nil {
			api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("scan failed: %v", err))
			return
		}
		cameras = append(cameras, c)
	}

	api.writeJSON(w, http.StatusOK, cameras)
}

func (api *APIServer) HandleAddCamera(w http.ResponseWriter, r *http.Request) {
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
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	cameraID := uuid.New().String()
	_, err := api.db.Conn().Exec(`
		INSERT INTO cameras (id, name, rtsp_url, username, password, motion_sensitivity, motion_timeout, video_duration_minutes, screenshot_interval_minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cameraID, req.Name, req.RtspURL, req.Username, req.Password, req.MotionSensitivity, req.MotionTimeout, req.VideoDurationMinutes, req.ScreenshotIntervalMinutes)

	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}

	api.writeJSON(w, http.StatusCreated, map[string]string{"id": cameraID})
}

// Camera Status Endpoint
func (api *APIServer) HandleCameraStatus(w http.ResponseWriter, r *http.Request) {
	cameraID := strings.TrimPrefix(r.URL.Path, "/api/cameras/")
	cameraID = strings.TrimSuffix(cameraID, "/status")

	var status string
	err := api.db.Conn().QueryRow(`SELECT status FROM cameras WHERE id = ?`, cameraID).Scan(&status)
	if err != nil {
		api.writeError(w, http.StatusNotFound, "camera not found")
		return
	}

	response := map[string]interface{}{
		"camera_id":    cameraID,
		"status":       status,
		"last_updated": time.Now(),
	}

	api.writeJSON(w, http.StatusOK, response)
}

// Configuration Endpoints
func (api *APIServer) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	configType := strings.TrimPrefix(r.URL.Path, "/api/config/")

	var value string
	var err error

	switch configType {
	case "s3":
		value, _ = api.configStore.Get("s3_config")
	case "auth":
		value, _ = api.configStore.Get("auth_config")
	case "motion":
		value, _ = api.configStore.Get("motion_config")
	default:
		api.writeError(w, http.StatusBadRequest, "unknown config type")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{"config": value})
}

func (api *APIServer) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	type ConfigUpdateRequest struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Encrypted bool   `json:"encrypted"`
	}

	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	configKey := fmt.Sprintf("%s_config", req.Type)
	if err := api.configStore.Set(configKey, req.Value, req.Encrypted); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Archive Timeline Endpoint
func (api *APIServer) HandleTimeline(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	dateStr := r.URL.Query().Get("date")

	if cameraID == "" || dateStr == "" {
		api.writeError(w, http.StatusBadRequest, "missing camera_id or date parameter")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid date format: %v", err))
		return
	}

	timeline, err := api.archiveManager.GenerateTimeline(cameraID, date)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate timeline: %v", err))
		return
	}

	api.writeJSON(w, http.StatusOK, timeline)
}

// Motion Events Endpoint
func (api *APIServer) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	cameraID := r.URL.Query().Get("camera_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if cameraID == "" {
		api.writeError(w, http.StatusBadRequest, "missing camera_id parameter")
		return
	}

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			api.writeError(w, http.StatusBadRequest, "invalid from timestamp")
			return
		}
	} else {
		from = time.Now().AddDate(0, 0, -7) // Default: last 7 days
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			api.writeError(w, http.StatusBadRequest, "invalid to timestamp")
			return
		}
	} else {
		to = time.Now()
	}

	events, err := api.archiveManager.GetEventsByDateRange(cameraID, from, to)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
		return
	}

	api.writeJSON(w, http.StatusOK, events)
}

// Diagnostics Endpoint
func (api *APIServer) HandleDiagnostics(w http.ResponseWriter, r *http.Request) {
	diagnostics := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "ok",
		"database": map[string]string{
			"type": "sqlite",
		},
		"s3": map[string]string{
			"configured": "true",
		},
	}

	api.writeJSON(w, http.StatusOK, diagnostics)
}

// Helper methods
func (api *APIServer) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (api *APIServer) writeError(w http.ResponseWriter, statusCode int, message string) {
	api.writeJSON(w, statusCode, map[string]string{"error": message})
}

func (api *APIServer) RegisterMotionDetector(cameraID string, detector *motion.MotionDetector) {
	api.motionDetectors[cameraID] = detector
}
