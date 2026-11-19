# Tasks

## 1. Routing and shell

- [ ] 1.1 Add `react-router-dom` (v7) to `web/package.json` with `npm install` in `web/`, and commit the updated lockfile; verify `npm ls react-router-dom` shows one version and `npm run build` succeeds
- [ ] 1.2 Replace the template in `web/src/App.jsx` with a shell that has navigation links to Live, Archive and Settings and a router outlet; verify `npm run lint` and `npm run build` pass
- [ ] 1.3 Add routes `/` (redirect to `/live`), `/live`, `/archive`, `/settings` and a not-found view; verify with `npm run dev` that each path renders its page and that reload on `/archive` keeps the page

## 2. API helper and 401 handling

- [ ] 2.1 Add `web/src/api.js` with one wrapper around `fetch` that sends `credentials: 'same-origin'`, parses JSON, and on HTTP 401 navigates to `/login`; verify with a unit-style check in the browser that a 401 response redirects to `/login`
- [ ] 2.2 Switch `Live.jsx`, `Archive.jsx` and `Settings.jsx` from direct `fetch` calls to the helper without changing request paths or query parameters; verify with `grep -rn "fetch(" web/src/pages` returns nothing

## 3. Login and logout

- [ ] 3.1 Add the login page at `/login` with username and password fields that posts to `POST /api/auth/login`; on success navigate to `/live`, on 401 show an error message; verify against a running backend that correct credentials reach `/live` and wrong credentials show the error
- [ ] 3.2 Add a logout control in the shell that calls `POST /api/auth/logout` and navigates to `/login`; verify the control returns the user to the login page
- [ ] 3.3 Check on load that a session exists by calling `GET /api/cameras`; a 401 shows the login page; verify with a cleared cookie that the shell is not shown

## 4. Verify against the backend

- [ ] 4.1 Run the backend locally with the environment from `docs/deployment.md` and a reachable S3 endpoint; verify the Live page loads the camera list and the HLS player requests `/hls/{camera}/playlist.m3u8` with the session cookie (browser network panel)
- [ ] 4.2 Verify the Archive page sends `GET /api/archive?camera_id=...&date=...` and renders the returned timeline without errors for a camera that has data
- [ ] 4.3 Document in `docs/deployment.md` (or `web/README.md`) that the session cookie is `Secure`, works on `http://localhost` and needs HTTPS on other hosts; verify the note matches `internal/auth/auth.go`

## 5. CI and verification

- [ ] 5.1 Run the frontend CI steps locally (`npm ci`, `npm run lint`, `npm run build`); verify all pass with Node 20
- [ ] 5.2 Open a pull request and verify the `Frontend CI` workflow passes

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
