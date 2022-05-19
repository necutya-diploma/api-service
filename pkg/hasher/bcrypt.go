package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher
// Implement bcrypt hashing
// Bcrypt add salt to every hash under the hood
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(optionalCost ...int) BcryptHasher {
	var cost int

	if len(optionalCost) >= 1 {
		cost = optionalCost[0]
	} else {
		cost = bcrypt.DefaultCost
	}

	return BcryptHasher{cost: cost}
}

func (h BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), err
}

func (h BcryptHasher) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
