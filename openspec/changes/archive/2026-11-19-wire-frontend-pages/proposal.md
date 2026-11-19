# Proposal

## Why

The web UI in `web/` is still the Vite template: `App.jsx` shows "Get started" and the real pages `Live.jsx`, `Archive.jsx` and `Settings.jsx` are not imported anywhere. There is no router and no login screen. The backend protects everything except login and logout with a session token, so a browser cannot use the pages even when they are wired. Embedding the frontend into the backend binary is only useful once there is a working UI to embed, so the UI comes first.

## What Changes

- `web/src/App.jsx` is replaced with a navigation shell that renders `Live`, `Archive` and `Settings`.
- Client-side routing is added so each page has its own path (`/live`, `/archive`, `/settings`) and browser back and forward work.
- A login screen posts to `POST /api/auth/login` with username and password. On success the session token is kept in the cookie the backend sets, and the user is taken to the live page. On a 401 the screen shows an error. A logout control calls `POST /api/auth/logout` and returns to the login screen.
- Requests that return 401 send the user back to the login screen.
- The pages keep their current API calls: `GET /api/cameras`, `GET /api/archive?camera_id=&date=`, and `GET /hls/{camera}/playlist.m3u8` for video. No new backend endpoints are added.

- One backend fix is included because the UI cannot work without it: the login response sets the `session_token` cookie with `Path: "/"`. Without it the browser scopes the cookie to `/api/auth` and every other API call returns 401 after login. Found during browser testing.

Out of scope for this change: embedding `web/dist` into the backend binary, changing how static files are served, and any other backend route or auth behavior change. These are decided in a later change.

## Capabilities

### New Capabilities

<!-- None: the frontend has no spec in openspec/specs/. -->

### Modified Capabilities

<!-- None: backend behavior is unchanged. -->

## Impact

- `web/src/` only: `App.jsx`, a new login page, routing, a small API helper for 401 handling, and the pages' imports.
- `web/package.json`: one new dependency for routing (choice recorded in design.md).
- Frontend CI (`frontend.yml`) runs lint and build; these must keep passing.
- Backend: one line in `internal/auth/auth.go` (cookie `Path: "/"`) and a check for it in `internal/app/app_test.go`. No other backend change. The HLS player receives the session cookie from the login response because the page and the stream are on the same origin.
