# Spec Delta

## MODIFIED Requirements

### Requirement: Protect web UI endpoints
The system SHALL require authentication for all API and media endpoints: every `/api/*` path except login and logout, and every `/hls/*` path. Unauthenticated requests to these endpoints SHALL be rejected with 401 Unauthorized. The web UI shell and its static assets SHALL be served without authentication; the UI SHALL redirect to the login page when no valid session exists.

#### Scenario: Unauthenticated access attempt
- **WHEN** unauthenticated user visits `/live` or `/archive`
- **THEN** system serves the UI shell, and the UI redirects to `/login`

#### Scenario: Protected API endpoints
- **WHEN** unauthenticated request hits `/api/cameras` or `/api/config`
- **THEN** system returns 401 Unauthorized

#### Scenario: Protected media endpoints
- **WHEN** unauthenticated request hits `/hls/{camera}/playlist.m3u8`
- **THEN** system returns 401 Unauthorized

#### Scenario: Public login and logout
- **WHEN** unauthenticated request hits `POST /api/auth/login` or `POST /api/auth/logout`
- **THEN** system handles the request without requiring a session
