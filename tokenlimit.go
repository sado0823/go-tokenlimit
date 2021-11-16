package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sado0823/go-tokenlimit/internal/redisx"

	xrate "golang.org/x/time/rate"
)

const (
	tokenFormat     = "{%s}.token-limiter"
	timestampFormat = "{%s}.token-limiter.ts"
	heartbeat       = time.Millisecond * 100
)

const (
	remoteAliveNo = iota
	remoteAliveYes
)

type (
	RedisI interface {
		Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
		IsErrNil(err error) bool
		Ping() bool
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

func NewStore(addr string) RedisI {
	return redisx.New(addr)
}

func New(rate, cap int, store RedisI, key string) *Limiter {
	return &Limiter{
		rate:        rate,
		cap:         cap,
		local:       xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), cap),
		remote:      store,
		tokenKey:    fmt.Sprintf(tokenFormat, key),
		tsKey:       fmt.Sprintf(timestampFormat, key),
		remoteAlive: 1,
	}
}

// Allow is an alias for AllowN(time.Now(),1)
func (l *Limiter) Allow() bool {
	return l.AllowN(time.Now(), 1)
}

// AllowN until this time now, can get n token or not
func (l *Limiter) AllowN(now time.Time, n int) bool {
	return l.reserveN(now, n)
}

// reserveN get token from remote server, if not success, use local xrate
func (l *Limiter) reserveN(now time.Time, n int) bool {
	// if remote server fail, use local rate in process
	if atomic.LoadUint32(&l.remoteAlive) == remoteAliveNo {
		return l.local.AllowN(now, n)
	}

	resp, err := l.remote.Eval(context.Background(),
		tokenLimiter,
		[]string{
			l.tokenKey,
			l.tsKey,
		}, []string{
			strconv.Itoa(l.rate),
			strconv.Itoa(l.cap),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		})

	if l.remote.IsErrNil(err) {
		return false
	} else if err != nil {
		log.Printf("fail to exec eval script, err: %s, degrade to use in-process xrate", err)
		l.monitor()
		return l.local.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		log.Printf("fail to exec eval script, value: %v, degrade to use in-process xrate", resp)
		l.monitor()
		return l.local.AllowN(now, n)
	}

	// redis script allowed lua true => 1
	return code == 1

}

func (l *Limiter) monitor() {
	l.remoteMu.Lock()
	defer l.remoteMu.Unlock()

	if l.monitorStarted {
		return
	}

	l.monitorStarted = true
	atomic.StoreUint32(&l.remoteAlive, remoteAliveNo)

	go l.ping()
}

func (l *Limiter) ping() {
	ticker := time.NewTicker(heartbeat)
	defer func() {
		ticker.Stop()
		l.remoteMu.Lock()
		l.monitorStarted = false
		l.remoteMu.Unlock()
	}()

	for range ticker.C {
		if l.remote.Ping() {
			atomic.StoreUint32(&l.remoteAlive, remoteAliveYes)
			return
		}
	}
}
