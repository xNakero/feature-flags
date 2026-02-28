package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xNakero/feature-flags/internal/domain"
)

func TestValidateFlagName(t *testing.T) {
	t.Parallel()

	valid63 := strings.Repeat("a", 63)
	invalid64 := strings.Repeat("a", 64)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{name: "simple hyphenated", input: "my-flag", wantErr: false},
		{name: "with digits", input: "feature-v2", wantErr: false},
		{name: "single character", input: "a", wantErr: false},
		{name: "exactly 63 chars", input: valid63, wantErr: false},
		{name: "alphanumeric mixed", input: "a1b2c3-d4e5", wantErr: false},

		// Invalid names
		{name: "empty string", input: "", wantErr: true},
		{name: "starts with digit", input: "1feature", wantErr: true},
		{name: "starts with hyphen", input: "-feature", wantErr: true},
		{name: "uppercase first", input: "MyFeature", wantErr: true},
		{name: "uppercase mid", input: "my-Feature", wantErr: true},
		{name: "special char", input: "my@feature", wantErr: true},
		{name: "64 chars", input: invalid64, wantErr: true},
		{name: "trailing hyphen", input: "feature-", wantErr: true},
		{name: "leading hyphen", input: "-feature", wantErr: true},

		// Edge cases
		{name: "single digit", input: "1", wantErr: true},
		{name: "only digits", input: "123", wantErr: true},
		{name: "only hyphens", input: "---", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateFlagName(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidName)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFlagValue(t *testing.T) {
	t.Parallel()

	boolVal := true
	numVal := 3.14

	tests := []struct {
		name      string
		flagType  domain.FlagType
		flagValue domain.FlagValue
		wantErr   error
	}{
		{
			name:      "boolean flag with bool value",
			flagType:  domain.FlagTypeBoolean,
			flagValue: domain.FlagValue{Bool: &boolVal},
		},
		{
			name:      "numeric flag with numeric value",
			flagType:  domain.FlagTypeNumeric,
			flagValue: domain.FlagValue{Numeric: &numVal},
		},
		{
			name:      "boolean flag with numeric value",
			flagType:  domain.FlagTypeBoolean,
			flagValue: domain.FlagValue{Numeric: &numVal},
			wantErr:   domain.ErrTypeMismatch,
		},
		{
			name:      "numeric flag with bool value",
			flagType:  domain.FlagTypeNumeric,
			flagValue: domain.FlagValue{Bool: &boolVal},
			wantErr:   domain.ErrTypeMismatch,
		},
		{
			name:      "boolean flag with no value",
			flagType:  domain.FlagTypeBoolean,
			flagValue: domain.FlagValue{},
			wantErr:   domain.ErrTypeMismatch,
		},
		{
			name:      "numeric flag with no value",
			flagType:  domain.FlagTypeNumeric,
			flagValue: domain.FlagValue{},
			wantErr:   domain.ErrTypeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateFlagValue(tt.flagType, tt.flagValue)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
