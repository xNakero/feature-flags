# FEAT-008: Flag Name Validation

## Description

Implement validation logic for flag names. This ensures that flag names meet strict naming requirements before being persisted or used in the system.

## Specifications

### 1. Validation Function
- Location: `internal/domain/validation.go`
- Function signature: `func ValidateFlagName(name string) error`
- Validation rules (must satisfy ALL):
  - Not empty (length > 0)
  - Starts with lowercase letter [a-z]
  - Contains only lowercase letters, digits, and hyphens: [a-z0-9-]
  - Maximum length of 63 characters
  - No leading or trailing hyphens

### 2. Error Handling
- Returns `nil` on valid name
- Returns `ErrInvalidName` from `internal/domain/errors.go` on invalid name
- Error message should indicate what validation rule failed
- Consider using `fmt.Errorf("msg: %w", ErrInvalidName)` for detailed messages

### 3. Examples
- Valid names:
  - `feature-x` ✓
  - `enable-auth` ✓
  - `payment-processing-v2` ✓
  - `a` ✓
  - `a1b2c3-d4e5` ✓
  - Exactly 63 characters starting with lowercase letter ✓
- Invalid names:
  - Empty string ✗
  - Starting with digit: `1feature` ✗
  - Starting with hyphen: `-feature` ✗
  - Uppercase letter: `Feature-X` ✗
  - Special characters: `feature@x` ✗
  - 64 characters (too long) ✗
  - Trailing hyphen: `feature-` ✗

### 4. Unit Tests
- Location: `internal/domain/validation_test.go`
- Test cases:
  - **Valid names**:
    - Simple name: `my-flag`
    - With digits: `feature-v2`
    - Single character: `a`
    - Exactly 63 characters
  - **Invalid names**:
    - Empty string
    - Starts with digit: `1feature`
    - Starts with hyphen: `-feature`
    - Uppercase: `MyFeature`
    - Contains uppercase: `my-Feature`
    - Contains special chars: `my@feature`
    - Too long: 64 characters
    - Trailing hyphen: `feature-`
    - Leading hyphen: `-feature`
  - **Edge cases**:
    - Single digit: `1` (invalid)
    - Only digits: `123` (invalid)
    - Only hyphens: `---` (invalid)

## Acceptance Criteria

- [ ] `internal/domain/validation.go` exists
- [ ] `ValidateFlagName(name string) error` function defined and exported
- [ ] Validation checks for all 5 rules (empty, starts with letter, allowed chars, max length, no leading/trailing hyphens)
- [ ] Returns `nil` for valid names
- [ ] Returns `ErrInvalidName` for invalid names
- [ ] `internal/domain/validation_test.go` exists with all test cases
- [ ] All valid name test cases pass
- [ ] All invalid name test cases pass
- [ ] All edge case tests pass
- [ ] Error messages are descriptive (explain which rule failed)
- [ ] `go test ./internal/domain/validation_test.go` passes
- [ ] `go vet ./internal/domain` passes with no warnings
- [ ] Function is properly exported (capitalized)
- [ ] Integration: Function can be imported and used with domain.ErrInvalidName
