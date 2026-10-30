package auth

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(os.Getenv("JWT_SECRET")), nil
}
