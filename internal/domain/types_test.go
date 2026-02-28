package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func boolPtr(b bool) *bool    { return &b }
func floatPtr(f float64) *float64 { return &f }

func TestFlagValue_IsZero(t *testing.T) {
	t.Run("zero value returns true", func(t *testing.T) {
		var v FlagValue
		assert.True(t, v.IsZero())
	})

	t.Run("boolean value returns false", func(t *testing.T) {
		v := FlagValue{Bool: boolPtr(true)}
		assert.False(t, v.IsZero())
	})

	t.Run("numeric value returns false", func(t *testing.T) {
		v := FlagValue{Numeric: floatPtr(3.14)}
		assert.False(t, v.IsZero())
	})

	t.Run("both fields set returns false", func(t *testing.T) {
		v := FlagValue{Bool: boolPtr(false), Numeric: floatPtr(0)}
		assert.False(t, v.IsZero())
	})
}

func TestFlagValue_Type(t *testing.T) {
	t.Run("boolean value returns FlagTypeBoolean", func(t *testing.T) {
		v := FlagValue{Bool: boolPtr(true)}
		assert.Equal(t, FlagTypeBoolean, v.Type())
	})

	t.Run("numeric value returns FlagTypeNumeric", func(t *testing.T) {
		v := FlagValue{Numeric: floatPtr(42.0)}
		assert.Equal(t, FlagTypeNumeric, v.Type())
	})

	t.Run("both fields set returns FlagTypeBoolean", func(t *testing.T) {
		v := FlagValue{Bool: boolPtr(false), Numeric: floatPtr(1.0)}
		assert.Equal(t, FlagTypeBoolean, v.Type())
	})

	t.Run("zero value panics", func(t *testing.T) {
		var v FlagValue
		assert.Panics(t, func() { v.Type() })
	})
}
