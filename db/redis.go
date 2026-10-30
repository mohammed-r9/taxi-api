package db

import (
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
}

func NewRedisClient(cfg RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       0,
	})

	return rdb
}
