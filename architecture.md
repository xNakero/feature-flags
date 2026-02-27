# Feature Flags Service — Architecture

## 1. Overview

The Feature Flags Service is a lightweight REST API that stores and serves feature flag values. It supports two flag types — **boolean** and **numeric** — and uses a two-storage model:

- **PostgreSQL** — authoritative store for all flag data including metadata.
- **Redis** — fast read cache for flag values on the hot read path.

Writes go to Postgres first (hard fail on error), then to Redis (soft fail: cache inconsistency is tolerated because reads fall back to Postgres and repopulate the cache automatically).

---

## 2. Architecture Style — Hexagonal (Ports & Adapters)

The service is structured in three concentric rings:

```
┌──────────────────────────────────────────────┐
│  Adapters (HTTP, Postgres, Redis)            │
│  ┌────────────────────────────────────────┐  │
│  │  Ports (interfaces)                   │  │
│  │  ┌──────────────────────────────────┐ │  │
│  │  │  Domain (Flag, FlagValue, errors)│ │  │
│  │  └──────────────────────────────────┘ │  │
│  └────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

- **`internal/domain/`** — zero external dependencies; defines core entities and error sentinels. Everything else depends on what is defined here.
- **`internal/port/`** — interfaces only; defines the boundaries between rings. Neither domain nor adapters depend on each other — they both depend on ports.
- **`internal/adapter/`** — concrete implementations of ports (HTTP handler, Postgres store, Redis cache). Know about `domain` and `port` but never about `service`.
- **`internal/service/`** — application logic. Depends on `domain` and `port`. Knows nothing about HTTP, Postgres, or Redis.
- **`cmd/server/main.go`** — single composition root; wires all adapters together and starts the server.

---

## 3. Package Structure

Tests live in the **same directory** as the code they test (standard Go convention). Integration tests carry a `//go:build integration` tag and are excluded from plain `go test ./...` runs.

```
feature-flags/
├── cmd/server/
│   ├── main.go                       # Composition root — wires adapters, starts HTTP server
│   └── main_test.go                  # End-to-end HTTP tests  [build tag: integration]
│
├── internal/
│   ├── domain/
│   │   ├── flag.go                   # Flag entity, FlagType enum, FlagValue type
│   │   ├── flag_test.go              # Domain logic unit tests
│   │   └── errors.go                 # Sentinel errors (ErrNotFound, ErrAlreadyExists, …)
│   │
│   ├── port/
│   │   ├── service.go                # FlagService interface (primary / inbound port)
│   │   ├── store.go                  # FlagStore interface (secondary / outbound → Postgres)
│   │   └── cache.go                  # FlagCache interface (secondary / outbound → Redis)
│   │
│   ├── service/
│   │   ├── flag_service.go           # FlagService implementation (core application logic)
│   │   └── flag_service_test.go      # Unit tests with hand-written fakes
│   │
│   ├── adapter/
│   │   ├── http/
│   │   │   ├── handler.go            # HTTP handler (primary adapter)
│   │   │   ├── handler_test.go       # Unit tests via net/http/httptest
│   │   │   ├── router.go             # Route registration
│   │   │   ├── dto.go                # Request / response structs + JSON marshaling
│   │   │   └── middleware.go         # Logging, recovery, request-ID middleware
│   │   │
│   │   ├── postgres/
│   │   │   ├── flag_store.go         # FlagStore implementation (secondary adapter)
│   │   │   ├── flag_store_test.go    # Integration tests  [build tag: integration]
│   │   │   └── migrations/
│   │   │       └── 001_create_flags.sql
│   │   │
│   │   └── redis/
│   │       ├── flag_cache.go         # FlagCache implementation (secondary adapter)
│   │       └── flag_cache_test.go    # Integration tests  [build tag: integration]
│   │
│   ├── testutil/
│   │   ├── db.go                     # Shared helper: Postgres container bootstrap
│   │   └── redis.go                  # Shared helper: Redis container bootstrap
│   │
│   └── config/
│       └── config.go                 # Env-var loading (DB DSN, Redis addr, port, …)
│
├── docker-compose.yml                # Local dev: Postgres + Redis
├── .env.example
└── architecture.md
```

**Why `internal/testutil/` and not a top-level `test/` directory:**

- Test helpers shared across packages belong in `internal/testutil/` — importable by any `internal/` package without being a public API.
- A top-level `test/` directory is a Java/Python convention and is not idiomatic in Go. Go's toolchain discovers tests purely by the `_test.go` file suffix within each package directory.
- The `//go:build integration` tag separates integration tests from unit tests without directory games:
  - `go test ./...` — runs only unit tests.
  - `go test -tags integration ./...` — runs everything.

---

## 4. Domain Model

### `internal/domain/flag.go`

```go
package domain

import "time"

type FlagType string

const (
    FlagTypeBoolean FlagType = "boolean"
    FlagTypeNumeric FlagType = "numeric"
)

// FlagValue holds the actual value. Exactly one field is non-nil,
// determined by the parent Flag.Type. Pointer fields ensure that
// false and 0.0 are distinguishable from "not set".
type FlagValue struct {
    Boolean *bool
    Numeric *float64
}

type Flag struct {
    Name        string    // natural primary key, e.g. "dark-mode"
    Type        FlagType
    Description string
    Value       FlagValue
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Flag naming rules:** must match `^[a-z][a-z0-9\-]{0,62}$` — starts with a lowercase letter, contains only lowercase letters, digits, and hyphens, maximum 63 characters total.

### `internal/domain/errors.go`

```go
package domain

import "errors"

var ErrNotFound      = errors.New("flag not found")
var ErrAlreadyExists = errors.New("flag already exists")
var ErrTypeMismatch  = errors.New("value type does not match flag type")
var ErrInvalidFlagName = errors.New("invalid flag name")
var ErrInvalidValue  = errors.New("invalid flag value")
```

Callers use `errors.Is` so that wrapped errors (`fmt.Errorf("postgres: %w", domain.ErrNotFound)`) are handled correctly.

---

## 5. Port Interfaces

### `internal/port/service.go` — Primary (inbound) port

```go
type FlagService interface {
    CreateFlag(ctx context.Context, flag domain.Flag) (domain.Flag, error)
    GetFlag(ctx context.Context, name string) (domain.Flag, error)
    GetFlagValue(ctx context.Context, name string) (domain.FlagValue, error)
    UpdateFlagValue(ctx context.Context, name string, value domain.FlagValue) (domain.Flag, error)
}
```

### `internal/port/store.go` — Secondary (outbound) port — Postgres

```go
type FlagStore interface {
    Save(ctx context.Context, flag domain.Flag) (domain.Flag, error)
    FindByName(ctx context.Context, name string) (domain.Flag, error)
    UpdateValue(ctx context.Context, name string, value domain.FlagValue) (domain.Flag, error)
}
```

### `internal/port/cache.go` — Secondary (outbound) port — Redis

```go
type FlagCache interface {
    Set(ctx context.Context, name string, value domain.FlagValue) error
    Get(ctx context.Context, name string) (domain.FlagValue, error) // ErrNotFound on miss
    Delete(ctx context.Context, name string) error                  // no-op if missing
}
```

`Get` returns `domain.ErrNotFound` on a cache miss — callers do not need to know whether the key was absent or Redis was unreachable; the fallback behaviour is identical.

---

## 6. Data Model

### PostgreSQL Schema

```sql
-- internal/adapter/postgres/migrations/001_create_flags.sql

CREATE TYPE flag_type AS ENUM ('boolean', 'numeric');

CREATE TABLE flags (
    name         TEXT        PRIMARY KEY
        CHECK (name ~ '^[a-z][a-z0-9\-]{0,62}$'),
    type         flag_type   NOT NULL,
    description  TEXT        NOT NULL DEFAULT '',
    bool_value   BOOLEAN,
    num_value    NUMERIC(20,6),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT value_matches_type CHECK (
        (type = 'boolean' AND bool_value IS NOT NULL AND num_value IS NULL)
        OR
        (type = 'numeric' AND num_value IS NOT NULL AND bool_value IS NULL)
    )
);

CREATE INDEX idx_flags_type ON flags (type);
```

**Design notes:**
- Two value columns (`bool_value`, `num_value`) rather than a single `JSONB` column: keeps the CHECK constraint in pure SQL, allows type-safe reads without deserialization, and makes columns indexable.
- `NUMERIC(20,6)` avoids floating-point rounding artefacts in the DB layer.
- The `CHECK` on `name` is a last line of defence; the service validates first so error messages are controlled.
- `updated_at` is set by the application layer (not a DB trigger) so integration tests can inject a fixed timestamp.

### Redis Key Schema

**Pattern:** `flags:value:{name}`

**Value encoding:** plain string with a type prefix.

| Flag type | Example stored value |
|-----------|---------------------|
| boolean   | `b:true` / `b:false` |
| numeric   | `n:30.5` / `n:100`  |

A single `GET` retrieves both the type discriminator and the value — no additional round-trips. Values are human-readable via `redis-cli GET`.

**TTL policy:** no TTL by default. Write-through keeps the cache consistent. A configurable `CACHE_TTL_SECONDS` can be added as a safety net against edge cases.

---

## 7. API Catalogue

All responses use `Content-Type: application/json`.

### Common Error Envelope

```json
{
  "error": {
    "code":    "NOT_FOUND",
    "message": "flag 'dark-mode' not found"
  }
}
```

---

### POST /flags — Create a flag

**Request body:**
```json
{
  "name":        "dark-mode",
  "type":        "boolean",
  "description": "Enables dark mode for all users",
  "value":       true
}
```

`value` is a raw JSON boolean or number — not a wrapped object. The HTTP adapter uses `type` to determine how to parse it.

**Success: 201 Created**
```json
{
  "name":        "dark-mode",
  "type":        "boolean",
  "description": "Enables dark mode for all users",
  "value":       true,
  "created_at":  "2026-02-27T10:00:00Z",
  "updated_at":  "2026-02-27T10:00:00Z"
}
```

**Errors:**

| HTTP | Code             | Condition                                    |
|------|------------------|----------------------------------------------|
| 400  | `INVALID_NAME`   | Name is empty, too long, or contains illegal characters |
| 400  | `INVALID_VALUE`  | Value missing or wrong JSON kind             |
| 400  | `INVALID_TYPE`   | Type not `"boolean"` or `"numeric"`          |
| 409  | `ALREADY_EXISTS` | A flag with this name already exists         |

---

### GET /flags/:name — Full flag detail

Always reads from PostgreSQL. Never touches Redis.

**Success: 200 OK**
```json
{
  "name":        "dark-mode",
  "type":        "boolean",
  "description": "Enables dark mode for all users",
  "value":       true,
  "created_at":  "2026-02-27T10:00:00Z",
  "updated_at":  "2026-02-27T10:00:00Z"
}
```

**Errors:**

| HTTP | Code        | Condition           |
|------|-------------|---------------------|
| 404  | `NOT_FOUND` | Flag does not exist |

---

### GET /flags/:name/value — Cached value read

Redis-first, Postgres fallback.

**Success: 200 OK**
```json
{ "name": "dark-mode",        "type": "boolean", "value": true }
{ "name": "request-timeout",  "type": "numeric",  "value": 30.5 }
```

**Errors:**

| HTTP | Code        | Condition           |
|------|-------------|---------------------|
| 404  | `NOT_FOUND` | Flag does not exist |

---

### PUT /flags/:name/value — Update flag value (write-through)

**Request body:**
```json
{ "value": false }
{ "value": 42.0  }
```

The server resolves the expected JSON kind from the flag's stored type. A mismatch returns `TYPE_MISMATCH`.

**Success: 200 OK** — returns the full updated flag detail (same shape as GET /flags/:name).

**Errors:**

| HTTP | Code            | Condition                                          |
|------|-----------------|----------------------------------------------------|
| 400  | `INVALID_VALUE` | Body is missing, malformed JSON, or value is null  |
| 400  | `TYPE_MISMATCH` | JSON kind of `value` contradicts flag's type       |
| 404  | `NOT_FOUND`     | Flag does not exist                                |

---

## 8. Write-Through Flow (PUT /flags/:name/value)

```
HTTP Client  →  PUT /flags/dark-mode/value  { "value": false }
                              │
                   ┌──────────▼──────────┐
                   │   HTTP Handler       │  1. Parse + validate request body
                   └──────────┬──────────┘
                              │
                   ┌──────────▼──────────┐
                   │   FlagService        │  2. FindByName → ErrNotFound? → 404
                   │                      │  3. Type match check → ErrTypeMismatch? → 400
                   │                      │  4. FlagStore.UpdateValue (Postgres)
                   │                      │     HARD FAIL if Postgres errors → do not touch Redis
                   │                      │  5. FlagCache.Set (Redis)
                   │                      │     SOFT FAIL: log WARN, return 200 anyway
                   └──────────┬──────────┘
                              │
                   ┌──────────▼──────────┐
                   │   HTTP Handler       │  6. Serialise updated Flag → 200 OK
                   └─────────────────────┘
```

**Redis failure asymmetry:**
- **Postgres failure (step 4):** hard fail — value not persisted, return 5xx, do not touch Redis.
- **Redis failure (step 5):** soft fail — value IS persisted in Postgres (source of truth). Log a WARN. The next `/value` read will miss Redis, fall back to Postgres, read the correct value, and repopulate the cache automatically. No data loss, no client-visible error on writes.

**POST /flags write flow** follows the same pattern: save to Postgres first, then populate cache (soft-fail on Redis error).

---

## 9. Read Flows

### GET /flags/:name — Full Detail

```
FlagService.GetFlag
  └─ FlagStore.FindByName (Postgres)
       ├─ Found  → return Flag → 200 OK
       └─ Missing → ErrNotFound → 404
```

Redis is never consulted. The full detail response includes metadata (`description`, `created_at`, `updated_at`) not cached in Redis; fetching that separately would add complexity for a non-hot-path endpoint.

### GET /flags/:name/value — Cached Value

```
FlagService.GetFlagValue
  └─ FlagCache.Get (Redis)
       ├─ Cache HIT  → decode prefix → return FlagValue → 200 OK  (no Postgres query)
       └─ Cache MISS or Redis down → ErrNotFound
            └─ FlagStore.FindByName (Postgres)
                 ├─ Found  → FlagCache.Set (soft-fail) → return FlagValue → 200 OK
                 └─ Missing → ErrNotFound → 404
```

---

## 10. Error Catalogue

### Domain Errors → HTTP Mapping

```go
func domainErrToHTTP(err error) (int, string) {
    switch {
    case errors.Is(err, domain.ErrNotFound):        return 404, "NOT_FOUND"
    case errors.Is(err, domain.ErrAlreadyExists):   return 409, "ALREADY_EXISTS"
    case errors.Is(err, domain.ErrTypeMismatch):    return 400, "TYPE_MISMATCH"
    case errors.Is(err, domain.ErrInvalidFlagName): return 400, "INVALID_NAME"
    case errors.Is(err, domain.ErrInvalidValue):    return 400, "INVALID_VALUE"
    default:                                        return 500, "INTERNAL_ERROR"
    }
}
```

`errors.Is` unwraps chains correctly — a `fmt.Errorf("postgres: %w", domain.ErrNotFound)` still maps to 404 without leaking infrastructure detail into the response.

### Infrastructure Errors

| Situation | Handling |
|-----------|----------|
| Postgres connection lost | Propagated as infrastructure error → 503 |
| Redis connection lost (read) | Cache adapter returns `domain.ErrNotFound` → triggers Postgres fallback → transparent to caller |
| Redis connection lost (write) | Cache adapter returns non-domain error → service logs WARN, suppresses it → 200 returned |
| Unknown DB error | Propagated → 500 |
| Request body too large | Middleware enforces a max body size (e.g. 64 KB) → 413 |

---

## 11. Testing Strategy

Every `_test.go` file lives in the same directory as the file it tests. There is no top-level `test/` directory.

### Unit Tests — `go test ./...`

No external dependencies. Hand-written fakes (preferred over generated mocks — the interfaces are small):

**`internal/domain/flag_test.go`**
- `FlagValue` zero-value correctness
- Name validation rules (length, characters, empty)
- Error sentinel identity via `errors.Is`

**`internal/service/flag_service_test.go`** — fakes for `FlagStore` and `FlagCache`:

| Test case | Assertion |
|-----------|-----------|
| `CreateFlag` happy path | Store.Save called, Cache.Set called, returned flag matches input |
| `CreateFlag` invalid name | `ErrInvalidFlagName` returned, Store.Save not called |
| `CreateFlag` already exists | `ErrAlreadyExists` propagated |
| `CreateFlag` Redis Set fails | Store.Save succeeded, no error returned to caller |
| `GetFlag` found | Store.FindByName called, result returned |
| `GetFlag` not found | `ErrNotFound` propagated |
| `GetFlagValue` cache hit | Store.FindByName NOT called |
| `GetFlagValue` cache miss | Store.FindByName called, Cache.Set called |
| `GetFlagValue` cache down + found in DB | Same as cache miss |
| `GetFlagValue` cache down + not in DB | `ErrNotFound` propagated |
| `UpdateFlagValue` happy path | Store.UpdateValue called, Cache.Set called |
| `UpdateFlagValue` not found | `ErrNotFound` from store, Cache.Set not called |
| `UpdateFlagValue` boolean type mismatch | `ErrTypeMismatch`, Store.UpdateValue not called |
| `UpdateFlagValue` numeric type mismatch | `ErrTypeMismatch`, Store.UpdateValue not called |
| `UpdateFlagValue` Redis Set fails | Store.UpdateValue succeeded, no error returned |

**`internal/adapter/http/handler_test.go`** — fake `FlagService`, `net/http/httptest`:

| Test | Assertion |
|------|-----------|
| POST /flags 201 | correct JSON shape |
| POST /flags 400 bad name | error code `INVALID_NAME` |
| POST /flags 400 unknown type | error code `INVALID_TYPE` |
| POST /flags 409 | error code `ALREADY_EXISTS` |
| GET /flags/:name 200 | all fields present including timestamps |
| GET /flags/:name 404 | error code `NOT_FOUND` |
| GET /flags/:name/value 200 boolean | `value` is JSON bool |
| GET /flags/:name/value 200 numeric | `value` is JSON number |
| GET /flags/:name/value 404 | error code `NOT_FOUND` |
| PUT /flags/:name/value 200 | updated value in response |
| PUT /flags/:name/value 400 type mismatch | error code `TYPE_MISMATCH` |
| PUT /flags/:name/value 404 | error code `NOT_FOUND` |
| Malformed JSON body | 400 `INVALID_VALUE` |
| Unknown route | 404 |
| Wrong HTTP method | 405 |

### Integration Tests — `go test -tags integration ./...`

Marked `//go:build integration`. Use `testcontainers-go` for ephemeral Postgres and Redis — no external `docker-compose` dependency in CI.

**`internal/adapter/postgres/flag_store_test.go`:**
- Save + FindByName round-trip (values survive DB serialisation)
- Save duplicate name → `ErrAlreadyExists`
- FindByName missing → `ErrNotFound`
- UpdateValue boolean / numeric → correct column updated, `updated_at` bumped
- UpdateValue missing → `ErrNotFound`
- DB constraint: both value columns null → INSERT rejected
- DB constraint: both value columns non-null → INSERT rejected

**`internal/adapter/redis/flag_cache_test.go`:**
- Set + Get boolean round-trip (`"b:true"` encodes/decodes correctly)
- Set + Get numeric round-trip (`"n:30.5"` round-trips correctly)
- Get missing key → `ErrNotFound`
- Delete existing key → subsequent Get returns `ErrNotFound`
- Delete missing key → no error

**`cmd/server/main_test.go`** — full stack, real Postgres + Redis:
- POST then GET — values match
- POST then GET /value — cache populated on create
- PUT /value then GET /value — cache updated
- Force-delete Redis key, GET /value — Postgres fallback works, cache re-populated
- Concurrent PUT /value — last writer wins, no corruption (10 goroutines, final DB and cache agree)

### Shared Test Helpers

`internal/testutil/db.go` and `internal/testutil/redis.go` provide container lifecycle management importable by any `internal/` integration test.

### CI Commands

```
go test ./...                       # unit tests only
go test -tags integration ./...     # unit + integration tests
go test -race ./...                 # data race detection
go test -cover ./...                # coverage report (target ≥85% on internal/service/)
```

---

## 12. Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jackc/pgx/v5` | Postgres driver — superior context support and type safety; no ORM |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/testcontainers/testcontainers-go` | Ephemeral containers for integration tests |
| `github.com/stretchr/testify` | Test assertion helpers (optional) |

Standard library `net/http` with the Go 1.22 mux (path parameter support built-in) is sufficient — no HTTP framework needed.

---

## 13. Composition Root (`cmd/server/main.go`)

```
main()
  ├── Load config (env vars via internal/config)
  ├── Open *pgxpool.Pool  (Postgres connection pool)
  ├── Open *redis.Client  (Redis client)
  ├── Construct postgres.FlagStore  (implements port.FlagStore)
  ├── Construct redis.FlagCache     (implements port.FlagCache)
  ├── Construct service.FlagService (implements port.FlagService)
  ├── Construct http.Handler        (primary adapter, consumes port.FlagService)
  └── http.ListenAndServe(addr, router)
```

All dependency wiring is explicit and happens in one place. No dependency injection framework is used.
