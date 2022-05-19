package redis

import (
	"context"
	"strconv"
	"time"

	"necutya/faker/internal/domain/domain"
	redisdb "necutya/faker/pkg/database/redis"
)

type RequestCounterRepo struct {
	cli *redisdb.Client
}

func NewRequestCounterRepo(cli *redisdb.Client) *RequestCounterRepo {
	return &RequestCounterRepo{
		cli: cli,
	}
}

func (r *RequestCounterRepo) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return wrapError(r.cli.Expire(ctx, key, ttl))
}

func (r *RequestCounterRepo) Incr(ctx context.Context, userID string, _type domain.RequestType) error {
	return wrapError(r.cli.HIncrBy(ctx, userID, string(_type), 1))
}

func (r *RequestCounterRepo) GetByUserID(ctx context.Context, userID string, _type domain.RequestType) (int, error) {
	value := r.cli.HMGet(ctx, userID, string(_type))
	if len(value) < 0 {
		return 0, wrapError(ErrNotFound)
	}

	switch v := value[0].(type) {
	case string:
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return 0, wrapError(ErrInvalidValue)
		}

		return intValue, nil
	}

	return 0, nil
}
