package db

import (
	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return rdb
}
