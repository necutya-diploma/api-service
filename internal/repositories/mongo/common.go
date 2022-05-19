package mongo

import (
	"errors"

	core "necutya/faker/internal/domain"
	"necutya/faker/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	logger.Error(err)

	if mongo.IsDuplicateKeyError(err) {
		return core.ErrAlreadyExist
	}

	if errors.Is(mongo.ErrNoDocuments, err) {
		return core.ErrNotFound
	}

	return err
}
