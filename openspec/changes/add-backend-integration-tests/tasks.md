# Tasks

## 1. Backend image

- [x] 1.1 Add `Dockerfile` at the repository root: multi-stage build on `golang:1.26.0` that runs `go build -o /out/rtsprec ./cmd/rtsprec`, and a runtime stage that runs the binary as a non-root user; verify `docker build -t rtsprec-backend .` succeeds
- [x] 1.2 Verify the image starts with minimal environment (`DB_PATH`, `S3_BUCKET`, `S3_REGION`, `HTTP_PORT`, `AUTH_BASIC_*`, `ENCRYPTION_KEY_PATH` pointing at a writable directory, `AWS_ENDPOINT_URL` pointing at a reachable S3 endpoint) and answers `GET /api/cameras` with 401; verify with `docker run` against a MinIO container started by hand

## 2. Integration module

- [x] 2.1 Create `integration/go.mod` as a separate module that requires `github.com/testcontainers/testcontainers-go` and its MinIO module, and does not change the root `go.mod`; verify `go list -m all` in the root module does not list testcontainers
- [x] 2.2 Add a helper that starts a MinIO container, creates the test bucket, builds or reuses the backend image, starts the backend with environment variables pointing at MinIO, and waits until `GET /api/cameras` returns 401; registers cleanup with `t.Cleanup`; verify a smoke test starts and stops both containers
- [ ] 2.3 Verify the helper fails with a clear message when Docker is not reachable; verify by running the test with `DOCKER_HOST` set to an invalid socket and checking the message

## 3. Scenarios

- [x] 3.1 Test that `POST /api/auth/login` with correct credentials returns 200 and a token, and with wrong credentials returns 401 from the handler; verify both assertions pass against the running container
- [x] 3.2 Test that `GET /api/cameras` returns 401 without a token and 200 with the login token; verify both status codes
- [x] 3.3 Test that `GET /api/config/s3` with a token returns 200 and a JSON body; verify the status code and that the body parses as JSON
- [x] 3.4 Test that `GET /api/cameras/{id}/status` for an unknown ID returns 404 with the body "camera not found"; verify the status and body
- [x] 3.5 Test that `POST /api/auth/logout` returns 200 without a token; verify the status code
- [x] 3.6 Test that the container stops cleanly on SIGTERM: container exit code is 0 and its logs contain "shutdown complete"; verify both

## 4. Make and CI

- [ ] 4.1 Change `make test-integration` to check that Docker is available and then run `go test -count=1 -v ./...` in `integration/`; verify it fails with a clear message when Docker is stopped and passes when it runs
- [x] 4.2 Keep `make test-integration` out of `make test`'s unit dependency so the unit tests are not run twice; verify `make test-unit` and `make test-integration` each run their own tests once
- [x] 4.3 Update the `integration-test` job in `.github/workflows/backend.yml` to call `make test-integration` only; verify the YAML parses and the job lists no `go test` commands

## 5. Verification

- [x] 5.1 Run `make lint` and `make test-unit` to confirm the root module is unchanged; verify both pass
- [x] 5.2 Run `make test-integration` locally on a machine with Docker; verify all scenarios pass
- [ ] 5.3 Open a pull request and verify the `Integration Tests` job passes in GitHub Actions

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
