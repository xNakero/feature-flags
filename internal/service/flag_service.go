package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xNakero/feature-flags/internal/domain"
	"github.com/xNakero/feature-flags/internal/port"
)

type Service struct {
	store port.FlagStore
}

func New(store port.FlagStore) *Service {
	return &Service{store: store}
}

func (s *Service) CreateFlag(ctx context.Context, req port.CreateFlagRequest) (*port.FlagResponse, error) {
	if err := domain.ValidateFlagName(req.Name); err != nil {
		return nil, err
	}

	flagType, err := parseFlagType(req.Type)
	if err != nil {
		return nil, err
	}

	domainValue := domain.FlagValue{Bool: req.Value.Bool, Numeric: req.Value.Numeric}
	if err := domain.ValidateFlagValue(flagType, domainValue); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	flag := domain.Flag{
		Name:        req.Name,
		Type:        flagType,
		Description: req.Description,
		Value:       domainValue,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.store.Create(ctx, flag); err != nil {
		return nil, err
	}

	return flagToResponse(flag), nil
}

func parseFlagType(raw string) (domain.FlagType, error) {
	switch domain.FlagType(raw) {
	case domain.FlagTypeBoolean, domain.FlagTypeNumeric:
		return domain.FlagType(raw), nil
	}
	return "", fmt.Errorf("unknown flag type %q: %w", raw, domain.ErrInvalidValue)
}

func flagToResponse(flag domain.Flag) *port.FlagResponse {
	return &port.FlagResponse{
		Name:        flag.Name,
		Type:        string(flag.Type),
		Description: flag.Description,
		Value:       port.FlagValue{Bool: flag.Value.Bool, Numeric: flag.Value.Numeric},
		CreatedAt:   flag.CreatedAt,
		UpdatedAt:   flag.UpdatedAt,
	}
}
