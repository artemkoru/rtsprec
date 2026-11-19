# Design: Enhanced CI with Linter, Makefile, and Better Test Separation

## Context

Current CI has basic lint/test/build jobs but lacks:
- Industry-standard Go linting (golangci-lint)
- Local developer workflow consistency (Makefile)
- Optimized test execution (separate unit/integration)
- Artifact reuse (integration tests use build output)
- Path filtering means some MRs skip quality checks

## Goals / Non-Goals

**Goals:**
- Add golangci-lint for comprehensive Go code quality
- Create Makefile for local dev parity with CI
- Split unit and integration tests into separate jobs
- Integration tests use build artifacts (avoid rebuild)
- Run all CI checks on every MR (remove path filtering)
- Improve CI feedback speed with parallel unit tests

**Non-Goals:**
- Performance optimization beyond basic parallelization
- Docker-based CI execution
- Cross-platform builds (Linux/macOS/Windows)
- Code coverage reporting (can be added later)

## Decisions

### Decision 1: Use golangci-lint for Linting

**Choice:** Use golangci-lint with curated preset (default Go linters)

**Rationale:**
- Industry standard for Go projects
- Comprehensive: 15+ linters (staticcheck, errcheck, unused, etc.)
- Fast and cacheable
- Configuration via `.golangci.yml`
- Integrates easily with CI

**Alternatives:**
- go vet only: Insufficient; misses many issues
- Multiple individual linters: Slower, complex config management
- No linting: Risk of code quality drift

**Configuration:**
- Run on refactored packages only (domain, service, handler, repository, app)
- Exclude vendor/ and internal/capture/ (pre-existing issues)
- Fail on any reported issues

### Decision 2: Makefile Structure

**Choice:** Makefile with targets: `fmt`, `lint`, `test`, `test-unit`, `test-integration`, `build`, `clean`

**Rationale:**
- Local developers run `make lint` to match CI
- Consistent commands across macOS, Linux
- Simple, discoverable targets
- No dependencies (built-in Make)
- Useful for CI orchestration

**Targets:**
```makefile
make fmt              # Format code with gofmt
make lint             # Run golangci-lint
make test             # All tests (unit + integration)
make test-unit        # Unit tests only
make test-integration # Integration tests only
make build            # Build binary
make clean            # Clean build artifacts
```

### Decision 3: Split Test Jobs with Artifact Reuse

**Choice:** 
- Job 1 (Build): Compile binary
- Job 2 (Unit Tests): Fast parallel tests (no dependency)
- Job 3 (Integration Tests): Depends on Build, uses artifact

**Rationale:**
- Unit tests give fast feedback
- Integration tests reuse binary (no rebuild)
- Parallel execution: Build + Unit tests run simultaneously
- Clear separation of concerns
- Artifact sharing via GitHub Actions cache/upload

**Implementation:**
```yaml
jobs:
  build:
    - go build ./cmd/rtsprec
    - Upload binary as artifact
  
  unit-test:
    - go test ./internal/domain ./internal/repository ...
  
  integration-test:
    - needs: build
    - Download binary artifact
    - Run integration tests with binary
```

### Decision 4: Remove Path Filtering

**Choice:** Run all CI jobs on every MR (no path filtering)

**Rationale:**
- Ensures quality gates always apply
- Prevents skipping checks due to path mismatches
- Simpler CI logic
- Cost is minimal (jobs complete in <2 min)
- More reliable for detecting cross-module impacts

**Trade-off:** Slightly slower CI for guaranteed quality

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| golangci-lint rules too strict | Tune `.golangci.yml` to project style; can disable specific linters if needed |
| Makefile unfamiliar to new devs | Document in README; `make help` target shows all commands |
| Integration test flakiness | Use stable test infrastructure; mock external dependencies |
| CI always runs (no path filter) | Accept small cost; better reliability than conditional checks |

## Migration Plan

**Deploy steps:**
1. Create Makefile at repo root with 7 targets
2. Create `.golangci.yml` with linter configuration
3. Update `.github/workflows/backend.yml`:
   - Add lint job with golangci-lint
   - Split test into unit-test and integration-test jobs
   - Add build artifact upload/download
   - Remove path filtering (all jobs always run)
4. Test locally: `make lint`, `make test-unit`, `make test-integration`
5. Create PR and verify CI runs all jobs
6. Commit and merge

**Rollback:** Revert Makefile, .golangci.yml, and workflow changes; restore path filtering

**No special infrastructure required.**

## Open Questions

- Should integration tests use Docker for external services? (Deferred; can add later)
- Should we add code coverage tracking? (Deferred; separate change)
- Performance: should we add caching for go mod? (Can optimize incrementally)
