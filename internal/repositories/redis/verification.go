package redis

import (
	"context"

	redisdb "necutya/faker/pkg/database/redis"
)

type VerificationRepo struct {
	cli *redisdb.Client
}

func NewVerificationRepo(cli *redisdb.Client) *VerificationRepo {
	return &VerificationRepo{
		cli: cli,
	}
}

func (r *VerificationRepo) SetCode(ctx context.Context, email string, code string, seconds int) error {
	return wrapError(r.cli.Set(ctx, email, []byte(code), int64(seconds)))
}

func (r *VerificationRepo) GetCode(ctx context.Context, email string) (string, error) {
	code, err := r.cli.Get(ctx, email)
	if code == nil && err == nil {
		return "", wrapError(ErrNotFound)
	}

	return string(code), nil
}
