# FEAT-004: Inbound Port Interface and Service Request/Response Types

## Description

Define the inbound port interface and data transfer objects (DTOs) for the business logic service. These establish the contract for controllers/handlers to communicate with the domain service layer.

## Specifications

### 1. Request DTOs
- Location: `internal/port/service.go`
- `CreateFlagRequest` struct with fields:
  - `Name string` - Flag name (will be validated)
  - `Type string` - Flag type (boolean or numeric)
  - `Description string` - Human-readable description
  - `Value interface{}` - Initial flag value (type depends on Type field)
- `UpdateFlagValueRequest` struct with fields:
  - `Value interface{}` - New flag value to set

### 2. FlagService Interface
- Location: `internal/port/service.go`
- Methods:
  - `CreateFlag(ctx context.Context, req CreateFlagRequest) (*domain.Flag, error)`
    - Creates new flag with provided request data
    - Returns created Flag with timestamps populated
    - Returns `ErrInvalidName` if flag name doesn't meet validation rules
    - Returns `ErrAlreadyExists` if flag already exists
    - Returns `ErrInvalidValue` if value doesn't match type
  - `GetFlag(ctx context.Context, name string) (*domain.Flag, error)`
    - Retrieves complete flag by name
    - Returns `ErrNotFound` if flag doesn't exist
  - `GetFlagValue(ctx context.Context, name string) (*domain.FlagValue, error)`
    - Retrieves only the flag value (not full Flag)
    - Returns `ErrNotFound` if flag doesn't exist
  - `UpdateFlagValue(ctx context.Context, name string, req UpdateFlagValueRequest) (*domain.Flag, error)`
    - Updates value of existing flag
    - Returns updated Flag
    - Returns `ErrNotFound` if flag doesn't exist
    - Returns `ErrTypeMismatch` if new value type doesn't match flag type
    - Returns `ErrInvalidValue` if value validation fails

### 3. Interface Design
- All methods accept `context.Context` as first parameter
- Interface is unidirectional for inbound calls (controller → service)
- Interface uses domain types (domain.Flag, domain.FlagValue) as returns
- Request DTOs are simple data carriers without behavior

### 4. Verification
- Ensure method signatures map 1:1 to API catalogue endpoints (see architecture.md)
- Verify request/response types align with REST API requirements

## Acceptance Criteria

- [ ] `internal/port/service.go` contains both DTOs and interface
- [ ] `CreateFlagRequest` struct with 4 fields exists
- [ ] `UpdateFlagValueRequest` struct with 1 field exists
- [ ] `FlagService` interface has exactly 4 methods
- [ ] All methods use `context.Context` as first parameter
- [ ] All methods return error as last return value
- [ ] `CreateFlag` returns `(*domain.Flag, error)`
- [ ] `GetFlag` returns `(*domain.Flag, error)`
- [ ] `GetFlagValue` returns `(*domain.FlagValue, error)`
- [ ] `UpdateFlagValue` returns `(*domain.Flag, error)`
- [ ] Godoc comments document all DTOs and interface methods
- [ ] Error contracts documented (which errors returned when)
- [ ] Methods map 1:1 to API catalogue endpoints
- [ ] `go build ./internal/port` succeeds
- [ ] No compilation errors with domain package imports
