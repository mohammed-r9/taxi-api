package auth

import (
	"errors"
	"testing"

	"github.com/mohammed-r9/taxi-api/internal/apperrors"
)

func TestOpaqueTokenRoundTrip(t *testing.T) {
	token := NewOpaqueToken()

	parsed, err := ParseOpaqueToken(token.String())
	if err != nil {
		t.Fatalf("ParseOpaqueToken: %v", err)
	}
	if !token.Equal(parsed) {
		t.Fatal("parsed token differs from original")
	}
	if token.Hex() != parsed.Hex() {
		t.Fatal("hex representations differ")
	}
	if token.Hash() != parsed.Hash() {
		t.Fatal("hashes differ")
	}
}

func TestParseOpaqueTokenRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", "not-a-token", "AAAA"} {
		if _, err := ParseOpaqueToken(input); !errors.Is(err, apperrors.InvalidToken) {
			t.Errorf("ParseOpaqueToken(%q) error = %v, want InvalidToken", input, err)
		}
	}
}

func TestNewOpaqueTokenIsUnique(t *testing.T) {
	first := NewOpaqueToken()
	second := NewOpaqueToken()

	if first.Equal(second) {
		t.Fatal("two generated tokens are equal")
	}
}

func TestOpaqueTokenZeroValue(t *testing.T) {
	var token OpaqueToken
	if !token.IsZero() {
		t.Fatal("zero value token is not zero")
	}
	if NewOpaqueToken().IsZero() {
		t.Fatal("generated token is zero")
	}
}
