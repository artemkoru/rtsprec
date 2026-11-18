# RTSP Surveillance System - Architecture

## Overview

The RTSP Surveillance System has been refactored to follow clean architecture principles with clear separation of concerns between HTTP handlers, business logic, data access, and domain entities.

## Package Structure

```
internal/
├── domain/              # Core business entities and value objects
│   ├── camera.go        # Camera entity
│   ├── motion.go        # MotionEvent entity
│   ├── video.go         # VideoFile and VideoChunk entities
│   ├── archive.go       # Timeline and ArchiveQuery entities
│   ├── auth.go          # User and Session entities
│   └── errors.go        # Domain error types
│
├── repository/          # Data access abstraction layer
│   ├── repository.go    # All repository interfaces
│   ├── sqlite/          # SQLite implementations
│   │   ├── camera.go
│   │   ├── motion.go
│   │   ├── video.go
│   │   ├── archive.go
│   │   └── auth.go
│   └── s3/              # S3 implementations
│       └── video.go
│
├── service/             # Business logic orchestration
│   ├── camera.go        # CameraService
│   ├── motion.go        # MotionService
│   ├── video.go         # VideoService
│   ├── archive.go       # ArchiveService
│   └── auth.go          # AuthService
│
├── handler/             # HTTP request handlers
│   ├── camera.go        # Camera HTTP endpoints
│   ├── motion.go        # Motion event endpoints
│   ├── video.go         # Video file endpoints
│   ├── archive.go       # Archive query endpoints
│   ├── auth.go          # Authentication endpoints
│   └── errors.go        # Error conversion helpers
│
└── app/                 # Application bootstrap and setup
    ├── app.go           # App lifecycle (New, Run, Shutdown)
    ├── config.go        # Configuration loading
    ├── database.go      # Database initialization
    ├── storage.go       # S3 storage setup
    ├── dependencies.go  # Dependency injection wiring
    ├── router.go        # HTTP route registration
    └── middleware.go    # HTTP middleware setup
```

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

The domain layer contains all business entities and value objects:

- **Entities**: Represent core business concepts (Camera, MotionEvent, VideoFile, User, Session, Timeline)
- **Value Objects**: Immutable objects representing specific values (ArchiveQuery)
- **Errors**: Domain-specific error types (NotFound, Validation, Conflict, Unauthorized, Internal)

Domain entities are **independent** of any technology or framework. They contain no database logic, HTTP logic, or external service dependencies.

### 2. Repository Layer (`internal/repository/`)

The repository layer provides data access abstraction:

- **Interfaces**: Define contracts for data access (CameraRepository, MotionRepository, VideoRepository, ArchiveRepository, AuthRepository)
- **SQLite Implementations**: Concrete implementations using SQLite for persistent storage
- **S3 Implementations**: Concrete implementations for video storage in S3

Repositories hide storage implementation details behind interfaces. Services depend on repository interfaces, not concrete implementations, enabling easy testing with mocks and future backend swaps.

**Repository Pattern Benefits:**
- Testability: Mock repositories in service tests
- Swappability: Change storage without changing services
- Single Responsibility: Each repository handles one data source

### 3. Service Layer (`internal/service/`)

The service layer contains business logic and orchestration:

- **Services**: Orchestrate repositories and implement business workflows
- **Dependency Injection**: Services receive repositories via constructors
- **Context Support**: All service methods accept context for cancellation and timeouts

Services are **technology-agnostic**. They don't know about HTTP, databases, or other infrastructure details. They coordinate domain logic and delegate data access to repositories.

**Service Examples:**
- `CameraService.GetAll()`: Retrieves all cameras via repository
- `VideoService.SaveChunk()`: Saves video chunk and returns path via repository
- `ArchiveService.GenerateTimeline()`: Coordinates video and event repositories to build timeline

### 4. Handler Layer (`internal/handler/`)

The handler layer translates HTTP requests to domain operations:

- **HTTP Endpoints**: Map routes to domain operations
- **Marshaling**: Convert HTTP requests to domain objects and responses to JSON
- **Error Conversion**: Translate domain errors to HTTP status codes
- **Service Delegation**: All business logic delegated to services

Handlers are **thin routing layers** focused on HTTP concerns only. They don't contain database queries or business logic.

**Handler Responsibilities:**
- Parse HTTP request (query parameters, body, headers)
- Call appropriate service method
- Convert domain errors to HTTP responses
- Serialize response to JSON

### 5. App Layer (`internal/app/`)

The app layer bootstraps the application:

- **App Struct**: Holds all application state (server, database, services, config)
- **Initialization**: `New()` creates and wires all components
- **Lifecycle**: `Run()` starts server, `Shutdown()` gracefully stops
- **Middleware**: CORS, logging, panic recovery
- **Routes**: HTTP route registration
- **Dependencies**: Dependency injection wiring

## Dependency Injection Pattern

Services and repositories are wired together in `app/dependencies.go`:

```go
// Example: Wire camera service with its repository
cameraRepo := sqlite.NewCameraRepository(db)
cameraService := service.NewCameraService(cameraRepo, logger)
```

**Benefits:**
- **Testability**: Inject mock repositories in tests
- **Flexibility**: Easy to swap implementations
- **Clarity**: Dependencies explicit in function signatures
- **Maintenance**: Clear object ownership and lifecycle

## Data Flow Example: Get Cameras

```
HTTP Request: GET /api/cameras
    ↓
Handler.HandleGetCameras()
    ↓
CameraService.GetAll(ctx)
    ↓
CameraRepository.FindAll(ctx)  [via SQLite]
    ↓
Database query
    ↓
Domain.Camera objects
    ↓
JSON response
```

## Error Handling

Domain errors are converted to HTTP responses:

| Domain Error | HTTP Status | Meaning |
|---|---|---|
| `NotFoundError` | 404 | Resource doesn't exist |
| `ValidationError` | 400 | Invalid input data |
| `ConflictError` | 409 | Resource conflicts (e.g., duplicate name) |
| `UnauthorizedError` | 401 | Authentication required |
| `InternalError` | 500 | Unexpected server error |

## Configuration

Configuration is loaded from environment variables in `app/config.go`:

```bash
DB_PATH=/path/to/database.db
S3_BUCKET=my-bucket
S3_REGION=us-east-1
HTTP_PORT=8080
LOG_LEVEL=info
ENVIRONMENT=production
```

## Testing Strategy

**Unit Tests**: Test services with mock repositories
```go
// Mock repository
mockRepo := &MockCameraRepository{}
// Inject into service
service := service.NewCameraService(mockRepo, logger)
// Test business logic in isolation
```

**Integration Tests**: Test services with real repositories
- Use in-memory SQLite `:memory:` for fast tests
- Create test data fixtures
- Verify service → repository → database flow

**Handler Tests**: Test HTTP layer with mock services
- Mock service responses
- Verify HTTP marshaling
- Test error handling

## Future Considerations

### Transactions
Consider adding transaction support across repositories:
```go
tx := cameraRepo.BeginTx(ctx)
// Multiple operations
tx.Commit()
```

### Caching
Add caching layer between services and repositories if needed:
```go
service → cache → repository → database
```

### Middleware
Additional middleware can be added in `app/middleware.go`:
- Authentication/Authorization
- Rate limiting
- Request tracking
- Performance monitoring

### Monitoring
Add observability:
- Structured logging (already implemented)
- Distributed tracing
- Metrics/Prometheus integration
- Health checks endpoint

## Migration Notes

This refactoring moves from tightly-coupled handlers to a clean architecture:

- **External API**: Unchanged (same HTTP routes and responses)
- **Database Schema**: Unchanged
- **Dependencies**: No new external packages required
- **Deployment**: Single binary, same configuration

The refactoring improves **internal code quality** without affecting users or deployment.

## Key Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Dependency Inversion**: Services depend on abstractions (interfaces), not concrete implementations
3. **Domain-Driven Design**: Business entities independent of technology
4. **Testability**: Business logic testable in isolation without HTTP/database
5. **Go Idiomatic**: Follows Go conventions and best practices
