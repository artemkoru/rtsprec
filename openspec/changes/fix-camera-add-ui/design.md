# Design

## Context

`web/src/pages/Settings.jsx` renders the Cameras tab: a form with four `<input>` elements and an "Add Camera" `<button>`, with no state and no handler, and a list of cameras loaded with `apiFetch('/api/cameras')`. The helper `web/src/api.js` sends same-origin requests with the session cookie and turns 401 into a redirect to `/login`; for other errors it throws `ApiError` with the status. The backend endpoint `POST /api/cameras` decodes a JSON body with `name`, `rtsp_url`, `username`, `password`, `motion_sensitivity`, `motion_timeout`, `video_duration_minutes`, `screenshot_interval_minutes`, and returns 201 on success. The frontend has no test runner yet. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- An admin adds a camera from the Settings page and sees it in the list without reloading the page.
- Required fields are checked before the request; server errors are shown next to the form.
- The behavior is covered by browser tests that run in CI.

**Non-Goals:**
- Starting capture for the new camera (separate change).
- Edit and Delete for cameras, other Settings tabs.
- Changing backend validation or the API contract.

## Decisions

- **Controlled form in Settings.jsx.** One state object for the four fields, one handler for change, one submit handler. Alternative: a separate `CameraForm` component. Rejected for now: the page is small and a separate component adds files without benefit; it can be split later.
- **Required fields: name and RTSP URL.** The backend does not reject empty values in the handler itself, but a camera without a name or URL is useless. Username and password stay optional, as the form already labels them. Alternative: rely on the backend only. Rejected: the error would come back as a generic failure.
- **Request through `apiFetch`-style helper.** A new helper `postJSON(path, body)` in `api.js` keeps the 401 handling and cookie behavior in one place. The numeric fields use the backend defaults the README documents: sensitivity 50, motion timeout 120, duration 15, screenshot interval 0, so the form sends a complete body.
- **Refresh by re-fetching the list.** After a successful POST the page calls `GET /api/cameras` again. Alternative: append the response to local state. Rejected: the response shape of POST is not part of this change, and re-fetching keeps the list authoritative.
- **Playwright for tests.** Playwright drives the real browser against the Vite dev server with the backend behind the proxy, so the test covers the same path a user takes. Tests log in through the login page first. Alternative: Vitest with Testing Library. Not chosen: the user asked for browser-level tests; a unit layer can be added later.
- **Test data isolation.** Each test adds a camera with a unique name (timestamp suffix). The backend database is not reset between runs; the tests do not rely on the camera count.

## Risks / Trade-offs

- [Playwright needs a browser binary in CI] → Install with `npx playwright install --with-deps chromium` in the CI job; the job time grows.
- [Tests depend on a running backend and login credentials] → The test script reads the backend URL and credentials from environment variables with defaults matching `docker compose`; it fails with a clear message when the backend is unreachable.
- [Camera added in the test is not captured until backend restart] → Expected; the tests check only the API and UI behavior, not recording.
- [Cookie `Secure` on non-localhost hosts] → Tests run on `localhost`, where the cookie works.

## Migration Plan

1. Land the form and tests; no data migration.
2. Rollback: revert the commit; the static mock returns.

## Open Questions

- None that change the approach.
