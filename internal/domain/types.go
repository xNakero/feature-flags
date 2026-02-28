package domain

import "time"

type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeNumeric FlagType = "numeric"
)

type FlagValue struct {
	Bool    *bool
	Numeric *float64
}

func (v FlagValue) Type() FlagType {
	if v.Bool != nil {
		return FlagTypeBoolean
	}
	if v.Numeric != nil {
		return FlagTypeNumeric
	}
	panic("FlagValue.Type called on zero value")
}

func (v FlagValue) IsZero() bool {
	return v.Bool == nil && v.Numeric == nil
}

type Flag struct {
	Name        string
	Type        FlagType
	Description string
	Value       FlagValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
