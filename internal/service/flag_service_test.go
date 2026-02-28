package service_test

import (
	"context"
	"errors"
	"testing"

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
		wantNoErr   bool
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
			wantNoErr:   true,
			wantName:    "my-flag",
			wantType:    "boolean",
			wantDesc:    "a boolean flag",
			wantBool:    &boolVal,
			wantNumeric: nil,
		},
		{
			name: "valid numeric flag",
			req: port.CreateFlagRequest{
				Name:        "rate-limit",
				Type:        "numeric",
				Description: "a numeric flag",
				Value:       port.FlagValue{Numeric: &numVal},
			},
			wantNoErr:   true,
			wantName:    "rate-limit",
			wantType:    "numeric",
			wantDesc:    "a numeric flag",
			wantBool:    nil,
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

			// Pre-seed a flag so duplicate detection works.
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
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected errors.Is %v, got %v", tt.wantErr, err)
				}
				if resp != nil {
					t.Fatal("expected nil response on error")
				}
				return
			}

			if !tt.wantNoErr {
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.Name != tt.wantName {
				t.Errorf("Name: got %q, want %q", resp.Name, tt.wantName)
			}
			if resp.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", resp.Type, tt.wantType)
			}
			if resp.Description != tt.wantDesc {
				t.Errorf("Description: got %q, want %q", resp.Description, tt.wantDesc)
			}
			if tt.wantBool != nil && (resp.Value.Bool == nil || *resp.Value.Bool != *tt.wantBool) {
				t.Errorf("Value.Bool: got %v, want %v", resp.Value.Bool, tt.wantBool)
			}
			if tt.wantNumeric != nil && (resp.Value.Numeric == nil || *resp.Value.Numeric != *tt.wantNumeric) {
				t.Errorf("Value.Numeric: got %v, want %v", resp.Value.Numeric, tt.wantNumeric)
			}
			if resp.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if resp.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should not be zero")
			}
		})
	}
}
