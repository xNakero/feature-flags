//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xNakero/feature-flags/internal/adapter/postgres"
	"github.com/xNakero/feature-flags/internal/domain"
	"github.com/xNakero/feature-flags/internal/testutil"
)

func newStore(t *testing.T) *postgres.FlagStore {
	t.Helper()
	pool := testutil.NewPostgresPool(t)
	store := postgres.NewFlagStore(pool)
	require.NoError(t, store.CreateSchema(context.Background()))
	return store
}

func TestFlagStore_Create_GetByName_Boolean(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	boolVal := true
	now := time.Now().UTC().Truncate(time.Millisecond)
	flag := domain.Flag{
		Name:        "feature-x",
		Type:        domain.FlagTypeBoolean,
		Description: "a boolean flag",
		Value:       domain.FlagValue{Bool: &boolVal},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	require.NoError(t, store.Create(context.Background(), flag))

	got, err := store.GetByName(context.Background(), "feature-x")
	require.NoError(t, err)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Type, got.Type)
	assert.Equal(t, flag.Description, got.Description)
	assert.Equal(t, flag.Value.Bool, got.Value.Bool)
	assert.Nil(t, got.Value.Numeric)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestFlagStore_Create_GetByName_Numeric(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	numVal := 3.14
	now := time.Now().UTC().Truncate(time.Millisecond)
	flag := domain.Flag{
		Name:        "rate-limit",
		Type:        domain.FlagTypeNumeric,
		Description: "a numeric flag",
		Value:       domain.FlagValue{Numeric: &numVal},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	require.NoError(t, store.Create(context.Background(), flag))

	got, err := store.GetByName(context.Background(), "rate-limit")
	require.NoError(t, err)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Type, got.Type)
	assert.InDelta(t, numVal, *got.Value.Numeric, 1e-9)
	assert.Nil(t, got.Value.Bool)
}

func TestFlagStore_Create_Duplicate(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	boolVal := true
	now := time.Now().UTC()
	flag := domain.Flag{
		Name:      "dup-flag",
		Type:      domain.FlagTypeBoolean,
		Value:     domain.FlagValue{Bool: &boolVal},
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, store.Create(context.Background(), flag))
	err := store.Create(context.Background(), flag)
	require.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestFlagStore_GetByName_NotFound(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	_, err := store.GetByName(context.Background(), "ghost")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestFlagStore_UpdateValue(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	boolVal := true
	now := time.Now().UTC().Truncate(time.Millisecond)
	flag := domain.Flag{
		Name:      "toggle",
		Type:      domain.FlagTypeBoolean,
		Value:     domain.FlagValue{Bool: &boolVal},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, store.Create(context.Background(), flag))

	newBool := false
	updated, err := store.UpdateValue(context.Background(), "toggle", domain.FlagValue{Bool: &newBool})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, &newBool, updated.Value.Bool)
	assert.True(t, !updated.UpdatedAt.Before(now))
}

func TestFlagStore_UpdateValue_NotFound(t *testing.T) {
	t.Parallel()
	store := newStore(t)

	boolVal := true
	_, err := store.UpdateValue(context.Background(), "ghost", domain.FlagValue{Bool: &boolVal})
	require.ErrorIs(t, err, domain.ErrNotFound)
}
