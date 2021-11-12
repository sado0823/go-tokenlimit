package main

import (
	"context"
	"fmt"
	"sync"

	xrate "golang.org/x/time/rate"
)

type (
	RedisI interface {
		Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	}

	Limiter struct {
		rate int // generate token number each second
		cap  int // at most token to store

		local *xrate.Limiter // limiter in process

		remote         RedisI // for distributed situation, can use redis
		tokenKey       string
		tsKey          string // timestamp key, tag get token time
		remoteMu       sync.Mutex
		remoteAlive    uint32 // ping remote server is alive or not
		monitorStarted bool
	}
)

func New(rate,cap int, store RedisI, key string) *Limiter{
	return nil
}



func main() {
	fmt.Println("go token limit")
}
