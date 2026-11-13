package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	_, err = ValidateAccessToken(token)
	require.ErrorIs(t, err, apperrors.ExpiredToken)
}

func TestValidateAccessTokenMapsMalformedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	_, err := ValidateAccessToken("not-a-jwt")
	require.ErrorIs(t, err, apperrors.InvalidToken)
}

func TestAccessTokenRoundTrip(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	token, err := GenerateAccessToken(UserAccessData{UserID: 7, RoleID: 2})
	require.NoError(t, err)

	data, err := ValidateAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, int64(7), data.UserID)
	require.Equal(t, int64(2), data.RoleID)
}
