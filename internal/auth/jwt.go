package auth

import (
	"errors"
	"fmt"
	"os"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
)

const (
	tokenIssuer   = "taxi"
	tokenAudience = "taxi"

	ScopeAccess       = "access"
	ScopeMFAChallenge = "mfa_challenge"
)

type UserAccessData struct {
	UserID int64  `json:"user_id"`
	RoleID int64  `json:"role_id"`
	Scope  string `json:"scope"`
}

type userAccessClaims struct {
	UserAccessData
	jwt.RegisteredClaims
}

type MFAChallengeData struct {
	UserID  int64
	TokenID string
}

type mfaChallengeClaims struct {
	UserID int64  `json:"user_id"`
	Scope  string `json:"scope"`
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
		UserID:    user.UserID,
		RoleID:    user.RoleID,
		Scope:     ScopeAccess,
		Audience:  jwt.ClaimStrings{tokenAudience},
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenExp)),
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
	if claims.RoleID == 0 {
		return UserAccessData{}, apperrors.InvalidUserID
	}
	if claims.Issuer != tokenIssuer {
		return UserAccessData{}, apperrors.InvalidToken
	}

	return claims.UserAccessData, nil
}

func GenerateMFAChallengeToken(userID int64) (string, string, error) {
	tokenID := uuid.NewV4().String()
	now := time.Now()
	claims := mfaChallengeClaims{
		UserID:    userID,
		Scope:     ScopeMFAChallenge,
		ID:        tokenID,
		Audience:  jwt.ClaimStrings{tokenAudience},
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(MFAChallengeTTL)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
	return token, tokenID, err
}

func ValidateMFAChallengeToken(tokenString string) (MFAChallengeData, error) {
	claims := &mfaChallengeClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc, jwt.WithAudience(tokenAudience))
	if err != nil || !token.Valid || claims.Scope != ScopeMFAChallenge || claims.UserID <= 0 || claims.ID == "" || claims.Issuer != tokenIssuer {
		return MFAChallengeData{}, apperrors.InvalidChallengeToken
	}
	return MFAChallengeData{UserID: claims.UserID, TokenID: claims.ID}, nil
}
