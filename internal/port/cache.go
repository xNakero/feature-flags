package port

import (
	"context"

	"github.com/xNakero/feature-flags/internal/domain"
)

// FlagCache is the outbound port for caching feature flag values.
// Concrete implementations (e.g. Redis) must satisfy this interface.
//
// The cache stores only the FlagValue, not the full Flag entity, to keep
// cached entries lightweight.
//
// All methods accept a context.Context as the first argument to support
// cancellation and deadline propagation.
//
// Error contracts:
//   - domain.ErrNotFound is returned by Get when the requested key is not
//     present in the cache (cache miss). Callers should fall back to the
//     FlagStore on this error.
//   - Set and Delete do not return domain-specific errors; Delete is
//     idempotent and does not error when the key is absent.
type FlagCache interface {
	// Get retrieves the cached FlagValue for the given flag name.
	// Returns domain.ErrNotFound on a cache miss (key not present).
	Get(ctx context.Context, name string) (*domain.FlagValue, error)

	// Set stores the FlagValue for the given flag name in the cache.
	// Overwrites any existing cached value for that name.
	Set(ctx context.Context, name string, value domain.FlagValue) error

	// Delete removes the cached value for the given flag name.
	// This operation is idempotent: it does not return an error if the key
	// does not exist.
	Delete(ctx context.Context, name string) error
}
