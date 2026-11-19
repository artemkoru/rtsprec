# Design: GitHub CI and README

## Context

The project has completed a clean architecture refactoring with comprehensive test coverage and documentation. The refactored packages (domain, service, handler, repository, app) all build and test cleanly. A pre-existing gortsplib v4 API incompatibility exists in the legacy capture package (not part of the refactoring scope). Currently, there is no automated testing, no onboarding documentation, and no CI/CD pipeline.

## Goals / Non-Goals

**Goals:**
- Automatically verify code quality (lint, format, vet) on every push and PR
- Run unit and integration tests on refactored packages on every push and PR
- Build and cache the binary on main branch for deployment
- Provide new developers and operators a single entry point (README.md) to understand the project
- Document known issues and workarounds

**Non-Goals:**
- Fix pre-existing gortsplib incompatibility (out of scope)
- Full end-to-end deployment pipeline (can be added later)
- Performance benchmarking or profiling CI jobs
- Docker image building (optional, can be added later)

## Decisions

### Decision 1: GitHub Actions as CI Platform

**Choice:** Use GitHub Actions (not external CI service)

**Rationale:**
- Already hosted on GitHub, no external account needed
- Free for public repositories
- Native integration with PR workflow
- Simple YAML-based configuration
- Adequate for current project scale

**Alternatives considered:**
- GitLab CI/CD: Would require repo migration
- CircleCI/Travis: Adds external dependency; overkill for this stage
- Local pre-commit hooks only: Insufficient; needs centralized enforcement

### Decision 2: Separate Backend and Frontend Workflows

**Choice:** Create independent workflows for backend (Go) and frontend (React/Vite)

**Rationale:**
- Backend and frontend are independent applications with different toolchains
- Allows targeted testing and deployment for each
- Frontend can progress independently from backend gortsplib issue
- Clear separation of concerns in CI/CD pipeline
- Different dependencies and build requirements

**Backend Workflow (`backend.yml`):**
- Lint: `go vet`, `gofmt` on refactored packages
- Test: `go test` on domain, service, handler, repository, app packages
- Build: `go build ./cmd/rtsprec` (non-blocking on capture package error)

**Frontend Workflow (`frontend.yml`):**
- Install: `npm ci` in `/web`
- Lint: `npm run lint` (oxlint)
- Build: `npm run build`
- Test: `npm test` (if applicable)

**Triggers:** Both workflows on push to main and PRs

### Decision 3: Workflow Triggers and Path Filtering

**Choice:** 
- Backend workflow: Triggered on push to main, PRs, and changes to `cmd/`, `internal/`, `go.mod`, `go.sum`
- Frontend workflow: Triggered on push to main, PRs, and changes to `web/`

**Rationale:**
- Path filtering prevents unnecessary CI runs (backend workflow skips if only web files change)
- Faster feedback for developers
- Reduces CI resource usage
- Clear ownership: each workflow only runs when its files change

**Alternatives:**
- Run all workflows on every push: Wastes resources; frontend doesn't care about backend changes
- Manual trigger only: Misses automated quality gates

### Decision 4: README Structure

**Choice:** Single README.md with sections for:
1. Quick start (clone, config, run)
2. Architecture (links to docs/ARCHITECTURE.md)
3. Features (what the system does)
4. Configuration (reference to config.example.yaml)
5. Testing and development
6. Deployment
7. Known issues and troubleshooting

**Rationale:**
- GitHub displays README.md on project main page
- Covers both operators (deploy) and developers (contribute)
- Cross-links to detailed docs without duplicating them
- Provides context for first-time visitors

**Alternatives:**
- Separate CONTRIBUTING.md, OPERATORS.md: Splits information; first-time visitors still need direction
- Minimal README with links to everything: Insufficient onboarding

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| gortsplib build failure breaks developer workflow | Explicit warning in README + CI logs; docs point to workaround |
| README becomes stale as project evolves | CONTRIBUTING section with "keep README fresh" guideline; link to ARCHITECTURE.md for detailed updates |
| CI jobs slow down as project grows | Monitor job times; parallelize test jobs if needed |
| Linting/fmt failures block CI unnecessarily | Tune rules to match project style; provide `make fmt` target |

## Migration Plan

**Deploy steps:**
1. Create `.github/workflows/` directory structure
2. Add workflow files:
   - `backend.yml` - Backend CI: lint, test, build (Go); path-filtered to cmd/, internal/, go.mod, go.sum
   - `frontend.yml` - Frontend CI: lint, build (Node/npm); path-filtered to web/
3. Create `README.md` at root with separate backend and frontend getting-started sections
4. Create `docs/KNOWN_ISSUES.md` documenting gortsplib issue
5. Update CI badges in README (once passing)
6. Commit and push to `main`

**Rollback:** Delete workflow files and README.md; revert commits.

**No special deployment or infrastructure required.**

## Open Questions

- Should we add CodeCov or similar coverage tracking? (Can be deferred to follow-up change)
- Should we build Docker images in CI? (Deferred; can be added later)
- Do we want branch protection rules? (Deferred; depends on team process)
