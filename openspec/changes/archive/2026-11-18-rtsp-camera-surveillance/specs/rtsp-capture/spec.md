# Spec Delta: RTSP Capture

## Purpose

Continuously capture video from IP cameras via RTSP protocol, detect motion with configurable sensitivity, encode video into 5-minute segments, and upload to S3 with event metadata.

## ADDED Requirements

### Requirement: Connect to RTSP camera stream
The system SHALL establish and maintain a persistent connection to an RTSP camera stream specified by URL and authentication credentials (username/password).

#### Scenario: Successful connection
- **WHEN** valid RTSP URL and credentials are provided
- **THEN** system connects and begins receiving video frames

#### Scenario: Connection failure
- **WHEN** RTSP URL is unreachable or invalid
- **THEN** system logs error and retries connection periodically

#### Scenario: Authentication failure
- **WHEN** provided credentials are incorrect
- **THEN** system logs auth error and does not attempt further reconnections until credentials change

### Requirement: Detect motion in video frames
The system SHALL analyze incoming video frames to detect motion by comparing consecutive frames at configurable sensitivity threshold.

#### Scenario: Motion detected
- **WHEN** frame difference exceeds sensitivity threshold (0-100 scale)
- **THEN** system records motion event with timestamp

#### Scenario: Skip frames for efficiency
- **WHEN** motion detection is enabled
- **THEN** system analyzes every Nth frame (10-30 configurable) to reduce CPU load

#### Scenario: No motion
- **WHEN** consecutive frames show no significant changes
- **THEN** system continues passive monitoring without event logging

### Requirement: Encode video into time-based chunks
The system SHALL encode video stream into H.264 or H.265 video files, grouped into 5-minute segments (duration configurable), and package each into an MP4 container.

#### Scenario: Video chunk generation
- **WHEN** 5 minutes of video has been captured
- **THEN** system finishes encoding current chunk and creates a new one with incrementing sequence

#### Scenario: Codec selection
- **WHEN** system starts capturing from camera stream
- **THEN** system matches source codec (H.264 or H.265) or re-encodes if necessary

#### Scenario: Chunk duration boundary
- **WHEN** a motion event occurs near chunk boundary
- **THEN** event timestamp is logged accurately regardless of chunk split

### Requirement: Upload video chunks to S3
The system SHALL upload completed video chunks to S3 with configurable bucket, path prefix, and retry policy on network failure.

#### Scenario: Successful upload
- **WHEN** video chunk is complete
- **THEN** system uploads to S3 with naming scheme: `<prefix>/<camera-id>/YYYY/MM/DD/HH-MM-SS-<sequence>.mp4`

#### Scenario: S3 credentials provided
- **WHEN** system is configured with AWS credentials and bucket
- **THEN** system validates credentials on startup and logs warning if invalid

#### Scenario: Upload failure retry
- **WHEN** S3 upload fails (network error, timeout)
- **THEN** system retries up to 3 times with exponential backoff before logging failure

#### Scenario: Duplicate protection
- **WHEN** same chunk is uploaded twice
- **THEN** system prevents duplicate uploads by checking S3 object existence

### Requirement: Log motion events with metadata
The system SHALL record motion detection events with timestamp, camera ID, duration, and sensitivity threshold to persistent storage.

#### Scenario: Event stored
- **WHEN** motion is detected
- **THEN** system writes event record with: timestamp (ISO-8601), camera ID, motion start/end times, sensitivity level

#### Scenario: Event linkage
- **WHEN** motion event falls within a video chunk
- **THEN** event record includes reference to S3 video file name

#### Scenario: Continuous storage
- **WHEN** motion events are logged
- **THEN** events persist across system restarts

### Requirement: Capture scheduled screenshots
The system SHALL capture and save single-frame images (JPEG) to S3 at configurable time intervals per camera.

#### Scenario: Screenshot scheduled
- **WHEN** time matches configured schedule (e.g., every 10 minutes)
- **THEN** system captures current frame and uploads to S3 as JPEG: `<prefix>/<camera-id>/screenshots/YYYY/MM/DD/HH-MM-SS.jpg`

#### Scenario: Schedule configuration
- **WHEN** screenshot interval is set (e.g., 10, 30, 60 minutes)
- **THEN** system takes snapshots at specified intervals independently of motion detection

#### Scenario: Screenshot on startup
- **WHEN** system starts or reconnects to camera
- **THEN** system captures initial screenshot after connection is stable

### Requirement: Handle disconnection and reconnection
The system SHALL gracefully handle camera stream disconnections, complete any in-progress video chunk, and automatically reconnect with exponential backoff.

#### Scenario: Unexpected disconnect
- **WHEN** RTSP stream ends unexpectedly
- **THEN** system finishes current video chunk (or saves partial if duration exceeds timeout) and closes cleanly

#### Scenario: Auto-reconnection
- **WHEN** connection is lost
- **THEN** system waits 2 seconds, then 4, 8, 16... up to 5 minutes between retry attempts

#### Scenario: Connection recovery
- **WHEN** connection is re-established
- **THEN** system resumes capturing and updates status (online/offline)
