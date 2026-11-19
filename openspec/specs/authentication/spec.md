# authentication Specification

## Purpose
Secure access to web UI and API endpoints using OIDC (Google, Yandex) or Basic Auth, allowing administrators to control who can view cameras and modify settings.

## Requirements

### Requirement: Support OIDC authentication (Google, Yandex)
The system SHALL allow authentication via OIDC providers (Google, Yandex) with configurable client ID/secret and redirect URI.

#### Scenario: Google OIDC login
- **WHEN** unauthenticated user visits web UI
- **THEN** system redirects to Google login flow, user logs in, redirects back with valid session

#### Scenario: Yandex OIDC login
- **WHEN** admin configures Yandex OIDC provider with client ID/secret
- **THEN** users can authenticate via Yandex instead of Google

#### Scenario: OIDC configuration
- **WHEN** admin enters OIDC provider details (client ID, secret, discovery URL)
- **THEN** system validates configuration and enables OIDC login option

#### Scenario: Session creation
- **WHEN** OIDC authentication succeeds
- **THEN** system creates session token (JWT or cookie) valid for 24 hours

#### Scenario: User identifier
- **WHEN** user authenticates via OIDC
- **THEN** system stores user's email/ID for audit/logging purposes

### Requirement: Support Basic Authentication
The system SHALL allow simple username/password authentication (Basic Auth) as fallback when OIDC is not configured.

#### Scenario: Basic Auth login
- **WHEN** user submits username and password
- **THEN** system verifies against stored credentials and creates session

#### Scenario: Credential storage
- **WHEN** admin sets username and password for Basic Auth
- **THEN** password is hashed (bcrypt or similar) before storage

#### Scenario: Multiple users
- **WHEN** Basic Auth is enabled
- **THEN** system supports multiple username/password pairs (e.g., admin1, admin2)

#### Scenario: Default credentials
- **WHEN** system starts with Basic Auth enabled but no credentials configured
- **THEN** system uses default credentials (documented, requires change on first login)

### Requirement: Manage authentication methods
The system SHALL allow admin to choose and configure which authentication method to use (OIDC or Basic Auth).

#### Scenario: Authentication method selection
- **WHEN** admin opens authentication settings
- **THEN** admin can enable OIDC (with provider selection) or Basic Auth

#### Scenario: OIDC + Basic Auth fallback
- **WHEN** both methods are configured
- **THEN** web UI shows both login options (OIDC button, username/password form)

### Requirement: Session management
The system SHALL create, validate, and expire sessions; prevent unauthorized access to protected resources.

#### Scenario: Session creation
- **WHEN** user successfully authenticates
- **THEN** system generates session token with expiration time (24 hours)

#### Scenario: Session validation on API calls
- **WHEN** authenticated user makes API request with session token
- **THEN** system validates token and allows request to proceed

#### Scenario: Session expiration
- **WHEN** session expires
- **THEN** next API request is rejected with 401 Unauthorized, user is redirected to login

#### Scenario: Logout
- **WHEN** user clicks logout
- **THEN** system invalidates session token and clears cookies/local storage

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

### Requirement: Protect configuration API
The system SHALL restrict configuration changes to authenticated users only, requiring valid session for `POST/PUT/DELETE` operations.

#### Scenario: Authenticated configuration change
- **WHEN** authenticated admin makes `POST /api/cameras` with new camera config
- **THEN** system validates session and applies configuration

#### Scenario: Unauthenticated configuration change attempt
- **WHEN** request to `POST /api/cameras` lacks valid session token
- **THEN** system returns 401 Unauthorized and rejects change

#### Scenario: Read-only access (future)
- **WHEN** future versions support user roles
- **THEN** viewers can read camera/archive without changing settings

### Requirement: Secure credential transmission
The system SHALL enforce HTTPS/TLS for all authentication exchanges and API calls involving credentials or session tokens.

#### Scenario: HTTPS enforcement
- **WHEN** system is exposed over network
- **THEN** web UI requires HTTPS; HTTP requests are redirected or rejected

#### Scenario: Cookie security
- **WHEN** session cookie is created
- **THEN** cookie is marked Secure (HTTPS only) and HttpOnly (no JavaScript access)

#### Scenario: OIDC over HTTPS
- **WHEN** OIDC flow is initiated
- **THEN** all exchanges with OIDC provider use HTTPS with certificate validation
