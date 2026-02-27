# FEAT-005: Domain Types (FlagType and FlagValue)

## Description

Implement the domain value types for representing flag types and their values. These are the foundational types used throughout the service for type-safe flag value handling.

## Specifications

### 1. FlagType Type
- Location: `internal/domain/types.go`
- Define as: `type FlagType string`
- Constants:
  - `FlagTypeBoolean FlagType = "boolean"` - For boolean flags
  - `FlagTypeNumeric FlagType = "numeric"` - For numeric flags (decimal/float)

### 2. FlagValue Type
- Location: `internal/domain/types.go`
- Define struct:
  ```go
  type FlagValue struct {
    Bool    *bool
    Numeric *decimal.Decimal  // or *float64 if decimal not suitable
  }
  ```
- Only one field should be non-nil at a time (type-safe)
- Use pointers to allow nil representation of unset values

### 3. Methods on FlagValue
- `Type() FlagType` method:
  - Returns `FlagTypeBoolean` if Bool is non-nil
  - Returns `FlagTypeNumeric` if Numeric is non-nil
  - Panics or returns error if both are nil (invalid state)
- `IsZero() bool` helper:
  - Returns true if both Bool and Numeric are nil
  - Represents uninitialized/empty value

### 4. Unit Tests
- Location: `internal/domain/types_test.go`
- Test cases:
  - Boolean constructor/creation: FlagValue with Bool set
  - Numeric constructor/creation: FlagValue with Numeric set
  - `Type()` dispatch logic:
    - Boolean value returns FlagTypeBoolean
    - Numeric value returns FlagTypeNumeric
  - `IsZero()` handling:
    - Zero value (both nil) returns true
    - Non-zero values return false
  - Edge cases: both fields set (should handle gracefully)

## Acceptance Criteria

- [ ] `internal/domain/types.go` exists
- [ ] `FlagType` string type defined with 2 constants: Boolean, Numeric
- [ ] `FlagValue` struct defined with Bool and Numeric pointer fields
- [ ] `Type()` method returns correct FlagType based on field state
- [ ] `IsZero()` method returns true only when both fields are nil
- [ ] `internal/domain/types_test.go` exists with all test cases
- [ ] All boolean value tests pass
- [ ] All numeric value tests pass
- [ ] All type dispatch tests pass
- [ ] All zero-value tests pass
- [ ] `go test ./internal/domain/types_test.go` passes with 100% coverage
- [ ] `go vet ./internal/domain` passes with no warnings
- [ ] No panic in normal usage (handled gracefully)
- [ ] Types are properly exported (capitalized)
