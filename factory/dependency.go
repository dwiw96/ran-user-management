package factory

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authCache "github.com/dwiw96/ran-user-management/internal/features/users/cache"
	authHandler "github.com/dwiw96/ran-user-management/internal/features/users/handler"
	authRepository "github.com/dwiw96/ran-user-management/internal/features/users/repository"
	authService "github.com/dwiw96/ran-user-management/internal/features/users/service"
)

func InitFactory(router *gin.Engine, pool *pgxpool.Pool, rdClient *redis.Client, ctx context.Context) {
	iAuthRepo := authRepository.NewUsersRepository(pool, pool)
	iAuthCache := authCache.NewUsersCache(rdClient, ctx)
	iAuthService := authService.NewUsersService(iAuthRepo, iAuthCache, ctx)
	authHandler.NewUsersHandler(router, iAuthService, pool, rdClient, ctx)
}
