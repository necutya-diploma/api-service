package redisdb

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// NewClient established connection to a Redis instance.
func NewClient(uri, password string, poolSize int) (*Client, error) {
	cli := redis.NewClient(
		&redis.Options{
			PoolSize: poolSize,
			Addr:     uri,
			Password: password,
		})

	err := cli.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	client := &Client{cli: cli}

	return client, nil
}
