# Design

## Context

The system must run on constrained hardware (2 CPU, 2GB RAM, 20GB HDD) while handling 1-2 continuous RTSP streams at 1080p@30fps. Motion detection must be real-time but CPU-efficient. Video must be archived to S3 and accessible via HLS. Configuration and viewing happens through a web UI on the same machine.

See proposal.md for motivation and business goals.

## Goals / Non-Goals

**Goals:**
- Handle 1-2 cameras simultaneously with minimal CPU/memory footprint
- Enable real-time viewing via HLS with <15 second latency
- Archive all video to S3 with event metadata for timeline-based retrieval
- Provide secure, user-friendly web UI for configuration and playback
- Support OIDC (Google, Yandex) and Basic Auth authentication

**Non-Goals:**
- Cloud-based multi-site management or federation
- Advanced ML-based motion detection (threshold-based is sufficient)
- Mobile app (web UI responsive only)
- Real-time video analytics or object detection
- Sub-second motion event triggering

## Decisions

### 1. Go backend + React frontend architecture

**Decision:** Monolithic Go service with embedded React web UI (single binary/deployment).

**Rationale:**
- Go is lightweight, compiles to single binary, excellent for systems programming
- Minimal dependencies = smaller footprint on 2GB RAM
- React provides responsive, modern web UI
- Both run on same machine (HTTP only, no network overhead)

**Alternatives:**
- Python (Flask/FastAPI): More memory overhead, slower startup
- Microservices: Too complex for 1-2 cameras, operational burden

---

### 2. HLS for live streaming

**Decision:** Segment RTSP stream into 2-3 second HLS segments, serve M3U8 playlist and .ts files via HTTP.

**Rationale:**
- Avoids video re-encoding (use source codec H.264/H.265 directly)
- Standard, browser-native support (no special plugins)
- Handles network buffering naturally
- Stateless segment serving = simple horizontal scaling later
- ~10-15 second latency is acceptable for surveillance

**Alternatives:**
- WebRTC: Lower latency but requires STUN/TURN, more CPU for media negotiation
- RTSP proxy: Would require RTSP plugin in browser or custom player
- Raw MJPEG over HTTP: Too much bandwidth, doesn't scale

---

### 3. Threshold-based motion detection (skip frames)

**Decision:** Compare consecutive frames using pixel difference threshold, analyze every 10-30 frames (skipping frames 10-30X), store events in SQLite.

**Rationale:**
- Simple algorithm: sum absolute difference of pixels between frame N and frame N+10
- Skip frames to reduce CPU (1 analysis per 30 frames = ~1 FPS analysis instead of 30 FPS)
- SQLite for local metadata avoids adding external database
- Configurable threshold per camera (0-100 scale) allows tuning sensitivity

**Alternatives:**
- Optical flow: Too CPU-intensive for 2GB RAM
- ML models: Overkill, requires GPU or significant CPU overhead
- Full-frame analysis: Would saturate 2 cores quickly

---

### 4. Continuous 5-minute video chunks to S3

**Decision:** Encode video into 5-minute MP4 chunks continuously, upload to S3 immediately on completion. Record 24/7, don't skip non-motion periods.

**Rationale:**
- Continuous recording provides complete video history (not just motion events)
- 5-minute chunks balance S3 upload frequency with file size
- Motion detection marks "events" in metadata, doesn't skip recording
- S3 key structure: `<prefix>/<camera-id>/YYYY/MM/DD/HH-MM-SS-SEQ.mp4` enables date-based queries
- Avoid streaming directly to S3 (requires buffering, complex on failure)

**Alternatives:**
- Record only on motion: Misses context, motion detection false negatives expose gaps
- 1-minute chunks: More frequent uploads, higher S3 API calls and cost
- 30-minute chunks: Harder to trim/edit, inefficient archive retrieval

---

### 5. SQLite for metadata and configuration

**Decision:** Store event metadata (timestamps, motion records), camera config, and S3 credentials in SQLite (local file).

**Rationale:**
- No external database required, reduces operational complexity
- SQLite sufficient for 1-2 cameras with ~288 video records per day (40K+ events/day at most)
- Configuration persists across restarts
- Encrypt sensitive fields (S3 key) before storage

**Alternatives:**
- JSON config file: No structured query support for timeline generation
- PostgreSQL/MySQL: Overkill, adds network/service management overhead

---

### 6. Architecture: Three goroutines per camera

**Decision:** Per-camera: (1) RTSP reader, (2) motion detector, (3) video encoder/uploader.

```
RTSP Stream
    |
    v
[RTSP Reader Goroutine]
  - Decode H.264/H.265
  - Queue frames to channel
  - Handle reconnection
    |
    +---> [Motion Detector Goroutine]
    |     - Every 10-30 frames
    |     - Compute frame diff
    |     - Write events to SQLite
    |
    +---> [Video Encoder Goroutine]
          - Buffer video frames
          - Every 5 minutes: encode to MP4
          - Upload to S3
          - Update metadata in SQLite
```

**Rationale:**
- Separate goroutines allow independent scheduling (motion detection skips frames, encoding is time-based)
- Channels decouple components, handle backpressure
- One goroutine per camera allows independent stream handling

---

### 7. HLS segment generation concurrent with chunk encoding

**Decision:** Maintain sliding window of recent 2-3 second HLS segments in `/tmp/hls/<camera-id>/` while simultaneously encoding 5-minute chunks for S3.

**Rationale:**
- HLS segments are short-lived (expire after 20-30 seconds in playlist)
- Use same video frames for both HLS and S3 chunks (no duplicate capture)
- HLS segments stored on local disk (fast I/O)
- Completed 5-minute chunks uploaded to S3 asynchronously

---

### 8. Web UI: React SPA served from Go backend

**Decision:** Go serves static React build (`/static/`) + API endpoints (`/api/`).

**Rationale:**
- Single deployment unit (Go binary includes React assets)
- No separate frontend server required
- React client handles UI state, calls Go API for data

---

### 9. OIDC with Basic Auth fallback

**Decision:** Support configurable OIDC (Google, Yandex) with form-based Basic Auth as fallback.

**Rationale:**
- OIDC offloads authentication to trusted providers (no password storage)
- Basic Auth provides fallback if OIDC is unavailable or admin preference
- Both methods supported simultaneously (UI shows both options)

---

### 10. S3 presigned URLs for archive playback

**Decision:** Generate time-limited (1-hour) S3 presigned URLs for authenticated users, client downloads directly from S3.

**Rationale:**
- Reduces bandwidth from Go service (S3 handles delivery)
- Presigned URLs are time-limited and cryptographically signed
- Client-side download is more efficient than proxying through service

---

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Motion detection false positives** (e.g., shadows, light changes) | Configurable sensitivity threshold per camera allows tuning; users adjust empirically for each environment |
| **S3 upload failures** (network, quota) | Retry logic (3 attempts, exponential backoff); failures logged for manual investigation; chunks remain local until confirmed uploaded |
| **SQLite database corruption** (power loss, concurrent writes) | Use WAL (Write-Ahead Logging) mode in SQLite; backups optional (secondary concern for 1-2 cameras) |
| **HLS segment underflow** (player runs out of segments) | Maintain buffer of 3+ segments; acceptable gap = 10-15 second playback hiccup if network slowdown |
| **Disk space exhaustion** (video fills 20GB HDD fast) | Expected: ~1.5 GB/day @ 1080p 30fps MP4; 13-day retention on 20GB. No auto-cleanup in v1 (manual or S3-only operation). Consider adding retention policy in future |
| **CPU saturation on motion detection + encoding** | Analyze every 10-30 frames (skip frames); if CPU still high, reduce analysis granularity or disable screenshot scheduling on that camera |
| **Codec mismatch** (camera sends H.265, browsers don't support) | Detect codec on startup; offer option to re-encode to H.264 (increases CPU but ensures compatibility) or accept browser limitations |
| **Authentication bypass via misconfiguration** | Enforce HTTPS if exposed beyond localhost; mark session cookies Secure/HttpOnly; default to Basic Auth disabled until explicitly configured |

---

## Migration Plan (Deployment)

1. **Build**: Compile Go backend with embedded React frontend
2. **Install**: Copy binary to `/opt/rtsprec/` on target VM
3. **Configure**: Create `/opt/rtsprec/config.yaml` with S3 credentials, camera URLs
4. **Run**: Start as systemd service, listen on `http://localhost:8080`
5. **First login**: Use default Basic Auth credentials (or redirect to OIDC login if configured), change in UI
6. **Verify**: Check `/api/cameras/status` returns online status for configured cameras

**Rollback:** Stop service, previous version stays in `/opt/rtsprec.old/`, restore from backup if needed.

---

## Open Questions

1. **Codec re-encoding strategy**: Should system auto-detect H.265 and offer re-encode option in UI, or fail fast with error? (Answer: Offer option, detect at startup)
2. **Disk cleanup policy**: Should system auto-delete oldest video after space threshold, or leave to user/manual cleanup? (Defer: Implement manual cleanup in v2 if needed)
3. **Multi-site federation**: Any future plan to aggregate multiple machines' video? (Out of scope for v1; architecture assumes single machine)
4. **Audio**: Include audio from camera stream in video, or video-only? (Likely video-only to reduce storage, defer if not critical)
5. **Certificate management**: For HTTPS, use self-signed cert or Let's Encrypt? (Defer: v1 assumes localhost or internal network; HTTPS hardening in v2)
