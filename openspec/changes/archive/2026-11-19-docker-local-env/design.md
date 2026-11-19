# Design

## Context

The backend binary reads only environment variables (see `docs/deployment.md`). At startup it checks the S3 bucket with `HeadBucket` and exits if the bucket is unreachable. The root `Dockerfile` builds `./cmd/rtsprec` into a non-root image that expects `/home/rtsprec` to be writable. The integration tests already run this image against SeaweedFS on a shared network (`integration/helpers_test.go`). The web UI is a Vite dev server that proxies `/api` and `/hls` to `BACKEND_URL` (default `http://localhost:8080`), and it is not yet part of `main` (PR #7). See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- One command starts the backend and an S3-compatible server locally.
- The local backend uses the same image as the integration tests and CI build.
- Data survives `docker compose down` and `up`, unless the volumes are removed.
- The UI on the host reaches the containerized backend without changes to the Vite config.

**Non-Goals:**
- Running the web UI in a container.
- Production deployment, TLS, or secrets management.
- Changing backend code or configuration names.

## Decisions

- **`compose.yaml` at the repository root.** Compose is the standard way to describe a multi-container local stack, and the Docker CLI on the developer machine already includes it. Alternative: `docker run` commands in the Makefile. Rejected: they duplicate the environment and make restarts fragile.
- **SeaweedFS as the local S3.** It is public on Docker Hub, works with the AWS SDK as shown by the integration tests, and needs no account. Its S3 gateway runs with `server -s3 -s3.domainName=s3` so virtual-host bucket addressing works on the compose network. Alternative: MinIO, rejected because its images are not pullable without an account.
- **Bucket creation by a one-shot `s3-init` service** using the public `amazon/aws-cli` image: `aws --endpoint-url http://s3:8333 s3 mb s3://<bucket>`, ignoring "already exists". The backend does not create buckets, and this keeps the step inside compose. Alternative: document a manual command. Rejected: the stack would not start from one command.
- **Published ports.** `backend` publishes `8080` to the host so the Vite proxy works without changes. `s3` publishes its S3 port only for debugging; the backend uses the compose network name.
- **Named volumes** for `backend` data (`/home/rtsprec`) and S3 data (`/data`). Alternative: bind mounts into `./data`. Rejected for the first version because bind mounts of a container user's files cause permission problems on macOS and Linux alike.
- **Makefile targets wrap `docker compose`** and check Docker first with the same message as `make test-integration`.
- **README.** The "Local Run" section gets a Docker variant that uses the targets. The Known Issues section lists: logout does not invalidate the token on the server; the Settings page does not save cameras or settings; the encoder path `/var/cache/rtsprec/video` is hard-coded. It links to `docs/KNOWN_ISSUES.md` for the rest.

## Risks / Trade-offs

- [SeaweedFS `latest` tag changes behavior] → Pin a tag in `compose.yaml` when first verified; update deliberately.
- [Virtual-host bucket addressing needs the bucket name as a network alias] → The `s3` service gets an alias for the bucket name, the same approach used in the integration helper.
- [Backend exits if the bucket is missing] → Compose waits for the S3 service and the init step before the backend starts (`depends_on` with `service_completed_successfully`).
- [Ports 8080 and 9000/8333 may be in use on the host] → Documented in the README troubleshooting; ports are overridable with environment variables in `compose.yaml`.
- [Build time of the backend image] → Layer caching in the Dockerfile is already in place; the first build is slow.

## Migration Plan

1. Add `compose.yaml`, verify `make docker-up` on a machine with Docker running.
2. Add Makefile targets and README sections.
3. Rollback: delete `compose.yaml` and the targets; the non-Docker Local Run instructions stay valid.

## Open Questions

- None that change the approach. The exact `amazon/aws-cli` and SeaweedFS tags are chosen and pinned during tasks group 1.
