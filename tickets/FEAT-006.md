# FEAT-006: Flag Entity

## Description

Implement the core Flag domain entity that represents a feature flag with all its attributes. This is the primary domain model around which the service operates.

## Specifications

### 1. Flag Entity
- Location: `internal/domain/flag.go`
- Struct definition with fields:
  - `Name string` - The flag's unique name/identifier
  - `Type FlagType` - Type of flag (boolean or numeric)
  - `Description string` - Human-readable description
  - `Value FlagValue` - Current flag value
  - `CreatedAt time.Time` - Creation timestamp
  - `UpdatedAt time.Time` - Last update timestamp
- All fields should be exported (capitalized)
- Flags are immutable domain objects (no setters)

### 2. Unit Tests
- Location: `internal/domain/flag_test.go`
- Test cases:
  - Entity construction: Create Flag with all fields
  - Field access: Verify all fields accessible and correctly set
  - Timestamp handling: Verify CreatedAt and UpdatedAt are properly set
  - Type representation: Ensure Type field is FlagType correctly
  - Boolean flag: Create flag with boolean value
  - Numeric flag: Create flag with numeric value

### 3. Design Notes
- Flag entity is a simple data holder (no business logic at this level)
- Validation is handled separately (FEAT-008)
- No methods beyond field getters initially
- Entity should work with domain.FlagValue and domain.FlagType from FEAT-005

## Acceptance Criteria

- [ ] `internal/domain/flag.go` exists
- [ ] `Flag` struct with 6 fields defined: Name, Type, Description, Value, CreatedAt, UpdatedAt
- [ ] All fields are exported (capitalized)
- [ ] Field types correct: Name (string), Type (FlagType), Description (string), Value (FlagValue), timestamps (time.Time)
- [ ] `internal/domain/flag_test.go` exists with all test cases
- [ ] Entity construction test passes
- [ ] Field access tests pass
- [ ] Timestamp handling tests pass
- [ ] Boolean flag entity test passes
- [ ] Numeric flag entity test passes
- [ ] `go test ./internal/domain/flag_test.go` passes
- [ ] `go vet ./internal/domain` passes with no warnings
- [ ] No unexported fields or methods
- [ ] Struct is well-documented with godoc comments
