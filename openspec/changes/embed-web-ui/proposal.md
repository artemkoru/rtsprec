# Proposal

## Why

The backend binary and its Docker image serve only the API. Opening `http://localhost:8080/` returns `401 unauthorized: missing session token`, because the root path is a static file handler behind the authentication middleware, and the image contains no UI. Users have to know to open the Vite dev server on port 5173. A single binary that serves the UI itself makes deployment simpler: one port, one artifact, and the same UI in Docker and on a server.

## What Changes

- The production UI build (`web/dist`) is embedded into the backend binary with `go:embed`, so the binary serves the UI without a separate `./static` directory.
- Client-side routes (`/live`, `/archive`, `/settings`, `/login`) return the UI shell `index.html` when no file matches, so a direct visit or reload works.
- **BREAKING for the auth boundary**: the UI shell and its static assets are served without a session. Only the API and HLS endpoints require a session. The UI redirects to `/login` when there is no session, as it already does. Login and logout stay public.
- `make build` and `make docker-build` build the frontend first. The Dockerfile gets a Node build stage. CI's build job installs Node 20 before building.
- Go builds without a frontend build still work: the embedded directory contains a committed placeholder, so `go test` and `go vet` do not need Node.

Out of scope: changing the frontend's pages, the login flow, cookie settings, and the embedding of any other static asset.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `authentication`: requirement "Protect web UI endpoints" changes. The UI shell becomes public and the API and HLS endpoints remain protected.

## Impact

- `web/embed.go` (new, package `web`), `web/dist/` placeholder, `internal/app/router.go` (static and SPA handling, auth boundary), a new test file in `internal/app`.
- `Makefile` (`build` depends on a frontend build target), `Dockerfile` (Node stage), `.github/workflows/backend.yml` (Node setup in the build job), `web/.gitignore` (the placeholder must be tracked).
- Deployment: the binary no longer needs a `./static` directory next to it. Existing deployments that rely on `./static` keep working only if the embedded UI replaces it; this is documented in `docs/deployment.md`.
