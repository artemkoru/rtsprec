# Tasks

## 1. API helper

- [x] 1.1 Add `postJSON(path, body)` to `web/src/api.js`: sends `POST` with `Content-Type: application/json`, `credentials: 'same-origin'`, redirects to `/login` on 401, throws `ApiError` with the status for other errors; verify `npm run lint` and `npm run build` pass in `web/`

## 2. Add form

- [x] 2.1 Bind the four fields in the Cameras tab of `web/src/pages/Settings.jsx` to state; add a submit handler that validates name and RTSP URL, marks missing fields as required, and sends the camera with the numeric defaults from the README (sensitivity 50, motion timeout 120, duration 15, screenshot interval 0); verify `npm run lint` passes and an empty submit sends no request
- [x] 2.2 After a successful POST clear the form and reload the camera list with `GET /api/cameras`; show the server error next to the form on failure and keep the entered values; verify in the browser (task 3.2) that the new camera appears without a page reload

## 3. Browser tests

- [x] 3.1 Add Playwright to `web/` as a dev dependency, a `test:e2e` script, and `web/tests/cameras.spec.js`; the tests log in through `/login` using `E2E_USERNAME` and `E2E_PASSWORD` (defaults match `docker compose`) and read the backend URL from `E2E_BASE_URL` (default `http://localhost:5173`); verify `npm run test:e2e` runs and fails with a clear message when the backend is unreachable
- [x] 3.2 Test: adding a camera with name and RTSP URL shows it in the list; verify the test passes against the local stack
- [x] 3.3 Test: submitting with an empty name or RTSP URL sends no request and shows the required marker; verify the test passes
- [x] 3.4 Test: a server error (intercepted response with status 400) keeps the entered values and shows the error next to the form; verify the test passes

## 4. CI

- [x] 4.1 Add a `e2e` job to `.github/workflows/frontend.yml` that starts the backend with `docker compose up -d` and runs `npm run test:e2e` after installing Chromium; verify the workflow YAML parses and the job runs on pull requests to `main`

## 5. Verification

- [x] 5.1 Run `npm run lint`, `npm run build` and `npm run test:e2e` locally; verify all pass
- [ ] 5.2 Open a pull request and verify the frontend CI workflow, including the new e2e job, passes

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied, so the `web-configuration` delta is merged into `openspec/specs/web-configuration/spec.md`.
- Verify the archived result.
