# FEAT-007: Domain Error Sentinels

## Description

Define domain-level error sentinels that represent specific error conditions throughout the application. These enable type-safe error handling using Go's errors.Is() semantics.

## Specifications

### 1. Error Sentinels
- Location: `internal/domain/errors.go`
- Define five exported error variables:
  - `ErrNotFound` - Resource (flag) does not exist
  - `ErrAlreadyExists` - Resource (flag) already exists
  - `ErrTypeMismatch` - Type mismatch (e.g., updating numeric flag with boolean)
  - `ErrInvalidName` - Flag name doesn't meet validation criteria
  - `ErrInvalidValue` - Value doesn't match or violate constraints

### 2. Error Implementation
- Use standard `errors.New()` to create sentinel errors
- Each error should have a descriptive message
- Errors should be package-level variables (not constants)
- Follow Go convention: `var ErrXxx = errors.New("...")`

### 3. Error Semantics
- Each error maps to an HTTP status code per architecture.md:
  - `ErrNotFound` → 404 Not Found
  - `ErrAlreadyExists` → 409 Conflict
  - `ErrTypeMismatch` → 400 Bad Request
  - `ErrInvalidName` → 400 Bad Request
  - `ErrInvalidValue` → 400 Bad Request
- Errors must work with `errors.Is()` for type-safe checking

### 4. Unit Tests
- Location: `internal/domain/errors_test.go`
- Test cases:
  - All errors defined and exported
  - Each error is non-nil
  - `errors.Is()` correctly identifies each error:
    - `errors.Is(ErrNotFound, ErrNotFound)` returns true
    - `errors.Is(ErrNotFound, ErrAlreadyExists)` returns false
  - Wrapped errors: Verify `errors.Is()` works with wrapped errors
    - Create wrapped errors: `fmt.Errorf("context: %w", ErrNotFound)`
    - Verify `errors.Is(wrappedErr, ErrNotFound)` returns true

### 5. Documentation
- Include godoc comment on each error explaining when it's returned
- Document the HTTP status code mapping in comments

## Acceptance Criteria

- [ ] `internal/domain/errors.go` exists
- [ ] All 5 error sentinels defined as exported variables
- [ ] Errors created with `errors.New()` and descriptive messages
- [ ] `internal/domain/errors_test.go` exists with all test cases
- [ ] All errors are non-nil when tested
- [ ] `errors.Is()` semantics work for all error comparisons
- [ ] Wrapped error test passes (errors.Is with wrapped errors)
- [ ] Each error has godoc comment
- [ ] HTTP status code mappings documented in comments
- [ ] `go test ./internal/domain/errors_test.go` passes
- [ ] `go vet ./internal/domain` passes with no warnings
- [ ] All errors are properly exported (capitalized)
- [ ] No error messages leaked in public API (private implementation)
