package chache

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"testing"
	"time"

	auth "github.com/dwiw96/ran-user-management/internal/features/users"
	middleware "github.com/dwiw96/ran-user-management/pkg/middleware"
	generator "github.com/dwiw96/ran-user-management/pkg/utils/generator"
	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
	testUtils "github.com/dwiw96/ran-user-management/testutils"

	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cacheTest auth.ICache
	poolTest  *pgxpool.Pool
	client    *redis.Client
	ctx       context.Context
	key       *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	poolTest = testUtils.GetPool()
	defer poolTest.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("test_cache_users")

	password.JwtInit(poolTest, ctx)

	client = testUtils.GetRedisClient()
	defer client.Close()

	cacheTest = NewUsersCache(client, ctx)

	exitTest := m.Run()

	schemaCleanup()
	poolTest.Close()
	ctx.Done()
	client.Close()

	os.Exit(exitTest)
}

func createToken(t *testing.T) (payload *auth.JwtPayload) {
	var err error
	key, err = middleware.LoadKey(ctx, poolTest)
	require.NoError(t, err)
	require.NotNil(t, key)

	user := auth.User{
		ID:       generator.RandomInt32(1, 100),
		Username: generator.CreateRandomString(5) + " " + generator.CreateRandomString(7),
		Email:    generator.CreateRandomEmail(generator.CreateRandomString(5)),
	}
	token, err := middleware.CreateToken(user, 5, key)
	require.NoError(t, err)
	require.NotZero(t, len(token))

	payload, err = middleware.ReadToken(token, key)
	require.NoError(t, err)

	return
}

func TestCachingBlockedToken(t *testing.T) {
	var err error

	err = testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	tests := []struct {
		name    string
		payload *auth.JwtPayload
		err     bool
	}{
		{
			name:    "success",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "success",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "failed_duration_minus",
			payload: createToken(t),
			err:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.err {
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				res, err := client.Get(ctx, fmt.Sprint("block ", test.payload.ID)).Result()
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprint(test.payload.UserID), res)
			} else {
				now := time.Now().UTC().Add(1)
				test.payload.Exp = now.Unix()
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				res, err := client.Get(ctx, fmt.Sprint("block ", test.payload.ID)).Result()
				require.Error(t, err)
				assert.Empty(t, res)
			}
		})
	}
}

func TestCheckBlockedToken(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(poolTest)
	require.NoError(t, err)

	tests := []struct {
		name    string
		payload *auth.JwtPayload
		err     bool
	}{
		{
			name:    "valid",
			payload: createToken(t),
			err:     false,
		}, {
			name:    "blacklist",
			payload: createToken(t),
			err:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.err {
				err = cacheTest.CheckBlockedToken(*test.payload)
			} else {
				err = cacheTest.CachingBlockedToken(*test.payload)
				require.NoError(t, err)

				err = cacheTest.CheckBlockedToken(*test.payload)
				require.Error(t, err)
			}
		})
	}
}
