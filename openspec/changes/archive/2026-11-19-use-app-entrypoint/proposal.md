# Proposal

## Why

`cmd/rtsprec` assembles the backend by hand (YAML config, database, config store, S3, routes), while `internal/app` holds a separate clean-architecture wiring that is not used by the binary. Two parallel assemblies drift apart, and the integration tests planned for the backend need a single, testable entry point to start the real application.

## What Changes

- `cmd/rtsprec/main.go` only calls `app.New()`, `app.Init()` and `app.Run()`, with graceful shutdown through `app.Shutdown()`.
- `internal/app` gets an `Init()` step that builds the database, services, middleware, routes and HTTP server. `New()` only loads configuration.
- **BREAKING**: configuration comes only from environment variables. The YAML config file (`-config` flag, `/opt/rtsprec/config.yaml`) is no longer read by the binary.
- `internal/app` gets the variables that the binary needs and that currently exist only in YAML (HLS paths, auth method and credentials, encryption key path, S3 credentials).
- The `501 Not implemented` stubs in `internal/app/router.go` are replaced with the real handlers from `internal/api`, so the endpoints the binary serves today keep working.
- Port comes from configuration instead of a hardcoded `:8080`.
- `internal/capture` is ported from the non-existent gortsplib v4 subpackages to a compiling gortsplib version, so that `internal/api` and `internal/app` build. Without this, the new wiring cannot compile.
- `internal/auth` compile errors are fixed (placeholder JSON decoder, invalid SameSite constant).
- `Init()` also builds what the old binary built: the database via `internal/db`, config store, S3 uploader, archive manager, authentication, API server, HLS handler, and the HTTP routes including the static `./static` directory and logout handler.
- Camera capture (RTSP reader, motion detection, encoder, HLS generator, chunk upload, screenshots) moves from `cmd/rtsprec/main.go` into `internal/app` as `StartCameras`. It is started by `Run` and stopped by `Shutdown`, so recording keeps working.

- Login and logout are public routes: they are registered outside the authentication middleware.
- Camera status and HLS routes use proper `{id}` / `{camera}` path wildcards instead of literal `*`.
- Unused files are deleted: `internal/handler`, `internal/logging`, `internal/app/database.go`, `internal/app/dependencies.go`, `internal/app/storage.go`, `config.example.yaml`, `deployments/rtsprec.service`, empty `pkg/`. CI, Makefile and docs no longer reference them.

Not in scope: the `domain`, `service` and `repository` packages stay. After this change the binary does not import them, but they are still tested in CI and documented in ARCHITECTURE.md. Removing them is a separate decision.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

<!-- None: the change alters how the application is assembled and configured, not the behavior of the existing capabilities. -->

## Impact

- `cmd/rtsprec/main.go`: reduced to the app lifecycle.
- `internal/app/app.go`, `internal/app/config.go`, `internal/app/router.go`: new `Init()`, extended environment configuration, real routes.
- `internal/api`, `internal/config` are used by `internal/app` after the change.
- Deployment: the service must be started with environment variables instead of `--config`. `deployments/rtsprec.service` and `config.example.yaml` will need updating in a follow-up.
- `internal/capture` and its gortsplib dependency (`go.mod`, `go.sum`) change; `docs/KNOWN_ISSUES.md` gets the gortsplib entry removed or updated.
