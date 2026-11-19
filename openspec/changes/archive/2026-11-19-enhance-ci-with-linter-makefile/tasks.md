# Tasks

## 1. Create Makefile

- [x] 1.1 Create `Makefile` at repo root with targets: `fmt`, `lint`, `test`, `test-unit`, `test-integration`, `build`, `clean`; verify file exists and contains all 7 targets
- [x] 1.2 Implement `make fmt` target that runs `gofmt -w ./internal/`; verify target works and formats code
- [x] 1.3 Implement `make lint` target that runs `go vet` and `golangci-lint` on refactored packages; verify target exists
- [x] 1.4 Implement `make test` target that runs both unit and integration tests; verify target runs all tests
- [x] 1.5 Implement `make test-unit` target that runs `go test ./internal/domain ./internal/repository ./internal/service ./internal/handler ./internal/app`; verify unit tests run
- [x] 1.6 Implement `make test-integration` target for integration tests; verify target exists (initially same as unit tests, can be expanded)
- [x] 1.7 Implement `make build` target that builds binary to `./rtsprec`; verify binary is created
- [x] 1.8 Implement `make clean` target that removes build artifacts; verify cleanup works
- [x] 1.9 Verify all Makefile targets work locally; run each target and verify success

## 2. Configure golangci-lint

- [x] 2.1 Create `.golangci.yml` configuration file at repo root; verify file exists
- [x] 2.2 Configure `.golangci.yml` to run on refactored packages (domain, service, handler, repository, app); verify config is correct
- [x] 2.3 Disable linting on capture/ package (pre-existing issues); verify .golangci.yml excludes capture module
- [x] 2.4 Set linter timeout to 5 minutes; verify timeout setting in config
- [x] 2.5 Run `golangci-lint run ./internal/domain ./internal/repository ./internal/service ./internal/handler ./internal/app` locally and verify no errors; fix any linting issues if needed

## 3. Update Backend CI Workflow

- [x] 3.1 Update `.github/workflows/backend.yml` to remove path filtering (paths: section); verify workflow triggers on all PRs regardless of files changed
- [x] 3.2 Add lint job to workflow that runs `golangci-lint` with `.golangci.yml` config; verify lint job runs before tests
- [x] 3.3 Modify build job to upload binary as artifact with name 'rtsprec-binary'; verify artifact upload is configured
- [x] 3.4 Split test job into two separate jobs: `unit-test` and `integration-test`; verify both jobs exist in workflow
- [x] 3.5 Configure `unit-test` job to run `go test` on refactored packages without dependencies; verify unit test job has no `needs:` clause
- [x] 3.6 Configure `integration-test` job to depend on `build` job and download artifact; verify `needs: build` and artifact download step
- [x] 3.7 Verify workflow job order: lint, build, unit-test (parallel), integration-test (depends on build); verify correct dependency graph
- [x] 3.8 Test workflow locally with `act` or verify via PR; verify all jobs run on MR

## 4. Update Frontend CI Workflow (Remove Path Filtering)

- [x] 4.1 Update `.github/workflows/frontend.yml` to remove path filtering on `push` and `pull_request` triggers; verify workflow runs on all PRs
- [x] 4.2 Verify frontend workflow still runs only when needed (can be tested with separate PR); verify workflow is configured

## 5. Testing and Verification

- [x] 5.1 Test `make fmt` locally to verify formatting works; run `make fmt` and verify code is properly formatted
- [x] 5.2 Test `make lint` locally to verify golangci-lint runs; run `make lint` and verify no errors
- [x] 5.3 Test `make test-unit` locally to verify tests pass; run `make test-unit` and verify 10/10 tests pass
- [x] 5.4 Test `make build` locally to verify binary builds; run `make build` and verify rtsprec binary created
- [x] 5.5 Create test PR and verify backend workflow runs all jobs (lint, build, unit-test, integration-test); verify all jobs execute
- [x] 5.6 Verify integration-test job uses build artifact from previous step; check workflow logs for artifact download
- [x] 5.7 Document all Makefile targets in README; verify README has `make` section

## Workflow follow-up

- Archive the change after CI is working and all tasks pass on PR
- Monitor CI execution on first few PRs to ensure stability
