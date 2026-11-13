package auth

import (
	"testing"

	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/stretchr/testify/require"
)

func TestOpaqueTokenRoundTrip(t *testing.T) {
	token := NewOpaqueToken()

	parsed, err := ParseOpaqueToken(token.String())
	require.NoError(t, err)
	require.Equal(t, token, parsed)
	require.Equal(t, token.Hex(), parsed.Hex())
	require.Equal(t, token.Hash(), parsed.Hash())
}

func TestParseOpaqueTokenRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", "not-a-token", "AAAA"} {
		_, err := ParseOpaqueToken(input)
		require.ErrorIs(t, err, apperrors.InvalidToken)
	}
}

func TestNewOpaqueTokenIsUnique(t *testing.T) {
	first := NewOpaqueToken()
	second := NewOpaqueToken()

	require.NotEqual(t, first, second)
}

func TestOpaqueTokenZeroValue(t *testing.T) {
	var token OpaqueToken
	require.True(t, token.IsZero())
	require.False(t, NewOpaqueToken().IsZero())
}
