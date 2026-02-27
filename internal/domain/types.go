package domain

import "time"

// FlagType represents the type of a feature flag value.
type FlagType string

const (
	// FlagTypeBoolean indicates a boolean feature flag.
	FlagTypeBoolean FlagType = "boolean"
	// FlagTypeNumeric indicates a numeric (float64) feature flag.
	FlagTypeNumeric FlagType = "numeric"
)

// FlagValue holds the current value of a feature flag.
// Exactly one of Bool or Numeric should be non-nil at a time.
type FlagValue struct {
	Bool    *bool
	Numeric *float64
}

// Type returns the FlagType corresponding to the non-nil field.
// It panics if both fields are nil (invalid state).
func (v FlagValue) Type() FlagType {
	if v.Bool != nil {
		return FlagTypeBoolean
	}
	return FlagTypeNumeric
}

// IsZero reports whether the FlagValue is uninitialized (both fields nil).
func (v FlagValue) IsZero() bool {
	return v.Bool == nil && v.Numeric == nil
}

// Flag represents a feature flag domain entity.
type Flag struct {
	Name        string
	Type        FlagType
	Description string
	Value       FlagValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
