package domain

import "time"

// Flag represents a feature toggle with its associated metadata and current value.
type Flag struct {
	Name        string
	Type        FlagType
	Description string
	Value       FlagValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
