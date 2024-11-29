package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	conv "github.com/dwiw96/ran-user-management/pkg/utils/converter"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	SERVER_PORT    string
	DB_USERNAME    string
	DB_PASSWORD    string
	DB_HOST        string
	DB_PORT        string
	DB_NAME        string
	REDIS_HOST     string
	REDIS_PASSWORD string
	REDIS_DB       int
}

func GetEnvConfig() *EnvConfig {
	initEnvConfig()
	var err error

	var resEnvConfig EnvConfig
	resEnvConfig.SERVER_PORT = os.Getenv("SERVER_PORT")
	resEnvConfig.DB_USERNAME = os.Getenv("DB_USERNAME")
	resEnvConfig.DB_PASSWORD = os.Getenv("DB_PASSWORD")
	resEnvConfig.DB_HOST = os.Getenv("DB_HOST")
	resEnvConfig.DB_PORT = os.Getenv("DB_PORT")
	resEnvConfig.DB_NAME = os.Getenv("DB_NAME")
	resEnvConfig.REDIS_HOST = os.Getenv("REDIS_HOST")
	resEnvConfig.REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")
	resEnvConfig.REDIS_DB, err = conv.ConvertStrToInt(os.Getenv("REDIS_DB"))

	if err != nil {
		log.Fatal("get env config, err:", err)
	}

	return &resEnvConfig
}

func initEnvConfig() {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	projectPath := filepath.Dir((basePath))
	envPath := filepath.Join(projectPath, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Println("failed to load .env file, msg:", err)
		return
	}
}
