package testutils

import (
	"context"
	"os"
	"testing"
	"time"

	config "github.com/dwiw96/ran-user-management/config"
	conv "github.com/dwiw96/ran-user-management/pkg/utils/converter"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRedisClient(t *testing.T) {
	var err error

	t.Run("test_env", func(t *testing.T) {
		envAns := &config.EnvConfig{
			REDIS_HOST:     "localhost:6379",
			REDIS_PASSWORD: "user",
			REDIS_DB:       0,
		}

		redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
		require.NoError(t, err)

		envTest := &config.EnvConfig{
			REDIS_HOST:     os.Getenv("REDIS_HOST"),
			REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
			REDIS_DB:       redis_db,
		}
		require.Equal(t, envAns, envTest)
	})

	err = redisClientTest.Ping(ctxTest).Err()
	require.NoError(t, err)

	t.Run("set", func(t *testing.T) {
		err = redisClientTest.Set(ctxTest, "test", "test redis", 0).Err()
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		val, err := redisClientTest.Get(ctxTest, "test").Result()
		require.NoError(t, err)
		assert.Equal(t, "test redis", val)
	})

	t.Run("del", func(t *testing.T) {
		resDel, err := redisClientTest.Del(ctxTest, "test").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), resDel)
	})
}

func TestCloseRedisClient(t *testing.T) {
	client := GetRedisClient()
	require.NotNil(t, client)

	ctxTest, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := redisClientTest.Ping(ctxTest).Err()
	require.NoError(t, err)

	CloseRedisClient()
	err = redisClientTest.Ping(ctxTest).Err()
	require.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
}
