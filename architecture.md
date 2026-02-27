# Feature Flags Service — Architecture

## 1. Overview

The Feature Flags Service is a REST API for creating and evaluating feature flags. It supports two value types: **boolean** and **numeric**.

The service uses a two-storage model:

- **PostgreSQL** — authoritative store for all flag data, including metadata such as description and timestamps.
- **Redis** — fast read cache for flag values on the hot path.

Writes always go to Postgres first. Redis is updated after a successful Postgres write. This allows the system to degrade gracefully under a Redis outage: reads fall back to Postgres and the cache self-heals on the next read.

---

## 2. Architecture Style — Hexagonal (Ports & Adapters)

The service is structured in three concentric rings with a strict inward dependency rule: outer rings depend on inner rings, never the reverse.

```
┌───────────────────────────────────────────┐
│  Adapters  (HTTP, Postgres, Redis)        │
│  ┌─────────────────────────────────────┐  │
│  │  Ports  (interfaces)               │  │
│  │  ┌───────────────────────────────┐ │  │
│  │  │  Domain  (entities, errors)   │ │  │
│  │  └───────────────────────────────┘ │  │
│  └─────────────────────────────────────┘  │
└───────────────────────────────────────────┘
```

- **Domain** — core entities and error definitions; zero external dependencies.
- **Ports** — Go interfaces that define the boundaries between rings. Both the service layer and adapters depend on ports, never on each other.
- **Service** — application logic. Knows about domain and ports; knows nothing about HTTP, Postgres, or Redis.
- **Adapters** — concrete implementations of port interfaces. The HTTP adapter drives the service (primary/inbound); the Postgres and Redis adapters are driven by the service (secondary/outbound).
- **`cmd/server/main.go`** — the single composition root where all adapters are wired together and the HTTP server is started.

---

## 3. Package Structure

`_test.go` files live beside the code they test (standard Go convention). Integration tests carry a `//go:build integration` build tag so they are excluded from a plain `go test ./...` run.

```
feature-flags/
├── cmd/server/
│   ├── main.go          # Composition root — wires adapters, starts HTTP server
│   └── main_test.go     # End-to-end HTTP tests  [build tag: integration]
│
├── internal/
│   ├── domain/          # Flag entity, FlagType enum, FlagValue type, error sentinels
│   ├── port/            # Interfaces: FlagService (inbound), FlagStore, FlagCache (outbound)
│   ├── service/         # FlagService implementation (core application logic)
│   ├── adapter/
│   │   ├── http/        # REST handler, router, request/response DTOs, middleware
│   │   ├── postgres/    # FlagStore implementation; SQL migrations
│   │   └── redis/       # FlagCache implementation
│   ├── testutil/        # Shared integration-test helpers (container lifecycle)
│   └── config/          # Environment variable loading
│
└── docker-compose.yml   # Local dev: Postgres + Redis
```

`internal/testutil/` is used instead of a top-level `test/` directory because shared test helpers need to be importable by other `internal/` packages without becoming a public API. A top-level `test/` directory is not idiomatic in Go.

---

## 4. Domain Model

A **Flag** has a name, a type, a description, a value, and created/updated timestamps. The name is the natural primary key — lowercase letters, digits, and hyphens only, starting with a letter, maximum 63 characters.

A flag's **type** is either `boolean` or `numeric` and is immutable after creation. The **value** is typed by the flag's declared type: a boolean flag holds a true/false value; a numeric flag holds a decimal number (sufficient to represent both integers and fractional values like percentage thresholds).

---

## 5. Component Responsibilities

**FlagService** owns all application logic: validating flag names and values, enforcing type consistency between a flag's declared type and incoming values, orchestrating reads and writes across the store and cache, and deciding how to handle cache failures. It has no knowledge of HTTP, SQL, or Redis.

**Postgres adapter (FlagStore)** is responsible for durable persistence: inserting flags, looking them up by name, and updating values. It is the only place where domain types are mapped to and from database columns.

**Redis adapter (FlagCache)** is responsible for the fast read path: storing and retrieving flag values with a type discriminator so the value can be correctly decoded without a second lookup. It translates Redis-specific errors (key not found, connection failure) into the uniform domain error that callers expect.

**HTTP adapter** is responsible for parsing and validating requests, calling the service, serialising responses, and mapping domain errors to appropriate HTTP status codes and JSON error bodies. It contains no business logic.

---

## 6. Data Storage

### PostgreSQL

The `flags` table stores each flag's name (primary key), type, description, timestamps, and two nullable value columns — one for boolean values and one for numeric values. A database-level constraint ensures that exactly one value column is populated, matching the flag's declared type.

Two separate columns are used instead of a single JSON column so that the type constraint can be enforced by the database, reads require no deserialization, and value columns are individually indexable if needed.

### Redis

Keys follow the pattern `flags:value:{name}`. Values are plain strings with a short type prefix so that a single `GET` retrieves both the type discriminator and the value — no additional round-trips, and values remain human-readable via `redis-cli`.

No TTL is set by default. The write-through strategy keeps the cache consistent with Postgres. On a cache miss the service falls back to Postgres and repopulates the cache automatically.

---

## 7. API Catalogue

All responses use `Content-Type: application/json`. Errors share a common envelope with a machine-readable `code` field and a human-readable `message`.

| Method | Path                  | Description                              | Success |
|--------|-----------------------|------------------------------------------|---------|
| POST   | /flags                | Create a new flag                        | 201     |
| GET    | /flags/:name          | Full flag detail; always reads Postgres  | 200     |
| GET    | /flags/:name/value    | Flag value; Redis-first, Postgres fallback | 200   |
| PUT    | /flags/:name/value    | Update value; write-through to both stores | 200  |

---

## 8. Write-Through Flow

```
Client  →  PUT /flags/:name/value
                     │
            ┌────────▼────────┐
            │  HTTP Handler    │  Parse and validate request
            └────────┬────────┘
                     │
            ┌────────▼────────┐
            │  FlagService     │  Load flag from Postgres
            │                  │  → not found? return 404
            │                  │
            │                  │  Validate type match
            │                  │  → mismatch? return 400
            │                  │
            │                  │  Write to Postgres  ← HARD FAIL
            │                  │  → error? return 5xx; do not touch Redis
            │                  │
            │                  │  Write to Redis     ← SOFT FAIL
            │                  │  → error? log WARN; return 200 anyway
            └────────┬────────┘
                     │
            ┌────────▼────────┐
            │  HTTP Handler    │  Serialise updated flag → 200 OK
            └─────────────────┘
```

**Failure asymmetry:** a Postgres write failure means the value was not persisted — the request fails. A Redis write failure means the value is safely in Postgres but the cache is stale; the next read will miss Redis, fall back to Postgres, get the correct value, and repopulate the cache. No data is lost and no write error is surfaced to the caller.

---

## 9. Read Flows

### GET /flags/:name — Full Detail

Always reads from Postgres. Redis is not consulted because the full response includes metadata (description, timestamps) that is not cached.

### GET /flags/:name/value — Cached Value

```
FlagService.GetFlagValue
  │
  ├─ Redis GET
  │    ├─ HIT  → decode value → return immediately (no Postgres query)
  │    └─ MISS or Redis unavailable
  │         └─ Postgres SELECT
  │              ├─ found   → populate Redis cache (soft-fail) → return value
  │              └─ missing → 404
```

---

## 10. Error Handling

Domain errors map to specific HTTP responses:

| Condition                                      | HTTP | Error code       |
|------------------------------------------------|------|------------------|
| Flag does not exist                            | 404  | `NOT_FOUND`      |
| Creating a flag whose name is already taken    | 409  | `ALREADY_EXISTS` |
| Value type contradicts the flag's declared type | 400 | `TYPE_MISMATCH`  |
| Name is empty, too long, or contains illegal characters | 400 | `INVALID_NAME` |
| Value is missing, null, or wrong JSON kind     | 400  | `INVALID_VALUE`  |

Infrastructure errors are handled separately: a Postgres failure returns 503; an unknown error returns 500. Redis failures on the write path are suppressed (logged at WARN level); Redis failures on the read path trigger a transparent Postgres fallback.

---

## 11. Testing Strategy

**Unit tests** (`go test ./...`) use hand-written fakes for the port interfaces — no external processes needed. They cover all service logic branches (cache hit/miss, Redis soft-fail, type mismatch, not found) and all HTTP handler paths (status codes, JSON shapes, error codes).

**Integration tests** (`go test -tags integration ./...`) use `testcontainers-go` to spin up real Postgres and Redis containers. The Postgres adapter tests verify DB round-trips and constraint enforcement; the Redis adapter tests verify encoding/decoding and miss handling.

**End-to-end tests** live in `cmd/server/` and exercise the full stack including write-through consistency, cache fallback and repopulation, and concurrent updates.

CI runs `go test -race ./...` for data race detection and `go test -cover ./...` for coverage (target ≥ 85% on the service layer).

---

## 12. Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jackc/pgx/v5` | Postgres driver — strong context support and type safety; no ORM |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/testcontainers/testcontainers-go` | Ephemeral Postgres and Redis containers for integration tests |
| `github.com/stretchr/testify` | Test assertion helpers |
