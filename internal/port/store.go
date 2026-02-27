// Package port defines the inbound and outbound port interfaces for the
// feature flags service, following hexagonal (ports and adapters) architecture.
package port

import (
	"context"

	"github.com/xNakero/feature-flags/internal/domain"
)

// FlagStore is the outbound port for persisting and retrieving feature flags.
// Concrete implementations (e.g. PostgreSQL) must satisfy this interface.
type FlagStore interface {
	Create(ctx context.Context, flag domain.Flag) error
	GetByName(ctx context.Context, name string) (*domain.Flag, error)
	UpdateValue(ctx context.Context, name string, flagValue domain.FlagValue) (*domain.Flag, error)
}
