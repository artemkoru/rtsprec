-- SQLite schema for RTSP surveillance system

-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- Cameras table
CREATE TABLE IF NOT EXISTS cameras (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    rtsp_url TEXT NOT NULL,
    username TEXT,
    password TEXT,
    status TEXT DEFAULT 'offline',
    motion_sensitivity INTEGER DEFAULT 50,
    motion_timeout INTEGER DEFAULT 120,
    video_duration_minutes INTEGER DEFAULT 5,
    screenshot_interval_minutes INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Motion events table
CREATE TABLE IF NOT EXISTS motion_events (
    id TEXT PRIMARY KEY,
    camera_id TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    sensitivity_threshold INTEGER,
    video_file_path TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
);

-- Video metadata table
CREATE TABLE IF NOT EXISTS video_metadata (
    id TEXT PRIMARY KEY,
    camera_id TEXT NOT NULL,
    s3_path TEXT NOT NULL UNIQUE,
    file_name TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration_seconds INTEGER,
    file_size_bytes INTEGER,
    codec TEXT,
    resolution TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
);

-- Configuration table (for global settings)
CREATE TABLE IF NOT EXISTS configuration (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_motion_events_camera_id ON motion_events(camera_id);
CREATE INDEX IF NOT EXISTS idx_motion_events_start_time ON motion_events(start_time);
CREATE INDEX IF NOT EXISTS idx_video_metadata_camera_id ON video_metadata(camera_id);
CREATE INDEX IF NOT EXISTS idx_video_metadata_start_time ON video_metadata(start_time);
CREATE INDEX IF NOT EXISTS idx_video_metadata_s3_path ON video_metadata(s3_path);
