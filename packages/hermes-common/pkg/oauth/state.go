package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type StateCodec struct {
	gcm cipher.AEAD
}

// Creates a StateCodec from a 32-byte encryption key.
func NewStateCodec(key []byte) (*StateCodec, error) {
	if len(key) != 32 {
		return nil, errors.New("state encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("state cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("state gcm: %w", err)
	}
	return &StateCodec{gcm: gcm}, nil
}

// Encrypts userID + provider + timestamp into a URL-safe state string.
func (sc *StateCodec) Encode(userID, provider string) (string, error) {
	plaintext := fmt.Sprintf("%s|%s|%d", userID, provider, time.Now().Unix())

	nonce := make([]byte, sc.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate state nonce: %w", err)
	}
	ciphertext := sc.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypts and validates a state string. Returns (userID, provider, error).
// Rejects states older than maxAge.
func (sc *StateCodec) Decode(encoded string, maxAge time.Duration) (string, string, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("decode state base64: %w", err)
	}
	nonceSize := sc.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", "", errors.New("state ciphertext too short")
	}
	plaintext, err := sc.gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", "", fmt.Errorf("decrypt state: %w", err)
	}

	parts := strings.SplitN(string(plaintext), "|", 3)
	if len(parts) != 3 {
		return "", "", errors.New("malformed state payload")
	}

	var ts int64
	if _, err := fmt.Sscanf(parts[2], "%d", &ts); err != nil {
		return "", "", errors.New("invalid state timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > maxAge {
		return "", "", errors.New("state expired")
	}

	return parts[0], parts[1], nil
}
