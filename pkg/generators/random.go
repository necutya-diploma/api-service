package generators

import (
	"encoding/base64"
	"math/rand"
	"time"
)

var (
	numSet = []rune("0123456789")
)

type RandomGenerator struct {
}

func NewRandomGenerator() *RandomGenerator {
	return &RandomGenerator{}
}

func (rg *RandomGenerator) GenerateNumericCode(length int) string {
	rand.Seed(time.Now().Unix())

	code := make([]rune, length)
	for i := range code {
		code[i] = numSet[rand.Intn(len(numSet))]
	}

	return string(code)
}

func (rg *RandomGenerator) GenerateString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}
