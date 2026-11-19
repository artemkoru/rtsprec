# Tasks

## 1. Compose stack

- [x] 1.1 Create `compose.yaml` at the repository root with services `backend` (build from root `Dockerfile`, publish `8080`), `s3` (`chrislusf/seaweedfs` pinned tag, `server -s3 -s3.domainName=s3 -dir=/data`, network alias for the bucket name, named volume `s3-data`), and `s3-init` (`amazon/aws-cli` pinned tag, `aws --endpoint-url http://s3:8333 s3 mb s3://<bucket>` ignoring "already exists"); verify `docker compose config` validates and lists the three services
- [x] 1.2 Configure `backend` environment (`DB_PATH`, `ENCRYPTION_KEY_PATH`, `HLS_OUTPUT_PATH` under `/home/rtsprec`, `HTTP_PORT`, `AUTH_*`, `S3_*`, `AWS_ENDPOINT_URL=http://s3:8333`, `AWS_EC2_METADATA_DISABLED=true`), named volume `backend-data` for `/home/rtsprec`, and `depends_on` so the backend starts after `s3-init` completes successfully; verify `docker compose up -d` brings all services up and `curl -s -o /dev/null -w '%{http_code}' localhost:8080/api/cameras` returns 401
- [x] 1.3 Verify data survives a restart: add a camera through the API, run `docker compose down` (without `-v`) and `docker compose up -d`, then list cameras with a token; verify the camera is still there

## 2. Makefile targets

- [x] 2.1 Add `docker-build` (builds the backend image through compose), `docker-up` (starts the stack and waits until `GET /api/cameras` returns 401, with a timeout and a clear message), `docker-down`, and `docker-logs`; each checks that Docker is running first with the same message as `make test-integration`; add them to `help`; verify `make docker-build` succeeds and `make help` lists the new targets
- [x] 2.2 Verify `make docker-up` on a stopped stack starts it and prints the URL, and `make docker-down` stops it; verify `make docker-down` with Docker stopped prints the Docker message instead of a Docker error

## 3. README

- [x] 3.1 Add a Docker variant to the "Local Run" section: prerequisites (Docker with compose), `make docker-up`, that the UI still runs on the host with `npm run dev` and reaches the containerized backend on `localhost:8080`, and how to stop and reset data (`docker compose down -v`); verify the commands in the section match the Makefile targets
- [x] 3.2 Replace the "Known Issues & Troubleshooting" section about gortsplib v4 with a short list of current limitations (logout does not invalidate the token on the server; Settings does not save cameras or settings yet; the encoder path `/var/cache/rtsprec/video` is hard-coded) and a link to `docs/KNOWN_ISSUES.md`; verify `grep -n "gortsplib v4" README.md` returns nothing and the link target exists

## 4. Verification

- [x] 4.1 Run `make lint` and `make test-unit`; verify both pass
- [x] 4.2 Run the web UI on the host (`npm run dev` in `web/`) against the compose backend and sign in at `http://localhost:5173/login`; verify the Live page renders and the archive request reaches the backend
- [ ] 4.3 Open a pull request and verify the backend and frontend CI workflows pass

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
