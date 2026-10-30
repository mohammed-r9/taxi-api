package auth

import (
	"github.com/alexedwards/argon2id"
)

var defaultParams = &argon2id.Params{
	Memory:      32 * 1024,
	Iterations:  3,
	Parallelism: 2,
	KeyLength:   32,
	SaltLength:  16,
}

type ComparePasswordParams struct {
	Password   string
	StoredHash string
}

func HashPassword(plainText string) (string, error) {
	return argon2id.CreateHash(plainText, defaultParams)
}

func ComparePassword(params ComparePasswordParams) (bool, error) {
	return argon2id.ComparePasswordAndHash(params.Password, params.StoredHash)
}
