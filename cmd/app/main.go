package main

import (
	"context"
	"log"

	cfg "github.com/dwiw96/ran-user-management/config"
	factory "github.com/dwiw96/ran-user-management/factory"
	pg "github.com/dwiw96/ran-user-management/pkg/driver/postgresql"
	rd "github.com/dwiw96/ran-user-management/pkg/driver/redis"
	server "github.com/dwiw96/ran-user-management/server"

	password "github.com/dwiw96/ran-user-management/pkg/utils/password"
)

func main() {
	log.Println("-- run ran-user-management --")
	env := cfg.GetEnvConfig()
	pgPool := pg.ConnectToPg(env)
	defer pgPool.Close()

	rdClient := rd.ConnectToRedis(env)
	defer rdClient.Close()

	ctx := context.Background()
	defer ctx.Done()

	password.JwtInit(pgPool, ctx)

	router := server.SetupRouter()

	factory.InitFactory(router, pgPool, rdClient, ctx)

	server.StartServer(env.SERVER_PORT, router)
}
