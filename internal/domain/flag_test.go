package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xNakero/feature-flags/internal/domain"
)

func TestFlagEntityConstruction(t *testing.T) {
	boolVal := true
	now := time.Now()
	f := domain.Flag{
		Name:        "my-flag",
		Type:        domain.FlagTypeBoolean,
		Description: "a test flag",
		Value:       domain.FlagValue{Bool: &boolVal},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "my-flag", f.Name)
	assert.Equal(t, domain.FlagTypeBoolean, f.Type)
	assert.Equal(t, "a test flag", f.Description)
	assert.Equal(t, &boolVal, f.Value.Bool)
	assert.Equal(t, now, f.CreatedAt)
	assert.Equal(t, now, f.UpdatedAt)
}

func TestFlagFieldAccess(t *testing.T) {
	num := 3.14
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	f := domain.Flag{
		Name:        "numeric-flag",
		Type:        domain.FlagTypeNumeric,
		Description: "holds a numeric value",
		Value:       domain.FlagValue{Numeric: &num},
		CreatedAt:   created,
		UpdatedAt:   updated,
	}

	assert.Equal(t, "numeric-flag", f.Name)
	assert.Equal(t, domain.FlagTypeNumeric, f.Type)
	assert.Equal(t, "holds a numeric value", f.Description)
	assert.Equal(t, &num, f.Value.Numeric)
	assert.Nil(t, f.Value.Bool)
}

func TestFlagTimestampHandling(t *testing.T) {
	created := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 3, 20, 12, 30, 0, 0, time.UTC)

	f := domain.Flag{
		CreatedAt: created,
		UpdatedAt: updated,
	}

	assert.True(t, f.CreatedAt.Equal(created))
	assert.True(t, f.UpdatedAt.Equal(updated))
	assert.True(t, f.UpdatedAt.After(f.CreatedAt))
}

func TestFlagTypeRepresentation(t *testing.T) {
	boolVal := false
	f := domain.Flag{
		Type:  domain.FlagTypeBoolean,
		Value: domain.FlagValue{Bool: &boolVal},
	}

	assert.Equal(t, domain.FlagType("boolean"), f.Type)
	assert.IsType(t, domain.FlagType(""), f.Type)
}

func TestBooleanFlag(t *testing.T) {
	v := true
	f := domain.Flag{
		Name:  "feature-enabled",
		Type:  domain.FlagTypeBoolean,
		Value: domain.FlagValue{Bool: &v},
	}

	assert.Equal(t, domain.FlagTypeBoolean, f.Type)
	assert.NotNil(t, f.Value.Bool)
	assert.Nil(t, f.Value.Numeric)
	assert.True(t, *f.Value.Bool)
}

func TestNumericFlag(t *testing.T) {
	v := 42.5
	f := domain.Flag{
		Name:  "rate-limit",
		Type:  domain.FlagTypeNumeric,
		Value: domain.FlagValue{Numeric: &v},
	}

	assert.Equal(t, domain.FlagTypeNumeric, f.Type)
	assert.NotNil(t, f.Value.Numeric)
	assert.Nil(t, f.Value.Bool)
	assert.InDelta(t, 42.5, *f.Value.Numeric, 1e-9)
}
