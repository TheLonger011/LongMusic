package repository

import "github.com/redis/go-redis/v9"

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}
