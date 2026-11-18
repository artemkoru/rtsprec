# Proposal: RTSP Camera Surveillance System

## Why

Build a complete video surveillance system for 1-2 IP cameras that captures video continuously with motion detection, stores clips to S3, and provides live and archived viewing through a web interface. This enables reliable, scalable security monitoring with cloud storage and configurable sensitivity detection—all within minimal resource constraints (2 CPU, 2GB RAM).

## What Changes

- **New Go backend service** for RTSP video capture, motion detection, and S3 storage
- **React web frontend** for real-time viewing, archive access, and system configuration
- **Motion-triggered recording** with configurable sensitivity and recording duration
- **Scheduled screenshot capture** to S3 with configurable intervals
- **Live stream delivery** via HLS protocol (low CPU overhead)
- **Web-based configuration UI** for cameras, S3, motion sensitivity, recording parameters, and authentication
- **Event timeline** in archive view showing motion events and video segments
- **Authentication** via OIDC (Google, Yandex) or Basic Auth

## Capabilities

### New Capabilities

- `rtsp-capture`: Continuous video capture from RTSP cameras, frame-by-frame motion detection (configurable sensitivity), 5-minute video chunk encoding and S3 upload, event logging
- `live-streaming`: HLS live stream delivery for real-time camera viewing in web browser with minimal latency
- `archive-management`: Video clip storage and retrieval, event metadata indexing, timeline generation with motion markers
- `web-configuration`: Settings interface for adding cameras, S3 credentials, motion sensitivity, recording duration, screenshot schedule
- `authentication`: OIDC provider integration (Google, Yandex) or fallback to Basic Auth for web UI access
- `web-ui`: React frontend with live camera view, archive player with timeline, event markers, and configuration pages

### Modified Capabilities

None—this is a greenfield system.

## Impact

- **New dependencies**: Go RTSP library (gortsplib), H.264/H.265 video codec handling, AWS S3 SDK, HLS segmenter
- **New services**: Standalone Go service (Linux) + React frontend (served from same machine)
- **Storage**: SQLite for metadata/config (local), S3 for video (external)
- **Resource constraints**: 2 CPU, 2GB RAM—motion detection skips frames (every 10-30) to stay within budget; HLS live stream avoids full re-encoding
- **Network**: RTSP stream from cameras, WebSocket for config API, HTTP for HLS segments and web UI
