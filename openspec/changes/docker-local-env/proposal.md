# Proposal

## Why

Running the backend locally needs a Go toolchain, a reachable S3 bucket and a set of environment variables that are easy to get wrong. The integration tests already show that the backend image and a SeaweedFS S3 server work together with no account. A local stack in Docker gives every developer the same backend and S3 setup with one command, and it runs the same image that CI and the tests use.

The README also still describes a gortsplib v4 incompatibility that has been resolved. A new reader would expect the build to fail, so the section is misleading.

## What Changes

- New `compose.yaml` at the repository root with two services: `backend` (built from the root `Dockerfile`) and `s3` (`chrislusf/seaweedfs`, S3 gateway enabled, data kept in a named volume).
- The backend reads its configuration from environment variables in `compose.yaml`. S3 is reached through the compose network. The database, encryption key and HLS output live in a named volume, so restarts keep data.
- Makefile targets:
  - `make docker-build` builds the backend image;
  - `make docker-up` starts the stack and waits until the backend answers;
  - `make docker-down` stops it; `make docker-logs` follows the logs.
- The web UI stays on the host (`npm run dev` in `web/`), as in the Local Run section of the README. The dev proxy reaches the backend on `localhost:8080`, which compose publishes.
- README: the "Local Run" section gets a Docker variant. The stale "Known Issues" section about gortsplib v4 is replaced with a short list of current limitations and a link to `docs/KNOWN_ISSUES.md`.

Out of scope: running the web UI in a container (needs the dev proxy from PR #7 on `main`), embedding the frontend into the backend binary, and any backend behavior change.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

<!-- None: backend behavior does not change. -->

## Impact

- New: `compose.yaml`. Changed: `Makefile` (new targets), `README.md` (Local Run and Known Issues sections).
- The bucket name, region and credentials for the local S3 are fixed in `compose.yaml` and documented as local-only values; they are not production secrets.
- Requires Docker with the compose plugin. The stack pulls `chrislusf/seaweedfs` on first start.
