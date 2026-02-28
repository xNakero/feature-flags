package domain

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

