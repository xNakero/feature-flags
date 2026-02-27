package port

import (
	"context"

	"github.com/xNakero/feature-flags/internal/domain"
)

// FlagCache is the outbound port for caching feature flag values.
// Concrete implementations (e.g. Redis) must satisfy this interface.
type FlagCache interface {
	Get(ctx context.Context, name string) (*domain.FlagValue, error)
	Set(ctx context.Context, name string, flagValue domain.FlagValue) error
	Delete(ctx context.Context, name string) error
}
