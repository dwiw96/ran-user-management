package redisPkg

import (
	"context"
	"os"
	"testing"
	"time"

	config "github.com/dwiw96/ran-user-management/config"
	conv "github.com/dwiw96/ran-user-management/pkg/utils/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectToRedis(t *testing.T) {
	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "user")
	os.Setenv("REDIS_DB", "0")

	redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
	require.NoError(t, err)

	envConfig := &config.EnvConfig{
		REDIS_HOST:     os.Getenv("REDIS_HOST"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       redis_db,
	}

	client := ConnectToRedis(envConfig)
	defer client.Close()
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	t.Run("set", func(t *testing.T) {
		err = client.Set(ctx, "test", "test redis", 0).Err()
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		val, err := client.Get(ctx, "test").Result()
		require.NoError(t, err)
		assert.Equal(t, "test redis", val)
	})

	t.Run("del", func(t *testing.T) {
		resDel, err := client.Del(ctx, "test").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), resDel)
	})
}
