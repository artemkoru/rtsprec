# Tasks

## 1. Backend GitHub Actions Workflow

- [x] 1.1 Create `.github/workflows/` directory structure; verify directory exists and is empty
- [x] 1.2 Create `backend.yml` workflow with path filter for `cmd/`, `internal/`, `go.mod`, `go.sum`; verify workflow triggers on push to main and PRs only when backend files change
- [x] 1.3 Add lint job to `backend.yml` that runs `go vet` and `gofmt -d` on refactored packages; verify lint commands check domain, service, handler, repository, app packages
- [x] 1.4 Add test job to `backend.yml` that runs `go test ./internal/domain ./internal/repository ./internal/service ./internal/handler ./internal/app`; verify test commands are present
- [x] 1.5 Add build job to `backend.yml` that attempts `go build ./cmd/rtsprec` with gortsplib workaround; verify workflow handles capture package failure with warning message

## 2. Frontend GitHub Actions Workflow

- [x] 2.1 Create `frontend.yml` workflow with path filter for `web/`; verify workflow triggers on push to main and PRs only when web files change
- [x] 2.2 Add install job to `frontend.yml` that runs `npm ci` in `/web`; verify npm dependencies install correctly
- [x] 2.3 Add lint job to `frontend.yml` that runs `npm run lint` (oxlint); verify linting configuration is detected and runs
- [x] 2.4 Add build job to `frontend.yml` that runs `npm run build`; verify build artifacts are created successfully

## 3. README.md Documentation

- [x] 3.1 Create `README.md` with sections: Overview, Quick Start, Features; verify file exists and each section is populated
- [x] 3.2 Add Backend section with quick start for Go backend; verify section includes Go version requirement and build instructions
- [x] 3.3 Add Frontend section with quick start for React web UI; verify section includes npm setup and build instructions
- [x] 3.4 Add Architecture section linking to `docs/ARCHITECTURE.md`; verify link is functional
- [x] 3.5 Add Configuration section referencing `config.example.yaml`; verify section explains key settings for backend
- [x] 3.6 Add Testing & Development section; verify section explains how to run backend and frontend tests
- [x] 3.7 Add Deployment section linking to `deployment.md`; verify link is functional
- [x] 3.8 Add Known Issues section documenting gortsplib incompatibility; verify section explains the issue and references workaround
- [x] 3.9 Add CI status badges for both backend and frontend workflows; verify badges link to GitHub Actions and show current status

## 4. Known Issues Documentation

- [x] 4.1 Create `docs/KNOWN_ISSUES.md` documenting pre-existing gortsplib v4 API incompatibility; verify file explains issue, impact, and workaround
- [x] 4.2 Reference KNOWN_ISSUES.md from README; verify cross-link exists

## 5. Verification and Integration

- [x] 5.1 Trigger backend.yml manually and verify all jobs pass (lint, test, build); verify workflow runs successfully on main
- [x] 5.2 Trigger frontend.yml manually and verify all jobs pass (install, lint, build); verify workflow runs successfully on main
- [x] 5.3 Create a test PR changing only backend files and verify only backend.yml triggers; verify frontend.yml does not run
- [x] 5.4 Create a test PR changing only frontend files and verify only frontend.yml triggers; verify backend.yml does not run
- [x] 5.5 Review README end-to-end for clarity; verify no dead links and all sections make sense for new developers and operators

## Workflow follow-up

- Archive the change after CI is working and README is reviewed
- Monitor CI workflow execution on first few commits to main to ensure stability
