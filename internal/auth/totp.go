package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const CurrentTOTPKeyVersion int = 1
const TotpIssuer string = "Go Boiler"

func GenerateTOTPSecret(accountEmail string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TotpIssuer,
		AccountName: accountEmail,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func ValidateTOTP(userSecret, code string) bool {
	return totp.Validate(code, userSecret)
}

func getEncryptionKey(keyVersion int) ([]byte, error) {
	var encoded string
	switch keyVersion {
	case 1:
		encoded = os.Getenv("MASTER_ENCRYPTION_KEY_V1")
	default:
		return nil, fmt.Errorf("unknown encryption key version: %d", keyVersion)
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key encoding: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	return key, nil
}

func TotpURLFromSecret(secret, accountEmail string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?algorithm=SHA1&digits=6&issuer=%s&period=30&secret=%s",
		url.PathEscape(TotpIssuer),
		url.PathEscape(accountEmail),
		url.PathEscape(TotpIssuer),
		url.PathEscape(secret),
	)
}

func EncryptOTPSecret(secret string, keyVersion int) (string, error) {
	key, err := getEncryptionKey(keyVersion)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptOTPSecret(ciphertextB64 string, keyVersion int) (string, error) {
	key, err := getEncryptionKey(keyVersion)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext encoding: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt (wrong key or tampered data): %w", err)
	}

	return string(plaintext), nil
}
