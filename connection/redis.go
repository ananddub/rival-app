package connection

import (
	"context"
	"fmt"
	"time"

	"rival/config"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func GetRedisClient(config *config.RedisConfig) *redis.Client {
	if redisClient != nil {
		return redisClient
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     "",
		DB:           config.Db,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// health check
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	redisClient = rdb
	return rdb
}
