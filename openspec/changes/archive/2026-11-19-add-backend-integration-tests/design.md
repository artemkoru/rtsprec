# Design

## Context

The backend binary (`cmd/rtsprec`) builds the application with `app.New`, `app.Init` and `app.Run`. It reads configuration only from environment variables and needs a reachable S3 bucket at startup, because `storage.NewS3Uploader` calls `HeadBucket` in `Init`. It serves an HTTP API on `HTTP_PORT`, protects everything except login and logout with a bearer token, and handles SIGTERM gracefully. See proposal.md - Why.

The repository has no Dockerfile. Go tests in this repository use only the standard library and `net/http/httptest`, with no external services. `make test-integration` currently aliases the unit tests.

## Goals / Non-Goals

**Goals:**
- Test the built backend as a separate process, in a container, through its HTTP interface.
- Cover authentication, the protected route boundary, two read endpoints, the unknown-camera response, logout and shutdown.
- Run the same command locally and in CI through `make test-integration`.

**Non-Goals:**
- Camera capture, motion detection, HLS segments and S3 uploads of video. They need a real RTSP source and are covered by the separate real-stream test plan in `docs/testing.md`.
- Frontend tests.
- Changing the backend's behavior or its routes to make tests easier.

## Decisions

- **Separate Go module in `integration/`.** Testcontainers and its dependencies stay out of the main `go.mod`, so the production build graph does not change. Alternative: a `_test.go` file in the root module with a build tag. Rejected because it would pull container dependencies into `go.sum` and `go mod tidy` for the product.
- **Backend image from a multi-stage `Dockerfile` at the repository root.** Build stage uses `golang:1.26.0` and runs `go build -o /out/rtsprec ./cmd/rtsprec`; runtime stage copies the binary onto a minimal base image and runs it as a non-root user. Alternative: `go run` inside a generic Go container. Rejected because it would test a different process than the one shipped.
- **Registry login (superseded: not needed).** The SeaweedFS image is public on Docker Hub, so no registry login is required locally or in CI. The earlier MinIO-specific plan was:  MinIO images are not pullable anonymously: `quay.io/minio/minio:*` returns 401 and `docker.io/minio/minio` returns "pull access denied". The tests therefore need an authenticated registry session. Locally the developer runs `docker login` once; credentials stay in Docker's credential store. In CI the integration job logs in with repository secrets before `make test-integration`. Tests never read or print credentials.
- **SeaweedFS (`chrislusf/seaweedfs`, Docker Hub) for S3 instead of MinIO and instead of a stub server (decided 2026-11-19).** The MinIO images were not pullable even after `docker login` to Docker Hub (`pull access denied`), so MinIO was dropped. SeaweedFS runs `server -s3 -s3.domainName=seaweed`; virtual-host bucket addressing resolves through the network alias `test-bucket.seaweed`.
- **Original S3 choice (superseded).** The backend calls `HeadBucket` at startup and uploads objects later. MinIO implements both with real S3 semantics. Alternative: a small HTTP stub that returns 200. Rejected for the first version because it would hide signature and endpoint problems; it can be added later if MinIO proves too slow in CI.
- **Configuration through environment variables only**, with `AWS_ENDPOINT_URL` pointing at the MinIO container and fixed `AUTH_BASIC_*` credentials set by the test. The database path is inside the container, so each test run starts with a fresh database.
- **Test driver: standard `testing` package with a small helper that starts containers and waits for the health of the port.** Readiness is checked by polling `GET /api/cameras` until it returns 401 (the server is up and auth is enforced), with a timeout. Alternative: a log-based wait. Rejected because the HTTP check does not depend on log wording.
- **`make test-integration` runs `go test` inside `integration/`.** It checks that Docker is reachable first and prints a clear message if not, so a missing Docker is not reported as a test failure.

## Risks / Trade-offs

- [Docker is required locally] → The Makefile target fails early with an explanation. Developers without Docker can still run `make test-unit`.
- [MinIO image pull and container start add time to CI] → Use a fixed image tag; start one MinIO container per test package, not per test.
- [Container tests can leave containers behind after a failure] → Use testcontainers cleanup registered with `t.Cleanup`.
- [`Init` builds the database at `DB_PATH`; a stale file would change test results] → Each test run uses a fresh temporary path inside the container.
- [The backend image must match the Go version in CI] → Pin `golang:1.26.0` in the Dockerfile and in `backend.yml`.
- [Route patterns and auth behavior are asserted against current behavior] → The tests document the current contract; any intentional change updates them in the same change.

## Migration Plan

1. Add `Dockerfile` and verify that `docker build .` produces a runnable image.
2. Add `integration/` module with the helper and the first scenarios.
3. Point `make test-integration` at the module and remove the alias to `test-unit`.
4. Verify locally on a machine with Docker, then in CI on a pull request.

Rollback: remove the `integration/` module and restore the previous `test-integration` target; no runtime code changes are involved.

## Open Questions

- None that change the approach. The MinIO versus stub decision is recorded above as a default and can be revisited if CI time is unacceptable.
