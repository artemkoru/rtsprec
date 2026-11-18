# Spec Delta: Web UI

## Purpose

Provide a React-based web interface for viewing live camera feeds, browsing archived video with event timeline, and accessing system configuration.

## ADDED Requirements

### Requirement: Display live camera stream
The system SHALL show real-time video from selected camera using HLS stream embedded in web player (HTML5 video or media.js player).

#### Scenario: Live view page
- **WHEN** authenticated user navigates to Live View
- **THEN** page displays list of configured cameras and selected camera's HLS stream in player

#### Scenario: Camera selection
- **WHEN** user clicks on camera name in list
- **THEN** player switches to that camera's HLS stream and displays live video

#### Scenario: Stream status
- **WHEN** camera is offline
- **THEN** player shows "Offline" message and attempts reconnection periodically

#### Scenario: Stream latency
- **WHEN** HLS stream is playing
- **THEN** latency is 10-15 seconds (acceptable for surveillance use case)

#### Scenario: Player controls
- **WHEN** user watches live stream
- **THEN** player displays play/pause, volume, fullscreen controls (browser standard)

### Requirement: Display archive with timeline
The system SHALL show calendar/date picker and timeline of video segments and motion events, allowing user to select and play archived video.

#### Scenario: Archive page load
- **WHEN** user navigates to Archive
- **THEN** page displays date picker defaulting to today, and timeline view below

#### Scenario: Date selection
- **WHEN** user selects date in calendar
- **THEN** page loads timeline for that date, showing all video segments and motion events

#### Scenario: Timeline display
- **WHEN** timeline is generated
- **THEN** page displays horizontal timeline from 00:00 to 23:59 with segments/events marked

#### Scenario: Segment markers
- **WHEN** timeline contains 5-minute video chunks
- **THEN** each chunk appears as colored bar on timeline with start/end time

#### Scenario: Motion event markers
- **WHEN** motion events exist for the date
- **THEN** events appear as distinct markers (different color/icon) on timeline aligned with video time

#### Scenario: Zoom/pan timeline
- **WHEN** user hovers over timeline
- **THEN** can zoom in/out to see finer time granularity (hourly, 30-min, 5-min views)

### Requirement: Play archived video
The system SHALL provide video player for archived segments, allowing playback with standard controls and timestamp navigation.

#### Scenario: Play video segment
- **WHEN** user clicks video segment on timeline
- **THEN** player loads and plays that video file from start

#### Scenario: Seek within video
- **WHEN** user drags playback slider
- **THEN** video seeks to specified time within segment

#### Scenario: Continuous playback across segments
- **WHEN** video reaches end of segment
- **THEN** player can automatically load and play next segment (optional seamless transitions)

#### Scenario: Download video
- **WHEN** user clicks download icon
- **THEN** system generates S3 presigned URL and downloads MP4 file

### Requirement: Filter and search timeline
The system SHALL allow filtering timeline to show only motion events (excluding non-event video segments) for quick event review.

#### Scenario: Event-only view
- **WHEN** user toggles "Show Only Motion Events" filter
- **THEN** timeline displays only motion markers, hiding non-event segments

#### Scenario: Time range filter
- **WHEN** user specifies time range (e.g., 10:00-15:00)
- **THEN** timeline shows only segments/events within that range

#### Scenario: Sensitivity filter (optional)
- **WHEN** multiple events have varying sensitivity
- **THEN** user can filter by sensitivity threshold to find high-confidence motion

### Requirement: Configuration interface
The system SHALL provide settings page for adding cameras, configuring motion detection, S3 credentials, and scheduling.

#### Scenario: Settings page access
- **WHEN** authenticated admin clicks Settings
- **THEN** page displays tabs for Cameras, Motion Settings, Storage, Scheduling, Logs

#### Scenario: Camera management section
- **WHEN** admin opens Cameras tab
- **THEN** page shows list of cameras with Add, Edit, Delete options (see web-configuration spec)

#### Scenario: Motion settings section
- **WHEN** admin opens Motion Settings
- **THEN** page displays sensitivity sliders and timeout configs per camera

#### Scenario: Storage configuration
- **WHEN** admin opens Storage tab
- **THEN** page displays S3 form with bucket, credentials, and test connection button

#### Scenario: Schedule configuration
- **WHEN** admin opens Scheduling tab
- **THEN** page displays screenshot interval settings per camera

### Requirement: Display system status
The system SHALL show current health status (cameras online/offline, S3 connectivity, storage usage) on dashboard or status panel.

#### Scenario: Status dashboard
- **WHEN** user views Live View or Archive page
- **THEN** header/sidebar shows camera status indicators (green=online, red=offline)

#### Scenario: S3 connectivity
- **WHEN** dashboard loads
- **THEN** displays S3 connection status (last upload time, success/error)

#### Scenario: Storage metrics
- **WHEN** status panel is visible
- **THEN** shows storage used vs. available space (HDD percentage)

### Requirement: Responsive design
The system SHALL provide usable interface on desktop browsers (Chrome, Firefox, Safari) and mobile devices (optional).

#### Scenario: Desktop view
- **WHEN** viewed on desktop (1920×1080 or larger)
- **THEN** timeline has sufficient height, player is prominent, controls are accessible

#### Scenario: Mobile view (optional)
- **WHEN** viewed on mobile device
- **THEN** layout adapts: player takes full width, timeline below, settings in hamburger menu

### Requirement: Real-time UI updates
The system SHALL update camera status, event count, and storage metrics without full page reload using WebSocket or polling.

#### Scenario: Status polling
- **WHEN** web UI is open
- **THEN** every 5-10 seconds, page fetches updated camera statuses via `/api/cameras/status`

#### Scenario: New events appear
- **WHEN** motion event occurs while archive page is open
- **THEN** event appears on timeline in real-time (or on next refresh)

#### Scenario: Minimal network overhead
- **WHEN** polling occurs
- **THEN** each request is small JSON payload (~1KB) to avoid excessive bandwidth

### Requirement: Error handling and feedback
The system SHALL display user-friendly error messages for failures (connection errors, auth failures, missing video).

#### Scenario: Stream connection failure
- **WHEN** HLS stream fails to load
- **THEN** player displays "Failed to load stream. Retrying..." message

#### Scenario: API error
- **WHEN** API call fails (e.g., 500 Server Error)
- **THEN** UI displays notification: "Failed to fetch cameras. Please try again."

#### Scenario: Validation feedback
- **WHEN** user enters invalid S3 credentials
- **THEN** form shows error: "Failed to connect to S3. Check credentials."
