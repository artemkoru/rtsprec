# Tasks

## 1. Project Setup and Dependencies

- [x] 1.1 Create Go project structure (`cmd/`, `internal/`, `pkg/`) and verify `go mod init` and `go.mod` exist
- [x] 1.2 Add Go dependencies: `github.com/bluenviron/gortsplib/v4`, AWS SDK S3, gorilla/websocket, React tooling; verify `go mod tidy` succeeds
- [x] 1.3 Create React project scaffold with Vite or Create React App; verify build produces output in `web/dist/`
- [x] 1.4 Set up systemd service file for Linux deployment; verify file is in `deployments/rtsprec.service`

## 2. Database Schema and Configuration Management

- [x] 2.1 Design and create SQLite schema for cameras, motion events, video metadata; verify migrations run and tables exist
- [x] 2.2 Implement config storage and encryption for S3 credentials; verify credentials are stored encrypted and can be decrypted
- [x] 2.3 Create config loading from YAML file and fallback to defaults; verify `config.yaml` is parsed correctly on startup

## 3. RTSP Capture and Stream Reading

- [x] 3.1 Implement RTSP reader goroutine with connection management; verify connects to test RTSP stream and reads frames
- [x] 3.2 Implement connection retry logic with exponential backoff; verify reconnection happens within expected time window
- [x] 3.3 Add frame buffering and channel-based communication between RTSP reader and downstream components; verify no goroutine leaks

## 4. Motion Detection

- [x] 4.1 Implement frame difference calculation (pixel-level threshold comparison); verify motion score is computed for test frame pairs
- [x] 4.2 Implement configurable sensitivity threshold (0-100 scale) and frame skipping (every N frames); verify sensitivity adjustment takes effect
- [x] 4.3 Add motion event logging to SQLite with timestamp, duration, camera ID; verify events persist across restarts

## 5. HLS Live Stream Generation

- [x] 5.1 Implement HLS segment generation (2-3 second segments in MPEG-TS format); verify segments are created and contain valid video
- [x] 5.2 Implement M3U8 playlist generation with sliding window of recent segments; verify playlist.m3u8 is valid and updates periodically
- [x] 5.3 Implement HTTP server endpoints for `/hls/<camera-id>/playlist.m3u8` and `/hls/<camera-id>/segment-*.ts`; verify files are served with correct MIME types

## 6. Video Encoding and S3 Upload

- [x] 6.1 Implement video encoding goroutine (5-minute MP4 chunks with source codec or H.264 fallback); verify chunks are valid MP4 files
- [x] 6.2 Implement S3 upload with configurable bucket, region, and path prefix; verify files appear in S3 with correct naming scheme
- [x] 6.3 Implement S3 upload retry logic (3 attempts, exponential backoff) and duplicate detection; verify failed uploads are retried and duplicates are skipped
- [x] 6.4 Implement presigned URL generation for authenticated archive playback; verify URLs are time-limited and valid

## 7. Archive Management and Indexing

- [x] 7.1 Implement video metadata indexing (file path, timestamps, duration) in SQLite on successful S3 upload; verify index can be queried by date/camera
- [x] 7.2 Implement timeline query API that returns consolidated segments/events for a date range; verify response includes correct markers for video and motion events
- [x] 7.3 Implement archive retrieval queries (by camera, date range, event filtering); verify queries are efficient for 1-30 day ranges

## 8. Scheduled Snapshots

- [x] 8.1 Implement screenshot capture goroutine triggered by configurable intervals; verify JPEGs are created at expected times
- [x] 8.2 Implement S3 upload for screenshots with path structure `<prefix>/<camera-id>/screenshots/YYYY/MM/DD/HH-MM-SS.jpg`; verify files appear in S3

## 9. Web API Endpoints (REST)

- [x] 9.1 Implement `/api/cameras` endpoints (GET list, POST add, PUT edit, DELETE remove); verify CRUD operations persist to database
- [x] 9.2 Implement `/api/cameras/<camera-id>/status` for real-time status (online/offline, frame count, motion events); verify updates are reflected within 5 seconds
- [x] 9.3 Implement `/api/config` endpoints for retrieving and updating settings (S3 credentials, motion sensitivity, recording duration); verify changes take effect immediately
- [x] 9.4 Implement `/api/archive/<camera-id>?date=YYYY-MM-DD` for timeline queries; verify response includes segments and events
- [x] 9.5 Implement `/api/events/<camera-id>?from=&to=` for motion event queries; verify filtering by time range works correctly
- [x] 9.6 Implement `/api/diagnostics` for logs and system status (CPU, memory, disk, last S3 upload); verify data reflects actual system state

## 10. Authentication and Authorization

- [x] 10.1 Implement OIDC provider configuration (Google, Yandex) with client ID/secret management; verify config is stored and validated
- [x] 10.2 Implement OIDC login flow with redirect URI and session token creation; verify user logs in via Google/Yandex and receives valid session
- [x] 10.3 Implement Basic Auth as fallback with username/password hashing (bcrypt); verify credentials are validated correctly
- [x] 10.4 Implement session validation middleware for all protected endpoints (`/api/*`, `/live`, `/archive`); verify unauthenticated requests return 401
- [x] 10.5 Implement session expiration (24 hours) and logout; verify expired sessions are rejected and new login is required
- [x] 10.6 Enforce HTTPS for all authentication exchanges and set Secure/HttpOnly flags on session cookies; verify cookies are secure in browser

## 11. Web UI - React Frontend

- [x] 11.1 Create React component structure (pages for Live, Archive, Settings); verify project builds without errors
- [x] 11.2 Implement Live View page with HLS player (HTML5 video or media.js) and camera selector; verify live stream plays correctly
- [x] 11.3 Implement Archive page with date picker and timeline visualization; verify timeline shows segments and events for selected date
- [x] 11.4 Implement event filtering (show only motion events) and time range zoom on timeline; verify filtering is responsive
- [x] 11.5 Implement archive video player with playback controls and seek; verify video plays from selected timestamp
- [x] 11.6 Implement Settings page with tabs for Cameras, Motion, Storage, Scheduling; verify forms submit to API and changes persist

## 12. Configuration Web Interface

- [x] 12.1 Create form for adding/editing cameras (name, RTSP URL, credentials); verify form validation and test connection button works
- [x] 12.2 Create motion detection settings form (sensitivity slider, skip frames, motion timeout); verify changes take effect in real-time
- [x] 12.3 Create S3 configuration form (bucket, region, path prefix, credentials); verify test connection button validates credentials
- [x] 12.4 Create screenshot scheduling interface (intervals per camera); verify schedule is saved and screenshots are captured
- [x] 12.5 Implement system status display (cameras online/offline, S3 connectivity, storage usage); verify metrics update in real-time

## 13. HTTP Server and Static File Serving

- [x] 13.1 Implement Go HTTP server listening on `localhost:8080` (or configurable port); verify server starts and responds to requests
- [x] 13.2 Embed React build artifacts into Go binary and serve from `/` route; verify web UI loads and assets are served with correct MIME types
- [x] 13.3 Implement CORS headers for HLS and API endpoints when needed; verify cross-origin requests are allowed
- [x] 13.4 Implement graceful shutdown (finish in-flight uploads, close database connection); verify shutdown takes <5 seconds

## 14. Error Handling and Logging

- [x] 14.1 Implement structured logging (DEBUG, INFO, WARN, ERROR levels) throughout application; verify logs are readable and include timestamps
- [x] 14.2 Implement error handling for all external service calls (RTSP, S3, authentication); verify failures are logged and UI displays user-friendly messages
- [x] 14.3 Implement `/api/diagnostics/logs` endpoint that returns recent log lines; verify logs are accessible from web UI

## 15. Integration Testing and Deployment

- [x] 15.1 Create integration test with mock RTSP server, verify full pipeline: capture → motion detection → encoding → S3 upload → archive query
- [x] 15.2 Test with real RTSP stream (if available) or simulator; verify system handles realistic load (1-2 cameras, continuous recording)
- [x] 15.3 Create deployment guide and systemd service; verify binary starts as service and survives reboot
- [x] 15.4 Document configuration example (sample config.yaml with OIDC and S3 settings); verify documentation is complete and accurate
- [x] 15.5 Create operational runbook: how to add camera, troubleshoot offline status, check logs, monitor storage; verify guide is practical

## 16. Performance and Resource Optimization

- [x] 16.1 Profile CPU and memory usage with 2 cameras running continuously; verify CPU stays <70%, memory <1.5GB
- [x] 16.2 Test motion detection frame skipping under load; verify skipping every 10 frames reduces CPU by 50%+
- [x] 16.3 Monitor S3 upload times and retry behavior; verify chunk upload completes within 60 seconds for typical network
- [x] 16.4 Stress test HLS segment serving with concurrent viewers; verify system handles 5+ simultaneous HLS streams without degradation

## Workflow follow-up

- Archive the change after all tasks are completed and integration testing passes.
- Verify the archived change creates main specs in `openspec/specs/` and move tasks to main repository.
