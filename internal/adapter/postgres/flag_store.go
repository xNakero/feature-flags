package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xNakero/feature-flags/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS flags (
    name          TEXT PRIMARY KEY,
    type          TEXT             NOT NULL CHECK (type IN ('boolean', 'numeric')),
    description   TEXT             NOT NULL DEFAULT '',
    bool_value    BOOLEAN,
    numeric_value DOUBLE PRECISION,
    created_at    TIMESTAMPTZ      NOT NULL,
    updated_at    TIMESTAMPTZ      NOT NULL,
    CONSTRAINT exactly_one_value CHECK (
        (type = 'boolean' AND bool_value IS NOT NULL AND numeric_value IS NULL) OR
        (type = 'numeric' AND numeric_value IS NOT NULL AND bool_value IS NULL)
    )
);`

type FlagStore struct {
	pool *pgxpool.Pool
}

func NewFlagStore(pool *pgxpool.Pool) *FlagStore {
	return &FlagStore{pool: pool}
}

func (s *FlagStore) CreateSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schema)
	return err
}

func (s *FlagStore) Create(ctx context.Context, flag domain.Flag) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO flags (name, type, description, bool_value, numeric_value, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		flag.Name, string(flag.Type), flag.Description,
		flag.Value.Bool, flag.Value.Numeric,
		flag.CreatedAt, flag.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *FlagStore) GetByName(ctx context.Context, name string) (*domain.Flag, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT name, type, description, bool_value, numeric_value, created_at, updated_at
		 FROM flags WHERE name = $1`,
		name,
	)
	return scanFlag(row)
}

func (s *FlagStore) UpdateValue(ctx context.Context, name string, flagValue domain.FlagValue) (*domain.Flag, error) {
	now := time.Now().UTC()
	row := s.pool.QueryRow(ctx,
		`UPDATE flags
		 SET bool_value = $1, numeric_value = $2, updated_at = $3
		 WHERE name = $4
		 RETURNING name, type, description, bool_value, numeric_value, created_at, updated_at`,
		flagValue.Bool, flagValue.Numeric, now, name,
	)
	flag, err := scanFlag(row)
	if err != nil {
		return nil, err
	}
	return flag, nil
}

func scanFlag(row pgx.Row) (*domain.Flag, error) {
	var (
		flag    domain.Flag
		rawType string
	)
	err := row.Scan(
		&flag.Name, &rawType, &flag.Description,
		&flag.Value.Bool, &flag.Value.Numeric,
		&flag.CreatedAt, &flag.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w", domain.ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	flag.Type = domain.FlagType(rawType)
	return &flag, nil
}
