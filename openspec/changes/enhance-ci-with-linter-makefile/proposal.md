# Proposal: Enhance CI with Linter, Makefile, and Better Test Separation

## Why

The current CI setup lacks industry-standard Go linting, local developer workflows are manual and error-prone, and test execution is not optimized. A Makefile provides consistent commands for local development and CI. golangci-lint catches bugs and style issues early. Separating unit and integration tests with artifact reuse improves CI efficiency and debugging. Running all checks on every MR (regardless of path changes) ensures quality gates are always applied.

## What Changes

**New Infrastructure:**
- Add `Makefile` with targets: `fmt`, `lint`, `test`, `test-unit`, `test-integration`, `build`, `clean`
- Integrate golangci-lint into backend CI pipeline
- Split backend CI into separate unit test and integration test jobs
- Integration tests use build artifacts from previous step
- Remove path filtering: all CI jobs run on every MR for consistent quality gates

**Improved CI/CD:**
- Unit tests run in parallel (fast feedback)
- Integration tests depend on build job and use binary artifact
- golangci-lint provides comprehensive code quality checks
- Makefile ensures local `make lint` matches CI behavior

## Capabilities

### New Capabilities
None. This is pure infrastructure and tooling (no spec-level behavior changes).

### Modified Capabilities
None.

## Impact

**Code:**
- No changes to application code
- No API changes
- Makefile is developer convenience (optional for CI)

**CI/CD:**
- Workflow changes: `.github/workflows/backend.yml` modified to add linter, split tests, remove path filtering
- All checks now run on every MR (stricter but more reliable)
- Faster local development with `make` targets

**Developer Experience:**
- `make lint` runs same checks as CI
- `make test-unit` and `make test-integration` for local testing
- `make build` for consistent binary creation
- `make clean` for workspace reset

**Performance:**
- Parallel unit test jobs (faster CI feedback)
- Build artifact reuse for integration tests (avoid rebuild)
