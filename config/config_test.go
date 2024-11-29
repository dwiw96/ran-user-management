package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitEnvConfig(t *testing.T) {
	initEnvConfig()
	require.NotPanics(t, initEnvConfig)
}

func TestGetEnvConfig(t *testing.T) {
	ans := &EnvConfig{
		SERVER_PORT:    ":8080",
		DB_USERNAME:    "user",
		DB_PASSWORD:    "user",
		DB_HOST:        "postgres",
		DB_PORT:        "5432",
		DB_NAME:        "user",
		REDIS_HOST:     "redis:6379",
		REDIS_PASSWORD: "user",
		REDIS_DB:       0,
	}
	res := GetEnvConfig()
	require.NotNil(t, res)
	assert.Equal(t, ans, res)
}
