package redis

import (
	"context"
	"fmt"
	redisdb "necutya/faker/pkg/database/redis"
	"strings"
)

const (
	separator       = ":"
	blackListPrefix = "api:inv_token_id"
)

type BlacklistRepo struct {
	cli *redisdb.Client
}

func NewBlacklistRepo(cli *redisdb.Client) *BlacklistRepo {
	return &BlacklistRepo{
		cli: cli,
	}
}

func (r *BlacklistRepo) generateBlackListKey(tokenID string) string {
	return fmt.Sprintf("%s:%s", blackListPrefix, tokenID)
}

func (r *BlacklistRepo) parseBlackListKey(key string) string {
	return strings.Split(key, separator)[2]
}

func (r *BlacklistRepo) AddToken(ctx context.Context, tokenID string, ttl int) error {
	if ttl <= 0 {
		ttl = 0
	}

	return r.cli.Set(ctx, r.generateBlackListKey(tokenID), []byte{}, int64(ttl))
}

func (r *BlacklistRepo) CheckToken(ctx context.Context, tokenID string) error {
	res, err := r.cli.Get(ctx, r.generateBlackListKey(tokenID))

	if err != nil {
		return wrapError(err)
	}

	if res == nil {
		return nil
	}

	return wrapError(ErrInvalidToken)
}
