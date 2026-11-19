# Proposal: Add GitHub CI and README

## Why

The project has just completed a comprehensive clean architecture refactoring but lacks automated testing infrastructure and onboarding documentation. GitHub Actions CI ensures code quality and regression detection on every push and PR. A comprehensive README lowers the barrier to entry for developers and operators, covering both quick-start and deep architectural understanding.

## What Changes

**New Infrastructure:**
- GitHub Actions workflows for unit tests, integration tests, linting, and builds
- Multi-stage CI: test refactored packages (guaranteed pass), lint, build binary (with gortsplib workaround documentation)

**New Documentation:**
- `README.md`: Project overview, quick start, architecture summary, configuration, testing, deployment, known issues
- Consolidates information from `config.example.yaml`, `docs/ARCHITECTURE.md`, and project structure

## Capabilities

### New Capabilities
None. This is pure infrastructure and documentation (no spec-level behavior changes).

### Modified Capabilities
None.

## Impact

**Code:**
- No changes to application code
- No API changes

**Infrastructure:**
- Adds `.github/workflows/` directory with CI pipeline definitions
- Root-level `README.md` file

**Dependencies:**
- No new dependencies
- CI runs on existing Go, test, and build tooling

**Known Issues Addressed:**
- Workflow documents pre-existing gortsplib v4 API incompatibility in capture package (pre-refactoring issue, not in scope of this change)
- CI tests only refactored packages which all pass cleanly
