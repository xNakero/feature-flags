package port

import (
	"context"
	"time"
)

// The Value field must match the declared Type: a boolean value for Type "boolean",
// or a numeric value for Type "numeric".
type CreateFlagRequest struct {
	// Name is the desired flag name. Must contain only lowercase letters, digits,
	// and hyphens, start with a letter, and be at most 63 characters long.
	Name string
	// Type is the flag's value type. Accepted values: "boolean", "numeric".
	Type string
	// Description is a human-readable explanation of the flag's purpose.
	Description string
	// Value is the initial value for the flag. Its kind must match Type.
	Value interface{}
}

type UpdateFlagValueRequest struct {
	Value interface{}
}

// FlagResponse is the DTO returned by service methods that operate on a full flag.
type FlagResponse struct {
	Name        string
	Type        string
	Description string
	Value       interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FlagValueResponse is the DTO returned by GetFlagValue.
type FlagValueResponse struct {
	Value interface{}
}

// FlagService is the inbound port through which HTTP handlers interact with the application's core logic.
type FlagService interface {
	CreateFlag(ctx context.Context, req CreateFlagRequest) (*FlagResponse, error)
	GetFlag(ctx context.Context, name string) (*FlagResponse, error)
	// GetFlagValue retrieves only the current value of the flag, not the full record.
	GetFlagValue(ctx context.Context, name string) (*FlagValueResponse, error)
	UpdateFlagValue(ctx context.Context, name string, req UpdateFlagValueRequest) (*FlagResponse, error)
}
