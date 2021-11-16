package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// createRedis returns an in process redis.Redis.
func createRedis() (addr string, clean func(), err error) {
	mr, err := miniredis.Run()
	if err != nil {
		return "", nil, err
	}

	return mr.Addr(), func() {
		ch := make(chan struct{})
		go func() {
			mr.Close()
			close(ch)
		}()
		select {
		case <-ch:
		case <-time.After(time.Second):
		}
	}, nil
}

func Test_Remote(t *testing.T) {
	addr, clean, err := createRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	limiter := New(rate, total, NewStore(addr), "test-token-limit")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if limiter.Allow() {
			allowed++
		}
	}

	fmt.Printf("Test_Remote: allowed:%d, burst:%d, rate:%d \n", allowed, burst, rate)
	assert.Truef(t, allowed >= burst+rate, "allowed:%d, burst:%d, rate:%d", allowed, burst, rate)
}

func Test_Remote_Burst(t *testing.T) {
	addr, clean, err := createRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	limiter := New(rate, total, NewStore(addr), "test-token-limit")
	var allowed int
	for i := 0; i < total; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	fmt.Printf("Test_Remote_Burst: allowed:%d, burst:%d, rate:%d \n", allowed, burst, rate)
	assert.Truef(t, allowed >= burst, "allowed:%d, burst:%d, rate:%d", allowed, burst, rate)
}

func Test_Remote_NotAllow(t *testing.T) {
	addr, clean, err := createRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		total = 1
		rate  = 1
	)
	limiter := New(rate, total, NewStore(addr), "test-token-limit")
	var (
		allowed int
		wg      sync.WaitGroup
		mu      sync.Mutex
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if limiter.Allow() {
			mu.Lock()
			allowed++
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if limiter.Allow() {
			mu.Lock()
			allowed++
			mu.Unlock()
		}
	}()
	wg.Wait()

	fmt.Printf("Test_Remote_NotAllow: allowed:%d, rate:%d \n", allowed, rate)
	assert.Truef(t, allowed == 1, "allowed:%d, rate:%d", allowed, rate)
}

func Test_Local(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	const (
		total = 100
		rate  = 5
		burst = 10
	)

	limiter := New(rate, total, NewStore(s.Addr()), "test-token-limit")
	s.Close()

	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if i == total>>1 {
			assert.Nil(t, s.Restart())
		}
		if limiter.Allow() {
			allowed++
		}

		// test monitor
		limiter.monitor()
	}

	assert.Truef(t, allowed >= burst+rate, "allowed:%d, burst:%d, rate:%d", allowed, burst, rate)
}
