package redis

import (
	"errors"

	core "necutya/faker/internal/domain"
	"necutya/faker/pkg/logger"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidValue = errors.New("invalid value")
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	logger.Error("Wrapped error: ", err)

	if errors.Is(ErrNotFound, err) {
		return core.ErrNotFound
	}

	return err
}
