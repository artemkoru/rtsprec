# Design

## Context

`cmd/rtsprec/main.go` builds the system itself: it loads YAML via `internal/config`, opens the database, creates the config store with an encryption key file, initializes the S3 uploader, and registers handlers from `internal/api` and `internal/auth`. `internal/app` has a separate wiring (`New`, `Run`, `Shutdown`, `LoadConfig` from environment, `RegisterRoutes`), and its routes are `501` stubs. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- One assembly path: `main` → `app.New()` → `app.Init()` → `app.Run()`.
- Configuration from environment variables only.
- Every endpoint currently served by the binary is still served after the change.
- `internal/capture` compiles against a published gortsplib version.

**Non-Goals:**
- Deleting unused packages, example config, or the systemd unit.
- Changing request/response formats of any endpoint.

## Decisions

- **Split `New` into `New` and `Init`.** `New()` loads configuration and nothing else, so tests can build a `Config` without touching the database. `Init()` builds dependencies, routes and the HTTP server, and returns an error instead of partially constructing the app. Alternative: keep everything in `New` — rejected, because tests could not inject configuration without side effects.
- **Environment-only configuration.** Read by `LoadConfig()` via `os.LookupEnv` with defaults, as it already does. Variables that exist only in YAML today are added with the same naming style (`UPPER_SNAKE`). Alternative: keep YAML as an optional source — rejected by the user in favor of env only.
- **Routes come from `internal/api`.** `Init()` registers the handlers from `api.NewAPIServer(...)` on the mux. The `501` stubs in `router.go` are removed. Alternative: implement the stubs in `internal/app` — rejected, the handlers already exist and are used by the binary today.
- **Port from configuration.** `Config.HTTPPort` is used for `Addr` as `":" + port`, replacing the hardcoded `:8080`.
- **Shutdown.** `main` listens for SIGINT/SIGTERM and calls `app.Shutdown(ctx)` with a timeout, matching the current binary.

## Risks / Trade-offs

- [Deployments with `--config` break on upgrade] → Document the new environment variables in the change tasks and in a follow-up update to `deployments/rtsprec.service`.
- [Encryption key and config store move to env-based paths] → Keep the key path configurable through an environment variable with the current default `/opt/rtsprec/.encryption.key`.
- [Routes are wired to two packages with different handler signatures] → Verify each route with a test that calls the mux directly.
- [Porting `internal/capture` to a newer gortsplib changes its API surface] → Keep the exported `capture` interface used by `internal/api` and `internal/motion` unchanged; the port is the first task group and is verified before the wiring changes start.
- [gortsplib v5 may not be the intended replacement] → Pick the version during the first task group and record the choice in `go.mod` and in `docs/KNOWN_ISSUES.md`.

## Migration Plan

1. Deploy the new binary with environment variables set in the service unit.
2. Rollback: redeploy the previous binary and its YAML config; no database migration is involved.

## Open Questions

- None that change the approach. Which packages are removed is decided in a separate step.
