package redisdb

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

// SetClient sets the Redis client (initialized in connections layer)
func SetClient(client *redis.Client) {
	Rdb = client
}
