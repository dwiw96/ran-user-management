package postgresql

import (
	"context"
	"fmt"
	"log"
	"os"

	config "github.com/dwiw96/ran-user-management/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPg(envCfg *config.EnvConfig) *pgxpool.Pool {
	pgAddress := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", envCfg.DB_USERNAME, envCfg.DB_PASSWORD, envCfg.DB_HOST, envCfg.DB_PORT, envCfg.DB_NAME)
	configPgAddress, err := pgxpool.ParseConfig(pgAddress)
	if err != nil {
		log.Fatal("cannot parse config, msg:", err)
	}

	dbPool, err := pgxpool.New(context.Background(), configPgAddress.ConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbPool
}
