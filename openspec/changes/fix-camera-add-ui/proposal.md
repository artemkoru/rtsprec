# Proposal

## Why

On the Settings page the "Add Camera" form is a static mock: its fields are not bound to state and its button has no handler. Pressing the button does nothing, so the admin cannot add a camera from the web UI. The backend endpoint `POST /api/cameras` works and returns 201, so the gap is only in the frontend. The `web-configuration` spec already requires adding cameras through the web form; the UI does not do it.

## What Changes

- The Cameras tab of the Settings page gets a working add form: fields for name, RTSP URL, username and password are bound to state, and the Add button sends `POST /api/cameras` with the session cookie.
- On success the form is cleared and the camera list is reloaded, so the new camera appears without a page reload.
- On failure the page shows the error next to the form. A 401 sends the user to the login page through the existing API helper.
- The form checks the required fields (name and RTSP URL) before sending.
- Browser tests (Playwright) cover adding a camera, validation, and an API error. Unit-level checks for the helper are not added in this change.

Out of scope: starting capture for a camera added at runtime (separate change), Edit and Delete buttons, the motion, storage and scheduling tabs.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `web-configuration`: requirement "Configure camera sources" gains scenarios for adding a camera from the form, for validation of required fields, and for showing a server error.

## Impact

- `web/src/pages/Settings.jsx` (form state, submit handler, list refresh), possibly a small helper in `web/src/api.js` for POST with JSON.
- New browser test setup: Playwright as a dev dependency in `web/`, a test script, and a test file under `web/tests/`. The Playwright browser runs against the Vite dev server with the backend proxy.
- No backend change. Capture of a newly added camera still requires a backend restart (documented as a limitation until the separate change).
