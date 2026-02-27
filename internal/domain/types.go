package domain

import "time"

type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeNumeric FlagType = "numeric"
)

// FlagValue holds the current value of a feature flag.
// Exactly one of Bool or Numeric should be non-nil at a time.
type FlagValue struct {
	Bool    *bool
	Numeric *float64
}

type Flag struct {
	Name        string
	Type        FlagType
	Description string
	Value       FlagValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
