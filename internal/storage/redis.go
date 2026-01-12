package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	Cli *redis.Client
}

func NewRedisStorage(addr string) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Probar si Redis responde
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{Cli: rdb}, nil
}
