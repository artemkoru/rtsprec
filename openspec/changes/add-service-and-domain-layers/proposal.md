# Proposal: Add Service and Domain Layers

## Why

The current implementation mixes HTTP concerns, business logic, and data access in a tightly coupled structure. This makes the codebase harder to test, modify, and maintain. Extracting a clear service layer with domain entities and repository pattern creates a clean architecture that separates concerns, improves testability, and aligns with Go best practices. This refactoring enables future teams to reason about the system more clearly and modify behavior independently of persistence concerns.

## What Changes

- **Domain Package**: Consolidate all entity and value object definitions into `internal/domain/` (camera, motion, video, archive, auth)
- **Repository Pattern**: Extract all database and storage access into interfaces and implementations in `internal/repository/` with concrete implementations for SQLite and S3
- **Service Layer**: Create business logic layer in `internal/service/` that orchestrates repositories and handles domain logic
- **Application Package**: Move all infrastructure setup (HTTP server, database, storage, config, middleware, dependency injection) from `cmd/rtsprec/main.go` into `internal/app/`
- **HTTP Handlers**: Thin out handlers in `internal/handler/` to focus only on HTTP concerns (routing, marshaling, error responses)
- **Dependency Injection**: Wire services and repositories via `app/dependencies.go` with constructor-based injection
- **main.go**: Simplify to only bootstrap the app: `app.New()` and `app.Run()`

## Capabilities

### New Capabilities

This is a pure refactoring—no new external behaviors or requirements are introduced. The system continues to:
- Capture RTSP streams
- Detect motion
- Encode and store video
- Serve live streams via HLS
- Provide archive access
- Manage authentication

The change restructures how these capabilities are implemented, not what they do.

### Modified Capabilities

None—existing spec-level behavior remains unchanged. This is architectural restructuring only.

## Impact

- **Internal Structure**: Go package organization follows domain-driven design with clear separation of concerns
- **Dependencies**: No new external dependencies required (uses existing libraries)
- **Backward Compatibility**: External API and behavior unchanged; only internal implementation details change
- **Testability**: Service layer can be tested independently of HTTP handlers; repositories can be mocked
- **Maintainability**: Clear boundaries make it easier to understand, modify, and extend functionality
- **Code Quality**: Aligns with idiomatic Go patterns (clean architecture, dependency injection, interface-based design)

## Implementation Scope

- **Phase 1**: Extract domain entities and repository interfaces
- **Phase 2**: Implement repository layer for SQLite and S3
- **Phase 3**: Create service implementations with injected repositories
- **Phase 4**: Thin out HTTP handlers to delegate to services
- **Phase 5**: Build app package with wiring and dependency injection
- **Phase 6**: Simplify main.go to boot the application
