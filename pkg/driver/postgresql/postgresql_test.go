package postgresql

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	config "github.com/dwiw96/ran-user-management/config"

	"github.com/stretchr/testify/require"
)

func TestConnectToPg(t *testing.T) {
	os.Setenv("DB_USERNAME", "user")
	os.Setenv("DB_PASSWORD", "user")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "user")

	envConfig := &config.EnvConfig{
		DB_USERNAME: os.Getenv("DB_USERNAME"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
	}
	dbPool := ConnectToPg(envConfig)
	require.NotNil(t, dbPool)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	var greeting string
	err := dbPool.QueryRow(ctx, "select 'Hello, world!'").Scan(&greeting)
	require.NoError(t, err)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
	defer dbPool.Close()
}
