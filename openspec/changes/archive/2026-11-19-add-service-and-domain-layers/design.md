# Design: Clean Architecture Refactoring

## Context

The current implementation has HTTP handlers mixing routing, business logic, and database access. This tight coupling makes testing difficult and spreads domain knowledge across multiple files. The system uses SQLite for metadata and S3 for video storage with no abstraction layer. We need to restructure while preserving all existing functionality.

See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Separate concerns: HTTP layer, business logic, and data access are independent
- Enable unit testing of services without HTTP or database mocking infrastructure
- Create clear domain model with entities in one package
- Implement repository pattern to abstract storage details
- Use dependency injection for loose coupling
- Maintain all existing functionality during migration

**Non-Goals:**
- Change external API contracts or HTTP routes
- Add new features during refactoring
- Introduce new external dependencies
- Change database schema or S3 structure
- Modify system behavior or requirements

## Decisions

### 1. Single Domain Package vs. Domain Per Subdirectory

**Decision:** Single `internal/domain/` package containing all entity types.

**Rationale:**
- Go idiom: packages represent concepts, not layers
- Domains are tightly integrated (Camera has MotionEvents, VideoFiles, etc.)
- Single domain package avoids circular import issues
- Easier to maintain cross-domain invariants
- File organization is flat and clear: `camera.go`, `motion.go`, `video.go`, etc.

**Alternatives:**
- Subdirectories per domain (`domain/camera/`, `domain/motion/`) → adds complexity for small entities; harder to maintain relationships
- Separate `pkg` → exposes internal types publicly; not needed here

### 2. Repository Interfaces in Dedicated Package

**Decision:** Define all repository interfaces in `internal/repository/repository.go`, implement in subpackages (`sqlite/`, `s3/`).

**Rationale:**
- Interfaces live where they're defined (dependency inversion)
- Implementations grouped by backend technology
- Easy to swap implementations (mock in tests, real in production)
- Clear separation: `type CameraRepository interface { ... }` is unambiguous

**Alternatives:**
- Interfaces in domain package → creates circular imports (domain knows about storage)
- Interfaces in service package → violates single responsibility
- One file per interface → fragmented and hard to navigate

### 3. Service Layer Organization

**Decision:** `internal/service/` with one file per service (e.g., `camera.go`, `motion.go`).

**Rationale:**
- Service implements one or more domain interfaces
- Services depend on repositories via constructor injection
- File-per-service mirrors domain organization
- Each service has clear responsibilities

**Example:**
```go
// internal/service/camera.go
type cameraService struct {
  repo domain.CameraRepository
  logger *slog.Logger
}

func (s *cameraService) GetAll(ctx context.Context) ([]domain.Camera, error) {
  return s.repo.FindAll(ctx)
}
```

### 4. HTTP Handlers as Thin Routing Layer

**Decision:** Handlers only translate HTTP ↔ domain concerns; delegate to services.

**Rationale:**
- Handlers don't know about repositories or database
- HTTP errors (400, 404, 500) separate from domain errors
- Easy to test services independently
- Reusable in gRPC, CLI, or other interfaces

**Example:**
```go
// internal/handler/camera.go
func (h *CameraHandler) GetAll(w http.ResponseWriter, r *http.Request) {
  cameras, err := h.cameraService.GetAll(r.Context())
  if err != nil {
    // Convert domain error to HTTP error
  }
}
```

### 5. Application Package Consolidates Infrastructure

**Decision:** `internal/app/` contains server setup, database init, storage config, router, middleware, and DI wiring.

**Rationale:**
- Separates concerns: `cmd/` is bootstrap, `app/` is initialization
- `main.go` becomes a 10-line entry point
- All infrastructure wiring in one place
- Easy to test; `app.New()` can be called from tests

**Structure:**
```go
// internal/app/app.go
type App struct {
  server    *http.Server
  db        *sql.DB
  services  *Services
  config    *Config
}

func New() (*App, error) { /* init all */ }
func (a *App) Run() error { /* start server */ }
func (a *App) Shutdown(ctx context.Context) error { /* cleanup */ }
```

### 6. Dependency Injection via Constructor Pattern

**Decision:** Services and repositories receive dependencies via constructors; `app/dependencies.go` wires everything.

**Rationale:**
- No global state or service locator
- Dependencies explicit in function signatures
- Easy to mock in tests
- Go idiomatic pattern

**Example:**
```go
// internal/app/dependencies.go
func (a *App) setupServices() (*Services, error) {
  cameraRepo := sqlite.NewCameraRepository(a.db)
  cameraService := service.NewCameraService(cameraRepo, a.logger)
  
  return &Services{
    Camera: cameraService,
    // ... other services
  }, nil
}
```

### 7. main.go Simplified

**Decision:** `cmd/rtsprec/main.go` only bootstraps the app.

**Rationale:**
- Clear entry point
- All infrastructure details hidden in app package
- Easy to extend (add profiling, signal handling, etc.)

**Code:**
```go
func main() {
  app, err := app.New()
  if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
  }
  
  if err := app.Run(); err != nil {
    log.Fatalf("App error: %v", err)
  }
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Circular imports during refactor** | Interfaces in repository package, implementations separate. Test carefully when moving code. |
| **Duplicate logic when extracting** | Service methods may initially mirror repository methods. Refine in later phases; leave TODOs during migration. |
| **Performance overhead** | No new overhead; indirect function calls are negligible in Go. If profiling shows issues, optimize specific paths. |
| **More files to understand** | Architecture is clearer; easier for new developers. Each file has a single responsibility. |
| **Existing tests may break** | Tests currently call handlers directly; refactor to call services. Opportunity to add better unit tests. |
| **Deployment complexity** | None; single binary unchanged. Configuration and S3 setup unchanged. |

## Migration Plan

**Phase 1: Extract Domain (Day 1)**
- Create `internal/domain/` with entity types from existing code
- Move Camera, MotionEvent, VideoChunk, User structs
- Add validation methods as needed

**Phase 2: Extract Repositories (Day 2)**
- Define interfaces in `internal/repository/repository.go`
- Create `internal/repository/sqlite/` with implementations
- Create `internal/repository/s3/` for video S3 operations
- Update existing database code to implement interfaces

**Phase 3: Implement Services (Day 3)**
- Create `internal/service/` with service implementations
- Each service receives repository via constructor
- Services implement domain interfaces
- Services orchestrate repositories and business logic

**Phase 4: Refactor Handlers (Day 4)**
- Update `internal/handler/` to receive services
- Remove direct database calls
- Keep routing and HTTP logic only
- Add error type conversions (domain → HTTP)

**Phase 5: Create App Package (Day 5)**
- Extract infrastructure setup into `internal/app/`
- Create `app/app.go` with App struct and New()
- Create `app/dependencies.go` for DI wiring
- Create `app/router.go` for route registration
- Move database init to `app/database.go`
- Move config to `app/config.go`

**Phase 6: Simplify main.go (Day 6)**
- Replace all initialization code with `app.New()` and `app.Run()`
- Verify all functionality works
- Run tests
- Commit

**Rollback:** If issues arise:
- Revert to previous Git commit
- Each phase is an independent commit; partial rollback possible
- No database or schema changes, so no data risk

## Architectural Diagram

```
+-------+
| main  |
+---+---+
    |
    v
+--------+ New()
|  App   |---------> Setup
+---+----+          Config
    |               Database
    |               DI
    |
    v Run()
+------------------+
| HTTP Server      |
| + Router         |
+--+---------------+
   |
   v
+------------------+
| Handlers         |  Convert HTTP ↔ Domain
| (thin)           |
+--+---------------+
   |
   v
+------------------+
| Services         |  Business logic
+--+---------------+
   |
   v
+------------------+
| Repositories     |  Abstract storage
+--+--+--+---------+
   |  |  |
   v  v  v
 SQLite S3 (...)
```

## Implementation Notes

- **Testing**: Services can be tested with mock repositories; handlers with mock services
- **Error Handling**: Domain errors (validation, not found) separate from infrastructure errors
- **Logging**: Injected into services for consistent logging across system
- **Context**: All repository and service methods accept `context.Context` for cancellation and timeouts
- **Concurrency**: Repository implementations handle concurrent access (SQLite WAL, S3 semantics)

## Open Questions

1. **Error Types**: Should domain errors be custom types or wrapped stdlib errors? (Defer to implementation phase; start with stdlib wrapping)
2. **Caching**: Any caching layer between services and repositories? (Not in phase 1; add if performance testing shows need)
3. **Transactions**: Do services need explicit transaction support across multiple repositories? (Defer; handle in service implementations)
4. **Logging Format**: Structured logs already in place; maintain consistency across services? (Yes, inject existing logger)
