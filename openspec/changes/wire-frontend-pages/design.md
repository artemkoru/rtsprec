# Design

## Context

The frontend is a Vite + React 19 app in `web/` with no router and no HTTP helper. Pages call `fetch('/api/...')` with relative URLs and no `Authorization` header. The backend (`internal/app/router.go`) registers login and logout as public routes and puts every other route, including `/` and static files, behind `AuthMiddleware`. That middleware accepts a bearer token, a `session_token` cookie, or a `token` query parameter. Login sets the `session_token` cookie (`HttpOnly`, `Secure`, `SameSite=Lax`). See proposal.md - Why.

Because the HLS player in `Live.jsx` loads `/hls/{camera}/playlist.m3u8` as a plain media URL, it cannot send a bearer header. It depends on the cookie set by login, so UI and API must be served from the same origin.

## Goals / Non-Goals

**Goals:**
- Each page is reachable at its own path and survives reload.
- A user who is not logged in sees the login screen and nothing else.
- Every API call that returns 401 leads back to login.

**Non-Goals:**
- Embedding `web/dist` into the backend binary (later change).
- Changing backend routes, auth rules, or cookie settings.
- Visual redesign of the pages.
- Multi-user management or password change UI.

## Decisions

- **Routing library: `react-router-dom` v7.** It is the common choice for React SPAs, works with Vite's dev server and with a static file server that falls back to `index.html`. Alternative: state-based tabs in `App.jsx`. Rejected because paths are needed for reload and for the later embedded build. Version is pinned in `package.json` and the lockfile is updated in the same change.
- **Session state comes from the cookie, not from localStorage.** Login sets the cookie; the frontend never stores the token itself. The backend `session_token` cookie is `HttpOnly`, so JavaScript cannot read it, and that is correct. The frontend learns that a session is valid by calling `GET /api/cameras`: 200 means logged in, 401 means not. Alternative: read the token from the login JSON response and keep it in memory. Rejected: the HLS player would not get it, and a reload would lose it.
- **One fetch helper** (`web/src/api.js`) wraps `fetch`, sends `credentials: 'same-origin'`, and on 401 navigates to `/login`. Pages call the helper instead of `fetch` directly. This keeps the 401 handling in one place.
- **Login page at `/login`**, outside the navigation shell. Logout is a button in the shell that calls `POST /api/auth/logout` and then navigates to `/login`. The backend logout response is fixed JSON and does not invalidate the session, so the frontend cannot rely on logout to end the session; this is noted as a known limitation, not fixed here.
- **Default route `/` redirects to `/live`.** Unknown paths show a not-found message.
- **Dev proxy in `vite.config.js`** forwards `/api` and `/hls` to the backend (`http://localhost:8080`) so `npm run dev` works against a running backend on the same origin.

## Risks / Trade-offs

- [Session cookie is `Secure`] → Browsers store `Secure` cookies on `http://localhost`, so local development works. On any other plain-HTTP host the cookie is dropped and login appears to succeed but later calls return 401. Mitigation: document this in the change tasks and serve the UI over HTTPS elsewhere; the cookie settings are backend behavior and are not changed here.
- [Logout does not invalidate the server session] → The frontend clears its view and returns to login, but a copied token still works until it expires (24 h). Mitigation: recorded as a known issue; a backend fix is a separate change.
- [Archive endpoint returns a timeline, not video files] → The archive page shows the timeline JSON as the backend returns it; the page layout must match the real response shape, which is checked against a running backend in tasks.
- [Routing dependency adds bundle size] → Acceptable for this UI size; checked in the build output.

## Migration Plan

1. Land the change behind the existing routes; nothing else in the backend changes.
2. Rollback: revert the commit; the backend and the template frontend are unaffected.

## Open Questions

- None that change the approach. Whether the logout bug is fixed on the backend is decided in a separate change.
