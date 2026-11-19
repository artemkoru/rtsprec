package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/artemkoru/rtsprec/internal/domain"
	"github.com/google/uuid"
)

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// MockCameraRepository is a mock implementation for testing
type MockCameraRepository struct {
	cameras map[string]*domain.Camera
}

func NewMockCameraRepository() *MockCameraRepository {
	return &MockCameraRepository{
		cameras: make(map[string]*domain.Camera),
	}
}

func (m *MockCameraRepository) FindAll(ctx context.Context) ([]domain.Camera, error) {
	var result []domain.Camera
	for _, camera := range m.cameras {
		result = append(result, *camera)
	}
	return result, nil
}

func (m *MockCameraRepository) FindByID(ctx context.Context, id string) (*domain.Camera, error) {
	if camera, exists := m.cameras[id]; exists {
		return camera, nil
	}
	return nil, &domain.NotFoundError{Resource: "Camera", ID: id}
}

func (m *MockCameraRepository) Create(ctx context.Context, camera *domain.Camera) (string, error) {
	m.cameras[camera.ID] = camera
	return camera.ID, nil
}

func (m *MockCameraRepository) Update(ctx context.Context, camera *domain.Camera) error {
	if _, exists := m.cameras[camera.ID]; !exists {
		return &domain.NotFoundError{Resource: "Camera", ID: camera.ID}
	}
	m.cameras[camera.ID] = camera
	return nil
}

func (m *MockCameraRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.cameras[id]; !exists {
		return &domain.NotFoundError{Resource: "Camera", ID: id}
	}
	delete(m.cameras, id)
	return nil
}

// MockMotionRepository is a mock for motion events
type MockMotionRepository struct {
	events map[string]*domain.MotionEvent
}

func NewMockMotionRepository() *MockMotionRepository {
	return &MockMotionRepository{
		events: make(map[string]*domain.MotionEvent),
	}
}

func (m *MockMotionRepository) FindByID(ctx context.Context, id string) (*domain.MotionEvent, error) {
	if event, exists := m.events[id]; exists {
		return event, nil
	}
	return nil, &domain.NotFoundError{Resource: "MotionEvent", ID: id}
}

func (m *MockMotionRepository) FindByCameraID(ctx context.Context, cameraID string, limit, offset int) ([]domain.MotionEvent, error) {
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

func (m *MockMotionRepository) FindByDateRange(ctx context.Context, cameraID string, start, end time.Time) ([]domain.MotionEvent, error) {
	var result []domain.MotionEvent
	for _, event := range m.events {
		if event.CameraID == cameraID && event.StartTime.After(start) && event.StartTime.Before(end) {
			result = append(result, *event)
		}
	}
	return result, nil
}

func (m *MockMotionRepository) Create(ctx context.Context, event *domain.MotionEvent) (string, error) {
	m.events[event.ID] = event
	return event.ID, nil
}

func (m *MockMotionRepository) Update(ctx context.Context, event *domain.MotionEvent) error {
	if _, exists := m.events[event.ID]; !exists {
		return &domain.NotFoundError{Resource: "MotionEvent", ID: event.ID}
	}
	m.events[event.ID] = event
	return nil
}

// Tests
func TestCameraService_GetAll(t *testing.T) {
	mockRepo := NewMockCameraRepository()
	service := NewCameraService(mockRepo, getTestLogger())

	// Add test data
	camera1 := &domain.Camera{ID: "cam1", Name: "Camera 1"}
	camera2 := &domain.Camera{ID: "cam2", Name: "Camera 2"}
	mockRepo.Create(context.Background(), camera1)
	mockRepo.Create(context.Background(), camera2)

	// Test GetAll
	cameras, err := service.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(cameras) != 2 {
		t.Fatalf("Expected 2 cameras, got %d", len(cameras))
	}
}

func TestCameraService_Create(t *testing.T) {
	mockRepo := NewMockCameraRepository()
	service := NewCameraService(mockRepo, getTestLogger())

	camera := &domain.Camera{
		ID:   uuid.New().String(),
		Name: "New Camera",
	}

	cameraID, err := service.Create(context.Background(), camera)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if cameraID != camera.ID {
		t.Fatalf("Expected ID %s, got %s", camera.ID, cameraID)
	}

	// Verify it was created
	retrieved, _ := service.GetByID(context.Background(), cameraID)
	if retrieved.Name != "New Camera" {
		t.Fatalf("Expected name 'New Camera', got '%s'", retrieved.Name)
	}
}

func TestCameraService_NotFound(t *testing.T) {
	mockRepo := NewMockCameraRepository()
	service := NewCameraService(mockRepo, getTestLogger())

	_, err := service.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatalf("Expected NotFoundError, got nil")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected NotFoundError, got %T", err)
	}
}

func TestMotionService_GetEvents(t *testing.T) {
	mockRepo := NewMockMotionRepository()
	service := NewMotionService(mockRepo, getTestLogger())

	// Add test data
	event := &domain.MotionEvent{
		ID:        "event1",
		CameraID:  "cam1",
		StartTime: time.Now(),
	}
	mockRepo.Create(context.Background(), event)

	// Test GetEvents
	events, err := service.GetEvents(context.Background(), "cam1", 10, 0)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
}

func TestMotionService_CreateEvent(t *testing.T) {
	mockRepo := NewMockMotionRepository()
	service := NewMotionService(mockRepo, getTestLogger())

	event := &domain.MotionEvent{
		ID:        uuid.New().String(),
		CameraID:  "cam1",
		StartTime: time.Now(),
	}

	eventID, err := service.CreateEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if eventID != event.ID {
		t.Fatalf("Expected ID %s, got %s", event.ID, eventID)
	}
}
