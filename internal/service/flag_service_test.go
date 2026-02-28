package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xNakero/feature-flags/internal/domain"
	"github.com/xNakero/feature-flags/internal/port"
	"github.com/xNakero/feature-flags/internal/service"
)

// fakeFlagStore is an in-memory hand-written fake implementing port.FlagStore.
type fakeFlagStore struct {
	flags map[string]domain.Flag
}

func newFakeFlagStore() *fakeFlagStore {
	return &fakeFlagStore{flags: make(map[string]domain.Flag)}
}

func (f *fakeFlagStore) Create(_ context.Context, flag domain.Flag) error {
	if _, exists := f.flags[flag.Name]; exists {
		return domain.ErrAlreadyExists
	}
	f.flags[flag.Name] = flag
	return nil
}

func (f *fakeFlagStore) GetByName(_ context.Context, name string) (*domain.Flag, error) {
	flag, ok := f.flags[name]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &flag, nil
}

func (f *fakeFlagStore) UpdateValue(_ context.Context, name string, flagValue domain.FlagValue) (*domain.Flag, error) {
	flag, ok := f.flags[name]
	if !ok {
		return nil, domain.ErrNotFound
	}
	flag.Value = flagValue
	f.flags[name] = flag
	return &flag, nil
}

func TestService_CreateFlag(t *testing.T) {
	t.Parallel()

	boolVal := true
	numVal := 42.0

	tests := []struct {
		name        string
		req         port.CreateFlagRequest
		wantErr     error
		wantName    string
		wantType    string
		wantDesc    string
		wantBool    *bool
		wantNumeric *float64
	}{
		{
			name: "valid boolean flag",
			req: port.CreateFlagRequest{
				Name:        "my-flag",
				Type:        "boolean",
				Description: "a boolean flag",
				Value:       port.FlagValue{Bool: &boolVal},
			},
			wantName: "my-flag",
			wantType: "boolean",
			wantDesc: "a boolean flag",
			wantBool: &boolVal,
		},
		{
			name: "valid numeric flag",
			req: port.CreateFlagRequest{
				Name:        "rate-limit",
				Type:        "numeric",
				Description: "a numeric flag",
				Value:       port.FlagValue{Numeric: &numVal},
			},
			wantName:    "rate-limit",
			wantType:    "numeric",
			wantDesc:    "a numeric flag",
			wantNumeric: &numVal,
		},
		{
			name: "empty name",
			req: port.CreateFlagRequest{
				Name:  "",
				Type:  "boolean",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "name starts with digit",
			req: port.CreateFlagRequest{
				Name:  "1-feature",
				Type:  "boolean",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "name has uppercase",
			req: port.CreateFlagRequest{
				Name:  "MyFlag",
				Type:  "boolean",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "name has trailing hyphen",
			req: port.CreateFlagRequest{
				Name:  "my-flag-",
				Type:  "boolean",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrInvalidName,
		},
		{
			name: "unknown flag type",
			req: port.CreateFlagRequest{
				Name:  "my-flag",
				Type:  "string",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrInvalidValue,
		},
		{
			name: "type mismatch: boolean type with numeric value",
			req: port.CreateFlagRequest{
				Name:  "my-flag",
				Type:  "boolean",
				Value: port.FlagValue{Numeric: &numVal},
			},
			wantErr: domain.ErrTypeMismatch,
		},
		{
			name: "type mismatch: numeric type with bool value",
			req: port.CreateFlagRequest{
				Name:  "my-flag",
				Type:  "numeric",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrTypeMismatch,
		},
		{
			name: "duplicate flag name",
			req: port.CreateFlagRequest{
				Name:  "existing-flag",
				Type:  "boolean",
				Value: port.FlagValue{Bool: &boolVal},
			},
			wantErr: domain.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newFakeFlagStore()

			if tt.wantErr == domain.ErrAlreadyExists {
				_ = store.Create(context.Background(), domain.Flag{
					Name:  tt.req.Name,
					Type:  domain.FlagTypeBoolean,
					Value: domain.FlagValue{Bool: &boolVal},
				})
			}

			svc := service.New(store)
			resp, err := svc.CreateFlag(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantName, resp.Name)
			assert.Equal(t, tt.wantType, resp.Type)
			assert.Equal(t, tt.wantDesc, resp.Description)
			if tt.wantBool != nil {
				assert.Equal(t, tt.wantBool, resp.Value.Bool)
			}
			if tt.wantNumeric != nil {
				assert.Equal(t, tt.wantNumeric, resp.Value.Numeric)
			}
			assert.False(t, resp.CreatedAt.IsZero())
			assert.False(t, resp.UpdatedAt.IsZero())
		})
	}
}
