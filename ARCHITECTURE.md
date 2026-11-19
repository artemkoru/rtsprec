# RTSP Surveillance System - Clean Architecture

This document describes the layered clean architecture of the RTSP Surveillance System refactoring (Phase 10 of the development plan).

## Overview

The system is organized into distinct layers, each with a single responsibility and clear dependencies flowing downward:

```
┌─────────────────────────────────┐
│   HTTP Handlers                 │  - HTTP request/response translation
│   (internal/handler/)           │  - Thin routing layer only
└──────────────┬──────────────────┘
               │ depends on
┌──────────────▼──────────────────┐
│   Service Layer                 │  - Business logic and orchestration
│   (internal/service/)           │  - Coordinates repositories
└──────────────┬──────────────────┘
               │ depends on
┌──────────────▼──────────────────┐
│   Repository Interfaces         │  - Abstract storage access
│   (internal/repository/)        │
└──────────────┬──────────────────┘
               │ depends on
┌──────┬───────▼──────┬──────────┐
│      │               │          │
▼      ▼               ▼          ▼
SQLite S3        (Future)   Other Backends
Implementation   Implementations
```

## Package Structure

### internal/domain/
**Responsibility:** Core business entities and value objects with no external dependencies.

**Files:**
- `camera.go` - Camera entity (RTSP stream source)
- `motion.go` - MotionEvent entity (motion detection records)
- `video.go` - VideoFile and VideoChunk entities (video metadata and segments)
- `archive.go` - Timeline, ArchiveQuery, TimelineItem entities (archive queries)
- `auth.go` - User and Session entities (authentication)
- `errors.go` - Domain-level error types (NotFoundError, ValidationError, etc.)

**Key Characteristics:**
- No dependencies on framework, storage, or HTTP
- Pure Go types with methods
- Validation methods as needed (e.g., camera.IsValid())
- Error types for domain-specific failures

**Example:**
```go
type Camera struct {
    ID                        string
    Name                      string
    RtspURL                   string
    Status                    string
    MotionSensitivity         int
    // ... other fields
}
```

### internal/repository/
**Responsibility:** Define storage abstraction interfaces and coordinate implementations.

**Files:**
- `repository.go` - Interface definitions for all repository types

**Interfaces:**
- `CameraRepository` - CRUD operations for cameras
- `MotionRepository` - Query and storage for motion events
- `VideoRepository` - Video metadata and chunk storage
- `ArchiveRepository` - Archive queries and timeline generation
- `AuthRepository` - User and session management

**Key Characteristics:**
- Pure Go interfaces with no implementation
- Methods use `context.Context` for cancellation/timeouts
- Error return values are untyped `error` (typed in domain/errors.go)
- Dependency inversion: handlers and services depend on these interfaces

**Example:**
```go
type CameraRepository interface {
    FindAll(ctx context.Context) ([]domain.Camera, error)
    FindByID(ctx context.Context, id string) (*domain.Camera, error)
    Create(ctx context.Context, camera *domain.Camera) (string, error)
    Update(ctx context.Context, camera *domain.Camera) error
    Delete(ctx context.Context, id string) error
}
```

### internal/repository/sqlite/
**Responsibility:** SQLite implementations of repository interfaces for metadata storage.

**Files:**
- `camera.go` - CameraRepository implementation
- `motion.go` - MotionRepository implementation
- `video.go` - VideoRepository implementation (metadata only)
- `archive.go` - ArchiveRepository implementation
- `auth.go` - AuthRepository implementation

**Key Characteristics:**
- Implements repository.go interfaces
- Uses `database/sql` standard library
- Handles all SQLite-specific schema and queries
- Database transactions managed here if needed
- Can be swapped with mock implementations in tests

### internal/repository/s3/
**Responsibility:** S3 implementations for video storage (chunks and files).

**Files:**
- `video.go` - VideoRepository S3 implementation (video chunks and files)

**Key Characteristics:**
- Implements VideoRepository interface
- Uses AWS SDK for S3 operations
- Handles video chunk upload and retrieval
- Object key naming and path management
- Can be replaced with local filesystem implementation if needed

### internal/service/
**Responsibility:** Business logic and service orchestration.

**Files:**
- `camera.go` - CameraService (camera CRUD and status)
- `motion.go` - MotionService (motion event queries)
- `video.go` - VideoService (video file operations)
- `archive.go` - ArchiveService (archive queries and timeline)
- `auth.go` - AuthService (authentication and sessions)

**Key Characteristics:**
- Receives repositories via constructor injection
- Orchestrates across multiple repositories when needed
- Implements business logic and validation
- Logs operations using injected logger
- Returns domain types and untyped errors
- Can be unit tested with mock repositories

**Example:**
```go
type CameraService struct {
    cameraRepo repository.CameraRepository
    logger     *slog.Logger
}

func NewCameraService(repo repository.CameraRepository, log *slog.Logger) *CameraService {
    return &CameraService{cameraRepo: repo, logger: log}
}

func (s *CameraService) GetAll(ctx context.Context) ([]domain.Camera, error) {
    cameras, err := s.cameraRepo.FindAll(ctx)
    if err != nil {
        s.logger.Error("failed to get cameras", "error", err)
        return nil, err
    }
    return cameras, nil
}
```

### internal/handler/
**Responsibility:** HTTP request/response translation only.

**Files:**
- `camera.go` - Camera HTTP handlers (GET /api/cameras, POST /api/cameras, etc.)
- `motion.go` - Motion event HTTP handlers
- `video.go` - Video file HTTP handlers
- `archive.go` - Archive HTTP handlers
- `auth.go` - Authentication HTTP handlers
- `errors.go` - Error response formatting

**Key Characteristics:**
- Receives services via constructor injection
- Translates HTTP requests to domain operations
- Converts domain/service errors to HTTP status codes
- Returns JSON responses
- No direct database access
- Can be unit tested with mock services

**Example:**
```go
type CameraHandler struct {
    cameraService CameraService  // interface, not concrete type
    logger        *slog.Logger
}

func (h *CameraHandler) HandleGetCameras(w http.ResponseWriter, r *http.Request) {
    cameras, err := h.cameraService.GetAll(r.Context())
    if err != nil {
        h.writeError(w, http.StatusInternalServerError, "failed to get cameras")
        return
    }
    h.writeJSON(w, http.StatusOK, cameras)
}
```

### internal/app/
**Responsibility:** Application bootstrap, configuration, and dependency injection wiring.

**Files:**
- `app.go` - Main App struct with lifecycle methods (New, Run, Shutdown)
- `dependencies.go` - Service and repository initialization (dependency injection)
- `database.go` - SQLite database initialization
- `storage.go` - S3 client initialization
- `config.go` - Configuration loading from environment/files
- `router.go` - HTTP route registration
- `middleware.go` - HTTP middleware setup (logging, CORS, auth)

**Key Characteristics:**
- Centralizes all infrastructure setup
- Wires all dependencies together
- Provides clean app initialization interface
- Handles graceful shutdown
- Can be used to initialize app in tests
- Separates bootstrap concerns from main.go

**Example:**
```go
type App struct {
    server   *http.Server
    db       *sql.DB
    services *Services
    config   *Config
}

func New() (*App, error) {
    cfg := LoadConfig()
    db := InitDB(cfg)
    s3 := InitS3(cfg)
    services := SetupDependencies(db, s3)
    router := RegisterRoutes(services)
    // ... setup complete
    return &app, nil
}

func (a *App) Run() error {
    return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
    return a.server.Shutdown(ctx)
}
```

### cmd/rtsprec/
**Responsibility:** Application entry point (main.go only).

**Files:**
- `main.go` - Bootstrap and run application

**Key Characteristics:**
- Minimal - typically 10-15 lines of code
- Creates app.New() and calls app.Run()
- Handles signal shutdown (SIGINT, SIGTERM)
- All infrastructure moved to internal/app

**Example:**
```go
func main() {
    app, err := app.New()
    if err != nil {
        log.Fatal(err)
    }
    
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Dependency Flow

**Allowed dependencies:**
- Handlers → Services (via interface)
- Services → Repositories (via interface)
- Repositories → Storage backends (SQLite, S3)
- App → all packages (for wiring)

**Forbidden dependencies:**
- Handlers → Repositories (must go through services)
- Handlers → Storage (must go through services and repositories)
- Domain → anything (domain has no external dependencies)
- Services → Handlers (handlers depend on services, not vice versa)

## Testing Strategy

### Unit Testing
- **Services:** Mock repositories, test business logic
- **Handlers:** Mock services, test HTTP translation
- **Repositories:** Mock database/S3, test query construction

### Integration Testing
- **Services + Real Repositories:** Test with real SQLite/S3
- **Full HTTP Stack:** Test handlers + services + real repositories

### Example Service Test
```go
func TestCameraService_GetAll(t *testing.T) {
    mockRepo := &MockCameraRepository{}
    mockRepo.OnFindAll = func(ctx context.Context) ([]domain.Camera, error) {
        return []domain.Camera{{ID: "1", Name: "Camera 1"}}, nil
    }
    
    svc := service.NewCameraService(mockRepo, testLogger)
    cameras, err := svc.GetAll(context.Background())
    
    assert.NoError(t, err)
    assert.Len(t, cameras, 1)
}
```

## Error Handling

**Domain errors** (defined in internal/domain/errors.go):
- `NotFoundError` - Resource not found
- `ValidationError` - Input validation failure
- `ConflictError` - Operation conflict (duplicate key, etc.)

**Handler conversions** (internal/handler/errors.go):
- Domain errors → HTTP status codes (404, 400, 409, etc.)
- Logging of all errors for debugging

**Example:**
```go
// Service returns domain error
cameras, err := service.GetCamera(ctx, id)
if err != nil {
    if _, ok := err.(*domain.NotFoundError); ok {
        h.writeError(w, http.StatusNotFound, "Camera not found")
    } else {
        h.writeError(w, http.StatusInternalServerError, "Internal server error")
    }
}
```

## Configuration

Configuration is centralized in `internal/app/config.go`:
- Environment variables for secrets (DB connection, S3 credentials)
- Configuration file for non-sensitive settings (port, timeouts, etc.)
- Loaded once at app startup via `app.New()`

## Future Enhancements

The clean architecture supports future improvements:
- **New storage backends:** Implement Repository interfaces for PostgreSQL, etc.
- **Caching layer:** Add cache between services and repositories
- **Event publishing:** Add event types and publishers to services
- **gRPC interface:** Add gRPC handlers alongside HTTP handlers
- **CLI interface:** Add CLI command handlers alongside HTTP handlers
- **Background jobs:** Add job scheduling to App package

All of these can be added without modifying existing code due to dependency inversion.

## Migration Notes

**What changed:**
- Database/S3 access moved behind repository interfaces
- Business logic extracted to service layer
- HTTP handlers simplified to routing only
- Infrastructure setup consolidated in app package
- Main.go simplified to app bootstrap

**What stayed the same:**
- External API contracts (HTTP routes, payloads)
- Database schema
- S3 bucket structure
- Deployment process
- Configuration format

**Benefits:**
- Testability: Each layer can be tested independently with mocks
- Maintainability: Clear responsibilities, easy to understand data flow
- Extensibility: New features added to appropriate layer
- Flexibility: Easy to swap implementations (e.g., different databases)
- Independence: Handlers don't care about storage details
