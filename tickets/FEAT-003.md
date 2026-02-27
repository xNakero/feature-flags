# FEAT-003: Outbound Port Interfaces (FlagStore and FlagCache)

## Description

Define the outbound port interfaces for data persistence and caching layers. These interfaces establish the contracts that concrete adapters (PostgreSQL, Redis) will implement.

## Specifications

### 1. FlagStore Interface
- Location: `internal/port/store.go`
- Methods:
  - `Create(ctx context.Context, flag domain.Flag) error`
    - Persists a new flag to storage
    - Must return `ErrAlreadyExists` if flag with same name already exists
  - `GetByName(ctx context.Context, name string) (*domain.Flag, error)`
    - Retrieves flag by name
    - Returns `ErrNotFound` if flag does not exist
    - Returns full Flag struct with all fields populated
  - `UpdateValue(ctx context.Context, name string, value domain.FlagValue) (*domain.Flag, error)`
    - Updates only the value of an existing flag
    - Returns updated Flag struct
    - Returns `ErrNotFound` if flag does not exist
    - Returns `ErrTypeMismatch` if new value type doesn't match flag's type

### 2. FlagCache Interface
- Location: `internal/port/cache.go`
- Methods:
  - `Get(ctx context.Context, name string) (*domain.FlagValue, error)`
    - Retrieves cached flag value by name
    - Returns `ErrNotFound` if key not in cache (cache miss)
    - Returns only the FlagValue, not full Flag
  - `Set(ctx context.Context, name string, value domain.FlagValue) error`
    - Stores flag value in cache
    - Allows overwriting existing cached values
  - `Delete(ctx context.Context, name string) error`
    - Removes cached value by name
    - Should not error if key doesn't exist (idempotent)

### 3. Documentation
- Include godoc comments on each interface
- Document error contracts in comments
- Explain when each error is returned
- Include context.Context usage pattern

## Acceptance Criteria

- [ ] `internal/port/store.go` exists with `FlagStore` interface
- [ ] `FlagStore` has exactly 3 methods: Create, GetByName, UpdateValue
- [ ] `internal/port/cache.go` exists with `FlagCache` interface
- [ ] `FlagCache` has exactly 3 methods: Get, Set, Delete
- [ ] All methods accept `context.Context` as first parameter
- [ ] Error contracts documented in godoc comments
- [ ] `FlagStore` methods reference domain.Flag and domain.FlagValue types
- [ ] `FlagCache` methods reference domain.FlagValue type only
- [ ] `go build ./internal/port` succeeds
- [ ] No circular dependencies with domain package
- [ ] Interfaces are exported (capitalized names)
