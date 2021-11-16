package redisx

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	*redis.Client
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	res := c.Client.Eval(ctx, script, keys, args...)
	return res.Val(), res.Err()
}

func (c *Client) IsErrNil(err error) bool {
	return err == redis.Nil
}

func (c *Client) Ping() bool {
	return c.Client.Ping(context.Background()).Err() == nil
}

func New(addr string) *Client {
	return &Client{
		redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}
