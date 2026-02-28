package port

import (
	"context"
	"time"
)

// FlagValue is the port-level representation of a flag's value.
// Exactly one of Bool or Numeric should be non-nil at a time.
type FlagValue struct {
	Bool    *bool
	Numeric *float64
}

type CreateFlagRequest struct {
	// Name is the desired flag name. Must contain only lowercase letters, digits,
	// and hyphens, start with a letter, and be at most 63 characters long.
	Name string
	// Type is the flag's value type. Accepted values: "boolean", "numeric".
	Type        string
	Description string
	Value       FlagValue
}

type UpdateFlagValueRequest struct {
	Value FlagValue
}

// FlagResponse is the DTO returned by service methods that operate on a full flag.
type FlagResponse struct {
	Name        string
	Type        string
	Description string
	Value       FlagValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FlagValueResponse is the DTO returned by GetFlagValue.
type FlagValueResponse struct {
	Value FlagValue
}

// FlagService is the inbound port through which HTTP handlers interact with the application's core logic.
type FlagService interface {
	CreateFlag(ctx context.Context, req CreateFlagRequest) (*FlagResponse, error)
	GetFlag(ctx context.Context, name string) (*FlagResponse, error)
	// GetFlagValue retrieves only the current value of the flag, not the full record.
	GetFlagValue(ctx context.Context, name string) (*FlagValueResponse, error)
	UpdateFlagValue(ctx context.Context, name string, req UpdateFlagValueRequest) (*FlagResponse, error)
}
