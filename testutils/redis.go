package testutils

import (
	"log"
	"os"

	cfg "github.com/dwiw96/ran-user-management/config"
	rd "github.com/dwiw96/ran-user-management/pkg/driver/redis"
	conv "github.com/dwiw96/ran-user-management/pkg/utils/converter"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
)

func GetRedisClient() *redis.Client {
	if redisClient == nil {
		os.Setenv("REDIS_HOST", "localhost:6379")
		os.Setenv("REDIS_PASSWORD", "user")
		os.Setenv("REDIS_DB", "0")

		redis_db, err := conv.ConvertStrToInt(os.Getenv("REDIS_DB"))
		if err != nil {
			log.Fatal(err)
		}

		env := &cfg.EnvConfig{
			REDIS_HOST:     os.Getenv("REDIS_HOST"),
			REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
			REDIS_DB:       redis_db,
		}

		redisClient = rd.ConnectToRedis(env)
	}

	return redisClient
}

func CloseRedisClient() {
	if redisClient != nil {
		redisClient.Close()
	}
}
