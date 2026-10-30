package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-r9/go-boiler/internal/apperrors"
)

const (
	tokenIssuer   = "go-boiler"
	tokenAudience = "go-boiler"

	ScopeAccess = "access"
)

type UserAccessData struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role,omitempty"`
	Scope  string `json:"scope"`
}

type userAccessClaims struct {
	UserAccessData
	jwt.RegisteredClaims
}

func GenerateAccessToken(user UserAccessData) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	now := time.Now()

	tokenExp := 15 * time.Minute
	if os.Getenv("LOCAL") == "true" {
		tokenExp = 15 * time.Hour
	}

	claims := userAccessClaims{
		UserAccessData: UserAccessData{
			UserID: user.UserID,
			Role:   user.Role,
			Scope:  ScopeAccess,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{tokenAudience},
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExp)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func ValidateAccessToken(tokenString string) (UserAccessData, error) {
	claims := &userAccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithAudience(tokenAudience))
	if err != nil {
		publicErr := error(apperrors.InvalidToken)
		if errors.Is(err, jwt.ErrTokenExpired) {
			publicErr = apperrors.ExpiredToken
		}
		return UserAccessData{}, fmt.Errorf("%w: parse access token: %w", publicErr, err)
	}
	if !token.Valid {
		return UserAccessData{}, apperrors.InvalidToken
	}
	if claims.Scope != ScopeAccess {
		return UserAccessData{}, apperrors.InvalidToken
	}
	if claims.UserID == 0 {
		return UserAccessData{}, apperrors.InvalidUserID
	}
	if claims.Issuer != tokenIssuer {
		return UserAccessData{}, apperrors.InvalidToken
	}

	return claims.UserAccessData, nil
}
