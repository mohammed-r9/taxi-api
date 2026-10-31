package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
)

func TestValidateAccessTokenMapsExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	claims := userAccessClaims{
		UserAccessData: UserAccessData{UserID: 1, Scope: ScopeAccess},
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{tokenAudience},
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateAccessToken(token)
	if !errors.Is(err, apperrors.ExpiredToken) {
		t.Fatalf("error = %v, want ExpiredToken", err)
	}
}

func TestValidateAccessTokenMapsMalformedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	_, err := ValidateAccessToken("not-a-jwt")
	if !errors.Is(err, apperrors.InvalidToken) {
		t.Fatalf("error = %v, want InvalidToken", err)
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	token, err := GenerateAccessToken(UserAccessData{UserID: 7, RoleID: 2})
	if err != nil {
		t.Fatal(err)
	}

	data, err := ValidateAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if data.UserID != 7 || data.RoleID != 2 {
		t.Fatalf("claims = %+v", data)
	}
}
