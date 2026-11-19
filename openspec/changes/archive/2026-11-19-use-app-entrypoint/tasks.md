# Tasks

## 1. Make internal/capture compile

- [x] 1.1 Choose a published gortsplib version (v5.x, see proposal.md) and update `go.mod`/`go.sum`; verify `go list -m github.com/bluenviron/gortsplib/v5` shows the chosen version
- [x] 1.2 Port `internal/capture/rtsp_reader.go` and `internal/capture/screenshot.go` to the chosen API, keeping the exported types and functions used by `internal/api`, `internal/motion` and `internal/encoder` unchanged; verify `go build ./internal/capture ./internal/api ./internal/motion ./internal/encoder` succeeds
- [x] 1.3 Remove the gortsplib v4 subpackage imports; verify `grep -rn "gortsplib/v4" --include=*.go .` returns nothing
- [x] 1.4 Update the gortsplib section of `docs/KNOWN_ISSUES.md` to show it resolved and name the version used; verify the section no longer says the build fails

## 2. Configuration from environment variables

- [x] 2.1 Extend `Config` and `LoadConfig` in `internal/app/config.go` with the HTTP port, HLS paths, auth method and credentials, S3 credentials and encryption key path, using `UPPER_SNAKE` names and the defaults from `config.example.yaml`; verify a unit test covers default and overridden values for each variable
- [x] 2.2 Document every environment variable from `LoadConfig` in `docs/deployment.md`; verify each `os.LookupEnv`/`getEnvOrDefault` key in `internal/app/config.go` appears in the doc (grep check)

## 3. Assembly in internal/app

- [x] 3.1 Split `New` into `New()` (loads configuration only) and `Init(ctx)` (opens the database, builds services, middleware, routes and the HTTP server); verify a unit test shows `New()` does not open the database and `Init()` returns an error for an invalid `DB_PATH`
- [x] 3.2 Replace the `501 Not implemented` stubs in `internal/app/router.go` with handlers from `internal/api`; verify a router test calls every route registered by `RegisterRoutes` and none returns 501
- [x] 3.3 Use `Config.HTTPPort` for the server address instead of `:8080`; verify a unit test asserts `Addr == ":" + HTTP_PORT`
- [x] 3.4 Fix the compile errors in `internal/auth/auth.go` (`NewDecoder` placeholder → `encoding/json`, `http.SameSiteLax` → `http.SameSiteLaxMode`); verify `go build ./internal/auth` and `go vet ./internal/auth` pass
- [x] 3.5 In `Init`, open the database with `internal/db.New`, create the config store, S3 uploader, archive manager, auth manager (basic or OIDC), API server and HLS handler from `Config` (same wiring as the old `cmd/rtsprec/main.go`); verify a unit test runs `Init` against a temporary `DB_PATH` and gets an error for an unusable `ENCRYPTION_KEY_PATH` or `DB_PATH`
- [x] 3.6 Move the camera capture pipeline (`startCameraCapture`, `captureCamera`) from `cmd/rtsprec/main.go` into `internal/app` as `StartCameras`, started by `Run` and stopped by `Shutdown`; verify a unit test runs `StartCameras` against a database with no enabled cameras and returns without error
- [x] 3.7 Serve the static `./static` directory and the logout handler from `Init` with the same routes as the old binary; verify a router test requests `/`, `POST /api/auth/logout` and an HLS route and gets a non-501 response
- [x] 3.8 Register `POST /api/auth/login` and `POST /api/auth/logout` outside the authentication middleware; verify a test logs in without a token and gets a token, and logout returns 200 without a token
- [x] 3.9 Replace literal `*` wildcards with `{id}` / `{camera}` / `{type}` / `{file}` patterns and read the camera ID with `r.PathValue`; verify a test requests `/api/cameras/{id}/status` and receives the handler's own "camera not found" response

## 4. Binary entrypoint

- [x] 4.1 Rewrite `cmd/rtsprec/main.go` to call `app.New()`, `app.Init()`, `app.Run()`, handle SIGINT/SIGTERM and call `app.Shutdown` with a timeout; remove the `-config` flag; verify `go vet ./cmd/rtsprec` passes and `go build -o rtsprec ./cmd/rtsprec` succeeds
- [x] 4.2 Start the built binary with environment variables set and request `GET /api/cameras` with the configured credentials; verify the response is not 501 and the process exits cleanly on SIGTERM

## 5. Verification

- [x] 5.1 Run `make lint` and `make test-unit`; verify both pass
- [x] 5.2 Confirm that `.github/workflows/backend.yml` build step now fails on build errors instead of logging a non-blocking warning; verify the step has no `exit 0` fallback

## 6. Remove unused files

- [x] 6.1 Delete `internal/handler`, `internal/logging`, `internal/app/database.go`, `internal/app/dependencies.go`, `internal/app/storage.go`, `config.example.yaml`, `deployments/rtsprec.service` and empty `pkg/`; verify `go build ./cmd/rtsprec` and `go test ./internal/app` still pass
- [x] 6.2 Remove references to deleted packages and files from `Makefile`, `.github/workflows/backend.yml`, `README.md`, `docs/deployment.md`, `docs/KNOWN_ISSUES.md`; verify `grep -rn "internal/handler\|config.example.yaml\|rtsprec.service" --exclude-dir=openspec --exclude-dir=.git .` returns only ARCHITECTURE.md descriptions

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
