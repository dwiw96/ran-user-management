package testutils

import (
	"context"
	"fmt"
	"log"
	"os"

	"path/filepath"
	"runtime"
	"sync"

	cfg "github.com/dwiw96/ran-user-management/config"
	pg "github.com/dwiw96/ran-user-management/pkg/driver/postgresql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ctx  context.Context
	pool *pgxpool.Pool
	once sync.Once
)

func GetPool() *pgxpool.Pool {
	once.Do(func() {
		os.Setenv("DB_USERNAME", "user")
		os.Setenv("DB_PASSWORD", "user")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_NAME", "user")

		env := &cfg.EnvConfig{
			DB_USERNAME: os.Getenv("DB_USERNAME"),
			DB_PASSWORD: os.Getenv("DB_PASSWORD"),
			DB_HOST:     os.Getenv("DB_HOST"),
			DB_PORT:     os.Getenv("DB_PORT"),
			DB_NAME:     os.Getenv("DB_NAME"),
		}

		pool = pg.ConnectToPg(env)
	})

	return pool
}

func ClosePool() {
	if pool != nil {
		pool.Close()
	}
}

func GetContext() context.Context {
	ctx = context.Background()

	return ctx
}

func SetupDB(schemaName string) func() {
	var err error

	dropSchema := func() {
		_, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
		if err != nil {
			log.Fatalf("db cleanup failed, err: %v", err)
		}
	}

	dropSchema()

	// create test schema
	_, err = pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+schemaName)
	if err != nil {
		log.Fatalf("schema creation faild. err: %v", err)
	}

	// use schema
	_, err = pool.Exec(ctx, "SET search_path TO "+schemaName)
	if err != nil {
		log.Fatalf("error while switching to schema. err: %v", err)
	}

	env := cfg.GetEnvConfig()
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	projectPath := filepath.Dir((basePath))
	migrationPath := "file:" + filepath.Join(projectPath, "/internal/migrations/")

	pgAddress := fmt.Sprintf("pgx5://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", env.DB_USERNAME, env.DB_PASSWORD, env.DB_HOST, env.DB_PORT, env.DB_NAME, schemaName)
	m, err := migrate.New(
		migrationPath,
		pgAddress)
	if err != nil {
		dropSchema()
		log.Fatal("migrate new, err:", err)
	}

	err = m.Up()
	if err != nil {
		log.Fatal("m.Up, err:", err)
		dropSchema()
	}

	return dropSchema
}

func DeleteSchemaTestData(pool *pgxpool.Pool) error {
	var err error

	tx, err := pool.Begin(ctx)

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	const query = `
	TRUNCATE TABLE
		refresh_token_whitelist,
		users
	RESTART IDENTITY CASCADE;
	`
	_, err = tx.Exec(ctx, query)

	return err
}
