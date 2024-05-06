package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var CTX = context.Background()

func CreateClient(DBNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       DBNo,
	})

	return rdb
}
