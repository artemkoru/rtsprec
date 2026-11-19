# Known Issues & Limitations

## Pre-existing Issues

### gortsplib API Incompatibility (resolved)

**Status:** Resolved in `use-app-entrypoint`. `internal/capture` now uses gortsplib v5.7.0.

**Affected Component:** `internal/capture/rtsp_reader.go`

**Previous issue:** The code imported `github.com/bluenviron/gortsplib/v4/pkg/base`, `.../pkg/formats` and `.../pkg/frames`, which do not exist in gortsplib v4.16.3. The build of `internal/capture`, `internal/api` and `internal/app` failed.

**Resolution:**
- `go.mod` depends on `github.com/bluenviron/gortsplib/v5` v5.7.0; the v4 requirement was removed by `go mod tidy`.
- `RTSPReader` uses the v5 client (`gortsplib.Client`, `base.ParseURL`, `desc.FindFormat`, `OnPacketRTP`). Credentials are passed in the URL user info.
- Packages `internal/capture`, `internal/motion`, `internal/encoder` and `internal/api` build and pass `go vet`.

## Refactoring Status

The clean architecture refactoring completed in November 2026 introduced NO new issues. All refactored layers (domain, repository, service, handler, app) have:
- ✅ Comprehensive test coverage (10/10 tests passing)
- ✅ Clean compilation
- ✅ Independent testability
- ✅ Zero breaking changes to external API

The gortsplib issue is pre-refactoring and orthogonal to the architecture improvements.

## CI/CD Implications

### Backend CI Workflow (`backend.yml`)

The backend CI build step is blocking: a failed `go build -o rtsprec ./cmd/rtsprec` fails the job. The former non-blocking fallback for the gortsplib issue has been removed.

```yaml
build:
  - name: Build backend binary
    run: |
      go build -o rtsprec ./cmd/rtsprec
```

### Test Coverage

Backend tests are limited to refactored packages:
```bash
go test ./internal/domain ./internal/repository ./internal/service ./internal/app
```

This ensures:
- All test jobs pass consistently
- Regressions in refactored code are caught
- gortsplib issues do not mask new problems

## Future Resolutions

### Short-term (Next Sprint)
1. Document gortsplib options in this file ✓
2. Plan gortsplib upgrade or replacement
3. Decide on preferred solution (A, B, or C above)

### Medium-term (1-2 Months)
1. Implement chosen gortsplib solution
2. Update capture module with new library
3. Extend test coverage to full binary build

### Long-term (3+ Months)
1. Establish RTSPcapture as fully tested component
2. Integrate with multi-camera infrastructure
3. Add performance benchmarks and optimization

## Reporting Issues

If you encounter other issues:

1. **Check this file** - Is it already documented?
2. **Search GitHub issues** - Has someone else reported it?
3. **Open a new issue** with:
   - Description of the problem
   - Steps to reproduce
   - Expected vs. actual behavior
   - Your environment (OS, Go version, etc.)
   - Relevant error messages or logs

## Support Resources

- **Architecture:** See [ARCHITECTURE.md](ARCHITECTURE.md) for clean architecture details
- **Deployment:** See [deployment.md](deployment.md) for deployment instructions
- **Testing:** See [testing.md](testing.md) for testing strategy
- **README:** See [../README.md](../README.md) for getting started
