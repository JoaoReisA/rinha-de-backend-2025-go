package database

import (
	"context"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var RedisCtx = context.Background()

func ConnectRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.CacheURL,
	})
	return nil
}
