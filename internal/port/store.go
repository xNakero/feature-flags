// Package port defines the inbound and outbound port interfaces for the
// feature flags service, following hexagonal (ports and adapters) architecture.
package port

import (
	"context"

	"github.com/xNakero/feature-flags/internal/domain"
)

// FlagStore is the outbound port for persisting and retrieving feature flags.
// Concrete implementations (e.g. PostgreSQL) must satisfy this interface.
//
// All methods accept a context.Context as the first argument to support
// cancellation and deadline propagation.
//
// Error contracts:
//   - domain.ErrNotFound is returned when a requested flag does not exist.
//   - domain.ErrAlreadyExists is returned when attempting to create a flag
//     whose name is already in use.
//   - domain.ErrTypeMismatch is returned when the value type supplied to
//     UpdateValue does not match the flag's declared type.
type FlagStore interface {
	// Create persists a new flag to storage.
	// Returns domain.ErrAlreadyExists if a flag with the same name already exists.
	Create(ctx context.Context, flag domain.Flag) error

	// GetByName retrieves a flag by its unique name.
	// Returns the full Flag struct with all fields populated.
	// Returns domain.ErrNotFound if no flag with that name exists.
	GetByName(ctx context.Context, name string) (*domain.Flag, error)

	// UpdateValue updates only the value field of an existing flag.
	// Returns the updated Flag struct on success.
	// Returns domain.ErrNotFound if no flag with that name exists.
	// Returns domain.ErrTypeMismatch if the new value's type does not match
	// the flag's declared type.
	UpdateValue(ctx context.Context, name string, value domain.FlagValue) (*domain.Flag, error)
}
