package testutils

import (
	"context"
	"os"
	"testing"

	cfg "github.com/dwiw96/ran-user-management/config"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	poolTest        *pgxpool.Pool
	redisClientTest *redis.Client
	ctxTest         context.Context
)

func TestMain(m *testing.M) {
	poolTest = GetPool()
	redisClientTest = GetRedisClient()
	ctxTest = GetContext()

	exit := m.Run()

	poolTest.Close()
	redisClientTest.Close()
	ctxTest.Done()

	os.Exit(exit)
}

var queryTest = `
	INSERT INTO users(
		username,
		email,
		hashed_password
	) VALUES (
		$1,
		$2,
		$3
	) returning username, email, hashed_password`

func TestGetPool(t *testing.T) {
	schemaCleanup := SetupDB("test_testutils_testgetpool")
	t.Run("test_env", func(t *testing.T) {
		envAns := &cfg.EnvConfig{
			DB_USERNAME: "user",
			DB_PASSWORD: "user",
			DB_HOST:     "localhost",
			DB_PORT:     "5432",
			DB_NAME:     "user",
		}

		envTest := &cfg.EnvConfig{
			DB_USERNAME: os.Getenv("DB_USERNAME"),
			DB_PASSWORD: os.Getenv("DB_PASSWORD"),
			DB_HOST:     os.Getenv("DB_HOST"),
			DB_PORT:     os.Getenv("DB_PORT"),
			DB_NAME:     os.Getenv("DB_NAME"),
		}
		require.Equal(t, envAns, envTest)
	})

	t.Run("test_db", func(t *testing.T) {
		type dbRes struct {
			name     string
			email    string
			password string
		}

		var res dbRes
		row := poolTest.QueryRow(ctxTest, queryTest, "getpool", "getpool@mail.com", "password123")
		err := row.Scan(&res.name, &res.email, &res.password)
		require.NoError(t, err)
	})
	schemaCleanup()
}

func TestClosePool(t *testing.T) {
	var err error

	schemaCleanup := SetupDB("test_testutils_testclosepool")

	type dbRes struct {
		name     string
		email    string
		password string
	}

	var res dbRes
	row := poolTest.QueryRow(ctxTest, queryTest, "closepool1", "closepool1@mail.com", "password456")
	err = row.Scan(&res.name, &res.email, &res.password)
	require.NoError(t, err)

	schemaCleanup()
	ClosePool()

	row = poolTest.QueryRow(ctxTest, queryTest, "closepool2", "closepool2@mail.com", "password789")
	err = row.Scan(&res.name, &res.email, &res.password)
	require.Error(t, err)
}
