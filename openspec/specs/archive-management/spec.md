# archive-management Specification

## Purpose
Manage video archive indexing, event metadata storage, and timeline generation for efficient retrieval and visualization of historical recordings and motion events.

## Requirements

### Requirement: Index video files in archive
The system SHALL maintain a database index of all video files stored in S3, including filename, camera ID, start/end timestamps, duration, and storage location.

#### Scenario: File indexed on upload
- **WHEN** video chunk is successfully uploaded to S3
- **THEN** system creates database record with S3 path, timestamps, and camera ID

#### Scenario: Index persistence
- **WHEN** system restarts
- **THEN** video index remains in database (SQLite or persistent storage)

#### Scenario: Index lookup
- **WHEN** web UI queries for videos from camera on specific date
- **THEN** system returns list of available video files with timestamps

### Requirement: Store motion event metadata
The system SHALL persist motion detection events with detailed metadata for timeline generation and event correlation.

#### Scenario: Event record creation
- **WHEN** motion is detected
- **THEN** system stores: camera ID, event start time, event end time, motion duration, motion sensitivity threshold, associated S3 video file

#### Scenario: Event uniqueness
- **WHEN** motion event is logged
- **THEN** event record includes unique ID for later reference

#### Scenario: Event retrieval
- **WHEN** web UI requests events for camera on specific date range
- **THEN** system returns sorted list with timestamps, durations, and video file references

### Requirement: Generate timeline for archive view
The system SHALL create a timeline representation showing video segments and motion events for a camera over a time period, optimized for UI rendering.

#### Scenario: Timeline query
- **WHEN** web UI requests timeline for camera on 2026-11-17
- **THEN** system returns array of segments/events with: start time, end time, type (video|event), metadata (file path, duration, severity)

#### Scenario: Timeline granularity
- **WHEN** user views 24-hour period
- **THEN** timeline consolidates adjacent segments into logical groups, showing ~50 timeline items

#### Scenario: Timeline event markers
- **WHEN** timeline is generated
- **THEN** motion events appear as distinct markers linked to video files

### Requirement: Query video availability
The system SHALL provide efficient queries to find video files and events within specified time ranges per camera.

#### Scenario: Time range query
- **WHEN** web UI requests videos from camera-1 between 10:00 and 15:00 on 2026-11-17
- **THEN** system returns all 5-minute video chunks that overlap with time range

#### Scenario: Date range spanning multiple chunks
- **WHEN** query spans entire day with many 5-minute video chunks
- **THEN** system efficiently returns all matching chunks without full table scan

#### Scenario: Event filtering
- **WHEN** web UI requests only motion events (not all video)
- **THEN** system returns filtered event list with linked video file references

### Requirement: Handle archive growth
The system SHALL manage database size as archive accumulates video metadata over time, without requiring manual cleanup.

#### Scenario: Long-term storage
- **WHEN** system runs for 30 days with daily captures
- **THEN** database contains ~8,640 video records (288 per day × 30) without performance degradation

#### Scenario: Index maintenance
- **WHEN** video index grows large
- **THEN** system periodically optimizes database (VACUUM, index rebuilds) or uses efficient schema

### Requirement: Correlate video and motion events
The system SHALL allow web UI to display motion events alongside corresponding video file, enabling user to watch detected activity.

#### Scenario: Event to video mapping
- **WHEN** user clicks motion event on timeline
- **THEN** system returns associated S3 video file name and timestamp within that file

#### Scenario: Video segment alignment
- **WHEN** motion event spans multiple video chunks
- **THEN** system returns all chunks and time offset where motion started within the file

### Requirement: Retrieve video segments for playback
The system SHALL generate authenticated S3 URLs or proxy video content for web player, allowing playback of archived video.

#### Scenario: Presigned URL generation
- **WHEN** web UI requests video file `video-001.mp4` for authenticated user
- **THEN** system generates time-limited S3 presigned URL valid for 1 hour and returns to client

#### Scenario: Playback authorization
- **WHEN** authenticated user accesses presigned URL
- **THEN** video downloads from S3 without additional authentication

#### Scenario: Access control
- **WHEN** unauthenticated user attempts to access presigned URL after expiration
- **THEN** request is rejected and URL is invalid
