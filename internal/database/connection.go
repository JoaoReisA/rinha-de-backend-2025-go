package database

import (
	"context"

	"github.com/JoaoReisA/rinha-de-backend-2025-go/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

var ConnectionPg *pgxpool.Pool
var PgCtx = context.Background()
var RedisClient *redis.Client
var RedisCtx = context.Background()

func ConnectPG() error {
	var err error
	ConnectionPg, err = pgxpool.Connect(PgCtx, config.DatabaseURL)

	return err
}

func Close() {
	if ConnectionPg == nil {
		return
	}
	ConnectionPg.Close()
}
func ConnectRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.CacheURL,
	})
	return nil
}
