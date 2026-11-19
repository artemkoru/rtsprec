# Design

## Context

The router (`internal/app/router.go`) registers the API and HLS routes on one mux and wraps that mux in `AuthMiddleware`. The static handler `http.FileServer(http.Dir("./static"))` is also on that mux, so the UI is protected. The frontend is a Vite build in `web/dist` (`index.html` plus `assets/`). `web/.gitignore` ignores `dist`. The frontend routes are client-side (`/live`, `/archive`, `/settings`, `/login`); the React app redirects to `/login` when `GET /api/cameras` returns 401. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- The backend binary serves the UI on the same port as the API, without a separate directory.
- A direct visit or reload of any client route returns the UI, not a 404.
- Unauthenticated users can load the UI and are sent to login by the UI itself. API and HLS stay protected.
- `go build`, `go vet` and `go test` work without Node.

**Non-Goals:**
- Changing pages, login or cookie behavior (except the cookie is already `Path=/`).
- Server-side rendering or a different frontend build tool.
- Serving the UI from a CDN.

## Decisions

- **`go:embed` from a package in `web/`.** `web/embed.go` (`package web`) embeds the `dist` directory next to it. Go embeds only from the package directory and below, so the embed must live in `web/`. Alternative: copy `web/dist` into `internal/`. Rejected: an extra copy step that can go stale. Pattern: `//go:embed all:dist`, so the placeholder file that starts with a dot is included.
- **Placeholder `web/dist/.gitkeep` is tracked.** `web/.gitignore` gets a negation for `dist/.gitkeep`. Without it `go build` fails on a clean checkout. The real `index.html` and assets are built by `npm run build` and stay ignored. With only the placeholder, the UI route returns 404 and the API still works, which is the documented developer state.
- **SPA handler.** For a request that is not an API or HLS path, serve the matching file from the embedded FS if it exists; otherwise serve `index.html` with status 200. Paths under `/api/` and `/hls/` never fall back to `index.html`, so an unknown API path returns 404 from the API mux, not HTML.
- **Auth boundary moves to route groups.** `/api/*` (except `POST /api/auth/login` and `POST /api/auth/logout`) and `/hls/*` go through `AuthMiddleware`. Everything else is served by the SPA handler without authentication. Alternative: keep the middleware on everything and whitelist static paths. Rejected: whitelisting by file path is easier to get wrong than a route boundary.
- **Build order.** `make build` depends on `web-build` (`npm ci` and `npm run build` in `web/`), which writes `web/dist`. `make docker-build` gets the same effect from a Node stage in the Dockerfile that builds `web/dist` before the Go stage copies it. CI's `build` job installs Node 20 and runs `make build`. The `go test` jobs run without the frontend build and see only the placeholder.
- **Dev server unchanged.** `npm run dev` on port 5173 keeps proxying `/api` and `/hls`; the embedded UI is for the built binary and the Docker image.

## Risks / Trade-offs

- [Public UI shell reveals the frontend code] → The bundle contains no secrets; all data comes from the protected API. Accepted.
- [Clean checkout builds a binary with no UI] → `make build` always runs the frontend build first; the placeholder is only for plain `go build`. Documented in the Makefile help text.
- [Node stage increases image build time] → Layer caching for `npm ci` in the Dockerfile.
- [A stale `dist` could be embedded] → `make build` removes and rebuilds `web/dist` before compiling.
- [Existing `./static` deployments] → The binary ignores `./static` after the change; deployment notes say so.

## Migration Plan

1. Land the change with the placeholder; `go build` and `make test-unit` keep passing without Node.
2. Build with `make build` or `make docker-build`; verify `http://localhost:8080/` serves the UI.
3. Rollback: revert the router and embed changes; the API behavior returns to the previous state.

## Open Questions

- None that change the approach.
