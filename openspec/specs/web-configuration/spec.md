# web-configuration Specification

## Purpose
Provide a web-based settings interface for configuring cameras, motion detection parameters, S3 storage, recording options, and scheduled tasks without requiring manual file editing.

## Requirements

### Requirement: Configure camera sources
The system SHALL allow adding, editing, and removing RTSP camera definitions (name, URL, username, password) through web UI.

#### Scenario: Add camera
- **WHEN** admin enters camera details in web form (name, RTSP URL, credentials)
- **THEN** system saves configuration and begins capturing from that camera

#### Scenario: Validate RTSP connection
- **WHEN** camera is added
- **THEN** system tests connection and reports success/failure to UI

#### Scenario: Edit camera
- **WHEN** admin modifies camera settings
- **THEN** system updates configuration and may restart capture if URL/credentials change

#### Scenario: Remove camera
- **WHEN** admin deletes camera configuration
- **THEN** system stops capturing and closes RTSP connection

#### Scenario: Multiple cameras
- **WHEN** system has 1-2 cameras configured
- **THEN** UI displays list of cameras with status (online/offline) and edit/delete options

### Requirement: Configure motion detection sensitivity
The system SHALL provide slider or input control to set motion detection threshold (0-100 scale) per camera.

#### Scenario: Sensitivity adjustment
- **WHEN** admin sets sensitivity to 50 (medium) for camera-1
- **THEN** system applies new threshold to motion detection algorithm

#### Scenario: Real-time effect
- **WHEN** sensitivity is changed
- **THEN** new threshold takes effect within seconds for next frame analysis

#### Scenario: Sensitivity defaults
- **WHEN** new camera is added without explicit sensitivity setting
- **THEN** system defaults to 50 (medium)

### Requirement: Configure video recording duration
The system SHALL allow setting video chunk duration (in minutes) per camera, default 5 minutes.

#### Scenario: Duration configuration
- **WHEN** admin sets video duration to 10 minutes
- **THEN** system creates 10-minute video chunks instead of 5-minute

#### Scenario: Duration boundaries
- **WHEN** system supports duration range 1-60 minutes
- **THEN** UI enforces range validation

### Requirement: Configure S3 storage
The system SHALL provide form for AWS credentials (access key, secret key) and S3 bucket configuration (bucket name, region, path prefix).

#### Scenario: S3 credentials input
- **WHEN** admin enters AWS access key and secret key
- **THEN** system validates credentials by testing S3 connection

#### Scenario: Bucket validation
- **WHEN** admin specifies bucket name
- **THEN** system verifies bucket exists and is writable

#### Scenario: Path prefix
- **WHEN** admin sets S3 path prefix to `surveillance/camera-room-1/`
- **THEN** video uploads use this prefix: `surveillance/camera-room-1/YYYY/MM/DD/...`

#### Scenario: Credential security
- **WHEN** credentials are saved
- **THEN** system stores encrypted (not plaintext) in persistent config

### Requirement: Configure screenshot scheduling
The system SHALL allow setting screenshot capture intervals per camera (e.g., every 10/30/60 minutes).

#### Scenario: Schedule configuration
- **WHEN** admin sets screenshot interval to 30 minutes for camera-1
- **THEN** system captures and uploads JPEG snapshots every 30 minutes

#### Scenario: Schedule persistence
- **WHEN** system restarts
- **THEN** screenshot schedule resumes at next scheduled interval

#### Scenario: Enable/disable screenshots
- **WHEN** admin toggles screenshot feature on/off
- **THEN** system starts/stops screenshot capture

### Requirement: Configure motion-triggered recording continuation
The system SHALL allow setting how long recording continues after motion ends (e.g., 30 seconds, 5 minutes).

#### Scenario: Motion timeout configuration
- **WHEN** admin sets motion timeout to 2 minutes
- **THEN** after motion stops, system continues recording for 2 minutes before stopping

#### Scenario: Continuous recording mode
- **WHEN** motion timeout is set to 0
- **THEN** system records continuously regardless of motion (continuous backup mode)

### Requirement: Provide configuration persistence
The system SHALL save all configuration changes to persistent storage and restore on system restart.

#### Scenario: Config survival
- **WHEN** admin configures cameras and S3 settings, then system restarts
- **THEN** all settings are restored and system resumes operation

#### Scenario: Config export
- **WHEN** admin requests to export settings
- **THEN** system provides downloadable JSON file with all configuration (excludes sensitive credentials)

#### Scenario: Config import (optional)
- **WHEN** admin uploads previously exported config file
- **THEN** system restores settings (may require re-entering sensitive credentials)

### Requirement: Display system status
The system SHALL show real-time status of capture services, cameras, and S3 connectivity in configuration UI.

#### Scenario: Camera status
- **WHEN** web UI loads
- **THEN** displays status of each camera: online/offline, frames captured, motion events

#### Scenario: S3 status
- **WHEN** web UI loads
- **THEN** displays S3 connection status and last successful upload

#### Scenario: Disk/resource usage
- **WHEN** system is running
- **THEN** UI shows CPU usage, memory, disk space for reference

### Requirement: Logs and diagnostics
The system SHALL provide access to recent system logs and diagnostics for troubleshooting (optional detailed logging level selector).

#### Scenario: View recent logs
- **WHEN** admin opens Diagnostics section
- **THEN** system displays last 100 log lines with timestamps and severity (DEBUG/INFO/ERROR)

#### Scenario: Log level adjustment
- **WHEN** admin sets log level to DEBUG
- **THEN** system increases logging verbosity for troubleshooting
