package storage

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	Cli *redis.Client
}

func NewRedisStorage(addr string) (*RedisStorage, error) {
	opt, err := redis.ParseURL(addr)

	if err != nil {
		log.Fatal("Error al parsear la URL de Redis: ", err)
	}
	rdb := redis.NewClient(opt)

	// Probar si Redis responde
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{Cli: rdb}, nil
}
