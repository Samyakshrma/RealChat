package utils

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func InitRedis(ctx context.Context) {
	Rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Optionally, test the connection
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}
