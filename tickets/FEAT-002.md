# FEAT-002: CI Pipeline

## Description

Implement GitHub Actions continuous integration pipeline for automated testing, linting, and build verification on every push and pull request.

## Specifications

### CI Workflow
- File location: `.github/workflows/ci.yml`
- Triggers: `push` and `pull_request` to `master` branch
- Runs on: `ubuntu-latest`

### Pipeline Steps
1. **Checkout**: Clone repository with full history
2. **Setup Go**: Install Go 1.23+
3. **Build**: Run `go build ./...`
4. **Lint**: Run `go vet ./...`
5. **Test**: Run `go test -race -count=1 ./...`
   - Exclude integration tests (tests that require `testcontainers` or external services)
   - Use `-race` flag to detect race conditions
   - Use `-count=1` to disable caching

### Configuration
- Set appropriate timeout for each step
- Fail fast on first error
- No additional environment variables needed for basic workflow
- Build should be reproducible and deterministic

## Acceptance Criteria

- [ ] `.github/workflows/ci.yml` exists in repository
- [ ] Workflow triggers on push to master
- [ ] Workflow triggers on pull requests to master
- [ ] All 5 steps (checkout, setup Go, build, vet, test) present and ordered correctly
- [ ] `go build ./...` step runs successfully
- [ ] `go vet ./...` step runs successfully
- [ ] `go test -race -count=1 ./...` step runs successfully
- [ ] Workflow completes without errors when pushed
- [ ] Integration tests are excluded from CI runs (by build tags or file naming)
