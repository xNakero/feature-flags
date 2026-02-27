package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_HappyPath(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://featureflags:featureflags@localhost:5432/featureflags")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("REDIS_ADDR", "localhost:6380")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "postgres://featureflags:featureflags@localhost:5432/featureflags", cfg.PostgresDSN)
	assert.Equal(t, ":9090", cfg.HTTPAddr)
	assert.Equal(t, "localhost:6380", cfg.RedisAddr)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoad_MissingPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://user:pass@db:5432/mydb")
	t.Setenv("HTTP_ADDR", ":3000")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "postgres://user:pass@db:5432/mydb", cfg.PostgresDSN)
	assert.Equal(t, ":3000", cfg.HTTPAddr)
	assert.Equal(t, "redis:6379", cfg.RedisAddr)
	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://featureflags:featureflags@localhost:5432/featureflags")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, "localhost:6379", cfg.RedisAddr)
	assert.Equal(t, "info", cfg.LogLevel)
}
