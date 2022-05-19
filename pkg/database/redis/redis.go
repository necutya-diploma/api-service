package redisdb

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client for accessing to Redis.
type Client struct {
	cli *redis.Client
}

func (r *Client) Incr(ctx context.Context, key string) error {
	return r.cli.Incr(ctx, key).Err()
}

func (r *Client) Exists(ctx context.Context, key string) (bool, error) {
	res, err := r.cli.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if res == 1 {
		return true, nil
	}

	return false, nil
}

func (r *Client) GetString(ctx context.Context, key string) string {
	return r.cli.Get(ctx, key).String()
}

// TTL - returns TTL by the KEY.
func (r *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.cli.TTL(ctx, key).Result()
}

// Get - implementation get value from redis by key.
func (r *Client) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := r.cli.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		// it means that key does not exist in redis
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return res, nil
}

// GetInt - implementation get int value from redis by key.
func (r *Client) GetInt(ctx context.Context, key string) (int, error) {
	res, err := r.cli.Get(ctx, key).Int()

	if errors.Is(err, redis.Nil) {
		// it means that key does not exist in redis
		return 0, nil
	}

	return res, err
}

// SetInt - implementation set int value to redis by key and set ttl.
func (r *Client) SetInt(ctx context.Context, key string, value int, ttl int64) error {
	return r.cli.Set(ctx, key, value, time.Second*time.Duration(ttl)).Err()
}

// Set - implementation set value to redis by key and set ttl.
func (r *Client) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	return r.cli.Set(ctx, key, value, time.Second*time.Duration(ttl)).Err()
}

// Del - implementation delete keys from redis.
func (r *Client) Del(ctx context.Context, keys ...string) error {
	return r.cli.Del(ctx, keys...).Err()
}

// HMSet - sets map(hash) to redis by key.
func (r *Client) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	return r.cli.HMSet(ctx, key, fields).Err()
}

// HMGet - returns `fields` from map(hash), found by `key`.
func (r *Client) HMGet(ctx context.Context, key string, fields ...string) []interface{} {
	return r.cli.HMGet(ctx, key, fields...).Val()
}

// HGetAll - returns map(hash) by `key`.
func (r *Client) HGetAll(ctx context.Context, key string) map[string]string {
	return r.cli.HGetAll(ctx, key).Val()
}

// HIncrBy - finds map(hash) by `key` and increments its `field`.
func (r *Client) HIncrBy(ctx context.Context, key, field string, incr int64) error {
	return r.cli.HIncrBy(ctx, key, field, incr).Err()
}

// Expire - adds `exp` - expiration time for the `key`.
func (r *Client) Expire(ctx context.Context, key string, exp time.Duration) error {
	return r.cli.Expire(ctx, key, exp).Err()
}

// Ping - pings the redis client connection.
func (r *Client) Ping(ctx context.Context) error {
	return r.cli.Ping(ctx).Err()
}
