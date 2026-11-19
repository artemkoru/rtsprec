package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/google/uuid"
)

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// MockCameraService for testing
type MockCameraService struct {
	cameras map[string]*domain.Camera
}

func NewMockCameraService() *MockCameraService {
	return &MockCameraService{
		cameras: make(map[string]*domain.Camera),
	}
}

func (m *MockCameraService) GetAll(ctx context.Context) ([]domain.Camera, error) {
	var result []domain.Camera
	for _, camera := range m.cameras {
		result = append(result, *camera)
	}
	return result, nil
}

func (m *MockCameraService) GetByID(ctx context.Context, id string) (*domain.Camera, error) {
	if camera, exists := m.cameras[id]; exists {
		return camera, nil
	}
	return nil, &domain.NotFoundError{Resource: "Camera", ID: id}
}

func (m *MockCameraService) Create(ctx context.Context, camera *domain.Camera) (string, error) {
	m.cameras[camera.ID] = camera
	return camera.ID, nil
}

func (m *MockCameraService) Update(ctx context.Context, camera *domain.Camera) error {
	if _, exists := m.cameras[camera.ID]; !exists {
		return &domain.NotFoundError{Resource: "Camera", ID: camera.ID}
	}
	m.cameras[camera.ID] = camera
	return nil
}

func (m *MockCameraService) Delete(ctx context.Context, id string) error {
	if _, exists := m.cameras[id]; !exists {
		return &domain.NotFoundError{Resource: "Camera", ID: id}
	}
	delete(m.cameras, id)
	return nil
}

// Tests
func TestCameraHandler_GetCameras(t *testing.T) {
	mockService := NewMockCameraService()
	handler := NewCameraHandler(mockService, getTestLogger())

	// Add test data
	camera := &domain.Camera{
		ID:   uuid.New().String(),
		Name: "Test Camera",
	}
	mockService.Create(context.Background(), camera)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/cameras", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetCameras(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Unmarshal response
	var cameras []domain.Camera
	body, _ := io.ReadAll(w.Body)
	json.Unmarshal(body, &cameras)

	if len(cameras) != 1 {
		t.Fatalf("Expected 1 camera, got %d", len(cameras))
	}
}

func TestCameraHandler_AddCamera(t *testing.T) {
	mockService := NewMockCameraService()
	handler := NewCameraHandler(mockService, getTestLogger())

	// Create request body
	reqData := map[string]interface{}{
		"name":                       "New Camera",
		"rtsp_url":                   "rtsp://example.com/stream",
		"motion_sensitivity":         50,
		"motion_timeout":             120,
		"video_duration_minutes":     5,
		"screenshot_interval_minutes": 0,
	}

	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest(http.MethodPost, "/api/cameras", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleAddCamera(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}

	// Check response body has ID
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] == "" {
		t.Fatalf("Expected camera ID in response")
	}
}

func TestCameraHandler_CameraStatus(t *testing.T) {
	mockService := NewMockCameraService()
	handler := NewCameraHandler(mockService, getTestLogger())

	// Add test camera
	camera := &domain.Camera{
		ID:     "cam123",
		Name:   "Test Camera",
		Status: "online",
	}
	mockService.Create(context.Background(), camera)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/cameras/cam123/status", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleCameraStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "online" {
		t.Fatalf("Expected status 'online', got '%v'", resp["status"])
	}
}

func TestCameraHandler_NotFound(t *testing.T) {
	mockService := NewMockCameraService()
	handler := NewCameraHandler(mockService, getTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/cameras/nonexistent/status", nil)
	w := httptest.NewRecorder()

	handler.HandleCameraStatus(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}
}

// MockMotionService for testing
type MockMotionService struct {
	events map[string]*domain.MotionEvent
}

func NewMockMotionService() *MockMotionService {
	return &MockMotionService{
		events: make(map[string]*domain.MotionEvent),
	}
}

func (m *MockMotionService) GetEvents(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error) {
	var result []domain.MotionEvent
	count := 0
	for _, event := range m.events {
		if event.CameraID == cameraID && count >= offset && count < offset+limit {
			result = append(result, *event)
		}
		if event.CameraID == cameraID {
			count++
		}
	}
	return result, nil
}

func (m *MockMotionService) GetEventsByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error) {
	var result []domain.MotionEvent
	for _, event := range m.events {
		if event.CameraID == cameraID && event.StartTime.After(start) && event.StartTime.Before(end) {
			result = append(result, *event)
		}
	}
	return result, nil
}

func (m *MockMotionService) CreateEvent(ctx context.Context, event *domain.MotionEvent) (string, error) {
	m.events[event.ID] = event
	return event.ID, nil
}

func (m *MockMotionService) UpdateEvent(ctx context.Context, event *domain.MotionEvent) error {
	m.events[event.ID] = event
	return nil
}

func TestMotionHandler_GetEvents(t *testing.T) {
	mockService := NewMockMotionService()
	handler := NewMotionHandler(mockService, getTestLogger())

	// Add test event
	event := &domain.MotionEvent{
		ID:       uuid.New().String(),
		CameraID: "cam1",
		StartTime: time.Now(),
	}
	mockService.CreateEvent(context.Background(), event)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/motion?camera_id=cam1", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleGetEvents(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var events []domain.MotionEvent
	json.Unmarshal(w.Body.Bytes(), &events)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
}
