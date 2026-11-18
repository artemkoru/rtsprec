# Spec Delta: Live Streaming

## Purpose

Deliver real-time video from IP cameras to web browsers using HLS (HTTP Live Streaming) protocol with minimal latency and low CPU overhead.

## ADDED Requirements

### Requirement: Generate HLS segment stream
The system SHALL segment incoming video into short HLS segments (2-3 seconds each) and generate an M3U8 playlist pointing to current and recent segments.

#### Scenario: HLS playlist generation
- **WHEN** RTSP stream is active
- **THEN** system generates M3U8 playlist with 3-5 available segments at `/hls/<camera-id>/playlist.m3u8`

#### Scenario: Segment availability
- **WHEN** HLS client requests playlist
- **THEN** system returns segments in sliding window, oldest segments dropped after 20-30 seconds

#### Scenario: Segment naming
- **WHEN** new segment is created
- **THEN** segment is named with sequence number: `segment-00000.ts`, `segment-00001.ts`, etc.

### Requirement: Serve HLS segments over HTTP
The system SHALL expose HLS segments and playlist via HTTP endpoints accessible by web browser clients, with appropriate CORS headers for cross-origin requests.

#### Scenario: HTTP endpoint
- **WHEN** web client requests `/hls/<camera-id>/playlist.m3u8`
- **THEN** system returns valid M3U8 playlist with absolute URLs to segment files

#### Scenario: Segment download
- **WHEN** client downloads `/hls/<camera-id>/segment-00001.ts`
- **THEN** system returns MPEG-TS video segment with 200 OK status

#### Scenario: CORS headers
- **WHEN** web client makes cross-origin HLS request
- **THEN** system returns `Access-Control-Allow-Origin: *` header

#### Scenario: 404 handling
- **WHEN** client requests non-existent segment
- **THEN** system returns 404 Not Found (segment has rolled off)

### Requirement: Support multiple concurrent viewers
The system SHALL allow multiple web clients to simultaneously view live streams from the same camera without degrading video quality or increasing encoding load.

#### Scenario: Concurrent viewers
- **WHEN** 3 clients request same camera HLS stream
- **THEN** all receive same segments without additional encoding

#### Scenario: Viewer isolation
- **WHEN** one viewer closes connection
- **THEN** other viewers continue receiving uninterrupted stream

### Requirement: Handle stream format and codec compatibility
The system SHALL maintain video encoding from source camera (H.264 or H.265) or use minimal re-encoding, ensuring compatibility with common browsers.

#### Scenario: H.264 passthrough
- **WHEN** RTSP source is H.264
- **THEN** system segments H.264 stream directly without re-encoding

#### Scenario: H.265 handling
- **WHEN** RTSP source is H.265
- **THEN** system either segments H.265 (if browser supports) or re-encodes to H.264 (CPU overhead acceptable for 1-2 cameras)

### Requirement: Provide stream status and health monitoring
The system SHALL report stream status (active, offline, error) and allow clients to detect and handle disconnections gracefully.

#### Scenario: Stream status endpoint
- **WHEN** web client requests `/api/cameras/<camera-id>/status`
- **THEN** system returns JSON: `{status: "online"|"offline", last_updated: timestamp}`

#### Scenario: Offline detection
- **WHEN** RTSP stream disconnects
- **THEN** system marks camera offline in status and stops updating HLS segments

#### Scenario: Reconnection
- **WHEN** RTSP stream reconnects
- **THEN** system resumes HLS segment generation and updates status to online

### Requirement: Secure stream access
The system SHALL require authentication before allowing HLS stream access, honoring configured authentication method (OIDC or Basic Auth).

#### Scenario: Unauthorized access
- **WHEN** unauthenticated client requests HLS stream
- **THEN** system returns 403 Forbidden

#### Scenario: Authenticated access
- **WHEN** authenticated client requests HLS stream
- **THEN** system serves HLS playlist and segments normally

#### Scenario: Session expiration
- **WHEN** client session expires during streaming
- **THEN** system returns 401 Unauthorized on next segment request
