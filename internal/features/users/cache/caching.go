package chache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	pUsers "github.com/dwiw96/ran-user-management/internal/features/users"
)

type usersCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewUsersCache(client *redis.Client, ctx context.Context) pUsers.ICache {
	return &usersCache{
		client: client,
		ctx:    ctx,
	}
}

func (c *usersCache) CachingBlockedToken(payload pUsers.JwtPayload) error {
	now := time.Now().UTC()
	exp := time.Unix(payload.Exp, 0)

	duration := time.Duration(exp.Sub(now).Nanoseconds())
	if duration <= 0 {
		return nil
	}

	err := c.client.Set(c.ctx, fmt.Sprint("block ", payload.ID), payload.UserID, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to caching token, msg: %v", err)
	}

	return nil
}

func (r *usersCache) CheckBlockedToken(payload pUsers.JwtPayload) error {
	check, err := r.client.Exists(r.ctx, "block "+payload.ID.String()).Result()
	if err != nil {
		return err
	}
	if check != 0 {
		return fmt.Errorf("token is blacklist")
	}

	return nil
}
