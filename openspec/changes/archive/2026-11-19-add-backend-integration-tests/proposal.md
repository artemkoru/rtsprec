# Proposal

## Why

The backend has only unit tests. `make test-integration` currently runs the same unit tests again, so nothing checks that the built binary starts, serves its HTTP API, enforces authentication and shuts down cleanly. Those behaviors were verified only by hand during the `use-app-entrypoint` change. Without automated coverage, changes to `internal/app`, routing or authentication can break the running service without any failing test.

## What Changes

- New Go module `integration/` (separate `go.mod`) that starts the real backend in a container with testcontainers-go and sends HTTP requests to it.
- New `Dockerfile` at the repository root that builds `./cmd/rtsprec` into a runtime image. The backend is configured only through environment variables, as documented in `docs/deployment.md`.
- The tests start an S3-compatible SeaweedFS container (`chrislusf/seaweedfs`, Docker Hub) and point the backend at it through `AWS_ENDPOINT_URL`, so `Init` can pass its bucket check without a real AWS account.
- First-version scenarios: login returns a token; protected routes return 401 without a token and 200 with one; `GET /api/cameras`; `GET /api/config/{type}`; `GET /api/cameras/{id}/status` for an unknown camera returns the handler's "camera not found"; logout returns 200; the container stops cleanly on SIGTERM.
- `make test-integration` runs the `integration/` module. It fails with a clear message when Docker is not available.
- CI `integration-test` job runs `make test-integration` on pull requests to `main`. The unit-test duplication in that job is removed.

Proposed defaults for the open questions, to be confirmed before implementation:
- An S3-compatible server rather than a stub, because it behaves like S3 for `HeadBucket` and uploads. SeaweedFS replaces MinIO because MinIO images are not pullable from Docker Hub.
- Tests live in `integration/` as their own module, so their dependencies do not enter the main `go.mod`.

## Capabilities

### New Capabilities

<!-- None: the tests check existing behavior, they do not introduce new capabilities. -->

### Modified Capabilities

<!-- None: no requirement changes. -->

## Impact

- New files: `integration/go.mod`, `integration/*_test.go`, `Dockerfile`.
- `Makefile`: `test-integration` target changes from an alias of `test-unit` to running the `integration/` module.
- `.github/workflows/backend.yml`: `integration-test` job keeps `make test-integration`; it now needs Docker, which `ubuntu-latest` provides.
- No change to runtime behavior of `cmd/rtsprec` or `internal/`.
- Test runtime depends on pulling the backend build context and the MinIO image; first run is slower.
