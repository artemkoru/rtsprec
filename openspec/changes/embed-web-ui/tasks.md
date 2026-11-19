# Tasks

## 1. Embed the UI

- [x] 1.1 Add `web/embed.go` (package `web`) with `//go:embed all:dist` and an exported `FS`; add the placeholder `web/dist/.gitkeep` and a negation in `web/.gitignore` so the placeholder is tracked while built files stay ignored; verify `go build ./...` and `go vet ./web` pass on a clean checkout without Node
- [x] 1.2 Add an SPA handler in `internal/app`: serve an existing embedded file, otherwise serve `index.html` with 200; paths under `/api/` and `/hls/` never fall back to `index.html`; verify a Go test covers: existing asset served, `/live` returns `index.html`, `/api/unknown` returns 404 without HTML

## 2. Authentication boundary

- [x] 2.1 Change `internal/app/router.go` so `/api/*` (except login and logout) and `/hls/*` go through `AuthMiddleware`, and the SPA handler serves everything else without authentication; verify a router test: `GET /` and `GET /live` return 200 without a token, `GET /api/cameras` returns 401 without a token and 200 with a token, `POST /api/auth/login` still works without a token
- [x] 2.2 Update `openspec/specs` on archive only (delta is in this change); verify `openspec validate embed-web-ui` passes

## 3. Build and Docker

- [x] 3.1 Add a `web-build` target to the Makefile (`npm ci` and `npm run build` in `web/`, removes stale `web/dist` contents except `.gitkeep` first) and make `build` depend on it; verify `make build` produces a binary that serves `/` with the built `index.html`
- [x] 3.2 Add a Node 20 build stage to the root `Dockerfile` that runs `npm ci` and `npm run build` before the Go build, copying `web/dist` into the Go build context; verify `make docker-build` succeeds and the image serves `/` with the UI
- [ ] 3.3 Keep `make docker-up` checks working: it waits on `/api/cameras` returning 401; verify it still passes

## 4. CI

- [x] 4.1 In `.github/workflows/backend.yml`, add `actions/setup-node@v4` (Node 20) to the `build` job before `make build`; verify the YAML parses and the job runs `make build`
- [x] 4.2 Keep `lint` and `unit-test` jobs without Node; verify they still pass on the placeholder

## 5. Documentation

- [x] 5.1 Update README "Local Run": the UI is served on `http://localhost:8080/` by the Docker stack and by the built binary; the Vite dev server on port 5173 is still available for development; verify the commands match the Makefile targets
- [x] 5.2 Update `docs/deployment.md`: the binary no longer needs a `./static` directory; verify no reference to `./static` remains except in the change history

## 6. Verification

- [x] 6.1 Run `make lint` and `make test-unit` without Node installed on PATH for Go steps; verify both pass
- [x] 6.2 Browser check against the Docker stack on `http://localhost:8080/`: the login page loads without a token, login reaches `/live`, Live, Archive and Settings render, and a direct visit to `/archive` after login works; verify all checks pass
- [ ] 6.3 Open a pull request and verify the backend and frontend CI workflows pass

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied, so the `authentication` delta is merged into `openspec/specs/authentication/spec.md`.
- Verify the archived result.
