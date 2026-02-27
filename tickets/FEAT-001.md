# FEAT-001: Go module, Docker Compose, and Config package

## Description

Set up the foundational infrastructure for the Feature Flags REST API project. This includes initializing the Go module, setting up local development containers, and implementing a configuration management system.

## Specifications

### 1. Go Module Initialization
- Create `go.mod` with module name `github.com/xNakero/feature-flags`
- Specify Go version 1.23 or higher
- Add required dependencies:
  - `github.com/jackc/pgx/v5` - PostgreSQL driver
  - `github.com/redis/go-redis/v9` - Redis client
  - `github.com/testcontainers/testcontainers-go` - Docker test containers
  - `github.com/stretchr/testify` - Testing assertions
- Run `go mod verify` to ensure module integrity

### 2. Docker Compose Configuration
- Create `docker-compose.yml` at repository root
- Services:
  - **PostgreSQL 16-alpine**:
    - Environment: `POSTGRES_DB=featureflags`, `POSTGRES_USER=featureflags`, `POSTGRES_PASSWORD=featureflags`
    - Port: 5432 (exposed)
    - Healthcheck: pg_isready
  - **Redis 7-alpine**:
    - Port: 6379 (exposed)
    - Healthcheck: redis-cli ping
- Both services should have healthchecks configured

### 3. Config Package
- Location: `internal/config/config.go`
- Create `Config` struct with fields:
  - `HTTPAddr string` - HTTP server address (default: `:8080`)
  - `PostgresDSN string` - Postgres connection string (required)
  - `RedisAddr string` - Redis address (default: `localhost:6379`)
  - `LogLevel string` - Logging level (default: `info`)
- Implement `Load()` function that:
  - Reads environment variables (uppercase, snake_case with prefix if needed)
  - Returns error if `POSTGRES_DSN` is missing
  - Applies defaults for optional fields
  - Returns `*Config` on success

### 4. Config Tests
- Location: `internal/config/config_test.go`
- Test cases:
  - Happy path: all required and optional vars set correctly
  - Missing required var: `POSTGRES_DSN` absent should error
  - Custom values: verify non-default values are applied
  - Defaults: verify default values applied when env vars missing

## Acceptance Criteria

- [ ] `go.mod` exists with correct module name and Go version
- [ ] All four dependencies added and verified with `go mod verify`
- [ ] `docker-compose.yml` contains both Postgres and Redis services with correct configuration
- [ ] Both containers have healthchecks configured
- [ ] `internal/config/config.go` implements `Load()` function with all four config fields
- [ ] All config tests pass: `go test ./internal/config`
- [ ] `go build ./...` succeeds with no errors
- [ ] `go vet ./...` passes with no warnings
- [ ] Docker Compose services start successfully: `docker compose up -d`
