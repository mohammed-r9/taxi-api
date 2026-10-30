package auth

import (
	"crypto/rand"
	"fmt"
)

const (
	RecoveryCodeLength = 10
	RecoveryCodesCount = 8
)

const recoveryCodeCharset = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// GenerateRecoveryCodes returns a batch of single-use backup codes shown to
// the user when enabling 2FA. Only their hashes are persisted.
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := range codes {
		buf := make([]byte, RecoveryCodeLength)
		if _, err := rand.Read(buf); err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}
		for j, b := range buf {
			buf[j] = recoveryCodeCharset[int(b)%len(recoveryCodeCharset)]
		}
		codes[i] = string(buf)
	}
	return codes, nil
}
