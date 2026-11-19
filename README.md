# RTSP Surveillance System

[![Backend CI](https://github.com/artemkoru/rtsprec/actions/workflows/backend.yml/badge.svg)](https://github.com/artemkoru/rtsprec/actions/workflows/backend.yml)
[![Frontend CI](https://github.com/artemkoru/rtsprec/actions/workflows/frontend.yml/badge.svg)](https://github.com/artemkoru/rtsprec/actions/workflows/frontend.yml)

A comprehensive RTSP video surveillance system with clean architecture, motion detection, video archiving, and HLS streaming capabilities.

## Overview

RTSP Surveillance is a full-stack surveillance platform featuring:
- **Multi-camera RTSP capture** from IP cameras
- **Motion detection** with configurable sensitivity
- **Video archiving** to AWS S3 and local SQLite database
- **HLS streaming** for live video playback
- **Web-based configuration** and management UI
- **User authentication** (OIDC or basic auth)
- **Clean architecture** with independent, testable layers

## Quick Start

### Backend (Go)

**Requirements:**
- Go 1.26.0 or later
- SQLite3
- AWS account (for S3 storage)

**Setup:**

```bash
# Clone the repository
git clone https://github.com/artemkoru/rtsprec.git
cd rtsprec

# Configure via environment variables (see the Configuration section):
# - S3 credentials and bucket
# - Authentication settings
# - Database and encryption key paths

# Build the backend
go build -o rtsprec ./cmd/rtsprec

# Run the server
./rtsprec
```

The backend will start an HTTP server on the configured port (default: 8080).

### Frontend (React/Vite)

**Requirements:**
- Node.js 20 or later
- npm or yarn

**Setup:**

```bash
cd web

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

The frontend will be available at `http://localhost:5173` in development.

## Features

### Core Capabilities

- **RTSP Camera Integration**
  - Multi-camera support
  - Automatic stream capture and encoding
  - Configurable video bitrate and resolution

- **Motion Detection**
  - Real-time motion event detection
  - Configurable sensitivity per camera
  - Motion event logging and query

- **Video Archive**
  - Persistent storage to SQLite + S3
  - Timeline view of recordings by date
  - Query videos by camera and date range
  - HLS streaming for playback

- **Web Interface**
  - Camera management and status
  - Live video preview (HLS)
  - Motion event history
  - Video archive browser
  - System configuration panel

- **Authentication**
  - OIDC integration (Google, Yandex)
  - Basic auth fallback
  - Session management

## Architecture

For a detailed explanation of the system architecture and design decisions, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

**High-level structure:**

```
Domain Layer (entities, value objects)
     ↓
Repository Layer (data access abstraction)
     ├─ SQLite (metadata, state)
     └─ S3 (video storage)
     ↓
Service Layer (business logic orchestration)
     ↓
HTTP Handlers (request/response mapping)
     ↓
App Package (dependency injection, wiring)
     ↓
cmd/rtsprec (main entry point)
```

## Configuration

The system is configured only through environment variables. The full list with defaults is in [docs/deployment.md](docs/deployment.md#environment-variables). Example:

```bash
export HTTP_PORT=8080
export DB_PATH=/opt/rtsprec/data/rtsprec.db
export S3_BUCKET=my-surveillance-bucket
export S3_REGION=us-east-1
export S3_ACCESS_KEY_ID=YOUR_KEY
export S3_SECRET_ACCESS_KEY=YOUR_SECRET
export AUTH_METHOD=oidc            # or basic
export AUTH_OIDC_PROVIDER=google
export AUTH_OIDC_CLIENT_ID=...
export AUTH_OIDC_SECRET=...
./rtsprec
```

## Testing

### Backend Tests

```bash
# Run all backend tests (refactored packages)
go test ./internal/domain ./internal/repository ./internal/service ./internal/app

# Run with verbose output
go test -v ./internal/...

# Run specific package tests
go test -v ./internal/service
```

For more details, see [docs/testing.md](docs/testing.md).

### Frontend Tests

```bash
cd web

# Run linting
npm run lint

# (Additional frontend tests can be added in future phases)
```

### Code Quality

```bash
# Format code
gofmt -w ./internal/

# Run Go vet
go vet ./internal/...
```

## Development

### Project Structure

```
rtsprec/
├── cmd/
│   └── rtsprec/           # Main application entry point
├── internal/
│   ├── domain/            # Business entities (clean architecture)
│   ├── repository/        # Data access abstraction & implementations
│   │   ├── sqlite/        # SQLite implementations
│   │   └── s3/            # AWS S3 implementations
│   ├── service/           # Business logic orchestration
│   ├── handler/           # HTTP request handlers
│   ├── app/               # Application bootstrap & wiring
│   ├── capture/           # RTSP stream capture (legacy)
│   ├── archive/           # Archive management (legacy)
│   ├── auth/              # Authentication (legacy)
│   └── ...                # Other legacy modules
├── web/                   # React frontend application
├── docs/
│   ├── ARCHITECTURE.md    # Detailed architecture guide
│   ├── deployment.md      # Deployment instructions
│   ├── KNOWN_ISSUES.md    # Known limitations and workarounds
│   └── testing.md         # Testing strategy
```

### Clean Architecture Layers

The refactored core (domain, repository, service) follows clean architecture principles:

- **Domain Layer** (`internal/domain/`): Pure business entities with no dependencies
- **Repository Layer** (`internal/repository/`): Data access abstraction with multiple implementations
- **Service Layer** (`internal/service/`): Business logic orchestration, decoupled from storage details
- **App Layer** (`internal/app/`): Dependency injection and application wiring

For detailed design decisions, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Deployment

For deployment instructions including Docker, systemd, and cloud platforms, see [docs/deployment.md](docs/deployment.md).

**Quick start:**

```bash
# Build release binary
go build -o rtsprec ./cmd/rtsprec

# Configure environment variables (see docs/deployment.md)

# Run as systemd service (see deployment.md)
sudo systemctl start rtsprec
```

## Known Issues & Troubleshooting

### gortsplib v4 Incompatibility

The project has a pre-existing incompatibility with gortsplib v4.16.3 in the RTSP capture module (`internal/capture/`). This does not affect the refactored core layers (domain, service, repository, handler).

**Impact:**
- Full binary build may fail when including capture package
- Refactored packages compile and test cleanly
- CI builds test only refactored packages
- Running without capture functionality is possible

**Workaround:**
See [docs/KNOWN_ISSUES.md](docs/KNOWN_ISSUES.md) for detailed information and mitigation strategies.

For other issues and troubleshooting, see [docs/KNOWN_ISSUES.md](docs/KNOWN_ISSUES.md).

## CI/CD

The project uses GitHub Actions for continuous integration:

- **Backend CI** (`.github/workflows/backend.yml`)
  - Triggered on: push to main, all PRs
  - Path-filtered to: `cmd/`, `internal/`, `go.mod`, `go.sum`
  - Jobs: lint (vet, gofmt), test, build

- **Frontend CI** (`.github/workflows/frontend.yml`)
  - Triggered on: push to main, all PRs
  - Path-filtered to: `web/`
  - Jobs: install, lint (oxlint), build

Workflows are independent and only trigger when their respective paths change.

## Contributing

Contributions are welcome! Please ensure:
- Backend changes pass: `go vet`, `gofmt`, `go test`
- Frontend changes pass: `npm run lint`, `npm run build`
- Commit messages are descriptive
- README and documentation stay current with changes

## License

[Add your license here]

## Support

For issues, questions, or discussions, please open a GitHub issue.

---

**Status:** Actively maintained | **Last Updated:** November 2026
