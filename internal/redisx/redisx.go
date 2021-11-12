package redisx

import (
	"github.com/go-redis/redis/v8"
)

func New(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
