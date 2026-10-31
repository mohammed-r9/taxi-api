package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

const opaqueTokenSize = 32

// ErrInvalidOpaqueToken is returned when a string cannot be decoded into an
// opaque token
var ErrInvalidOpaqueToken = errors.New("auth: invalid opaque token")

// OpaqueToken is a fixed-size, cryptographically random token
//
// The zero value is not a valid token, use NewOpaqueToken or ParseOpaqueToken
type OpaqueToken struct {
	raw [opaqueTokenSize]byte
}

// NewOpaqueToken generates a new random opaque token
func NewOpaqueToken() OpaqueToken {
	var token OpaqueToken
	if _, err := rand.Read(token.raw[:]); err != nil {
		panic("auth: failed to generate opaque token: " + err.Error())
	}
	return token
}

// ParseOpaqueToken decodes the canonical base64url (unpadded) representation
// produced by String
func ParseOpaqueToken(s string) (OpaqueToken, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return OpaqueToken{}, fmt.Errorf("%w: %v", ErrInvalidOpaqueToken, err)
	}
	if len(raw) != opaqueTokenSize {
		return OpaqueToken{}, fmt.Errorf(
			"%w: got %d bytes, want %d",
			ErrInvalidOpaqueToken,
			len(raw),
			opaqueTokenSize,
		)
	}

	var token OpaqueToken
	copy(token.raw[:], raw)
	return token, nil
}

// Bytes returns a copy of the raw token bytes
func (t OpaqueToken) Bytes() []byte {
	return t.raw[:]
}

// String returns the unpadded base64url representation of the token
func (t OpaqueToken) String() string {
	return base64.RawURLEncoding.EncodeToString(t.raw[:])
}

// Hex returns the lowercase hexadecimal representation of the token
func (t OpaqueToken) Hex() string {
	return hex.EncodeToString(t.raw[:])
}

// Hash returns the SHA-256 hash of the token
func (t OpaqueToken) Hash() [sha256.Size]byte {
	return sha256.Sum256(t.raw[:])
}

// Equal compares two tokens in constant time
func (t OpaqueToken) Equal(other OpaqueToken) bool {
	return subtle.ConstantTimeCompare(t.raw[:], other.raw[:]) == 1
}

// IsZero reports whether the token is the zero value
func (t OpaqueToken) IsZero() bool {
	return t == OpaqueToken{}
}
