package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"time"
)

const (
	twoFATTL = time.Minute * 5
)

func GenerateUpaqueToken() []byte {
	buf := make([]byte, 32)
	rand.Read(buf)
	return buf
}

func HashToken(t []byte) [32]byte {
	return sha256.Sum256(t)

}

func CompareOpaqueTokens(t1, t2 []byte) bool {
	return subtle.ConstantTimeCompare(t1, t2) == 1
}
