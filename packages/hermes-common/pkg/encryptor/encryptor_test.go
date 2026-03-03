package encryptor

import (
	"strings"
	"testing"
)

// Verifies that a 32 byte key produces a working Encryptor
// Happy-path constructor test
func TestNewEncryptor_ValidKey(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	enc, err := NewEncryptor(key)

	// Expecting a non nil encryptor and no error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if enc == nil {
		t.Fatalf("expected non-nil encryptor")
	}
}

// Ensures the constructor rejects keys that aren't exactly 32 bytes long
func TestNewEncryptor_InvalidKeyLength(t *testing.T) {

	//Expecting every one of these to fail
	cases := []struct {
		name string
		key  []byte
	}{
		{"too_short_16", []byte("0123456789012345")},
		{"too_short_1", []byte("a")},
		{"too_long", []byte("01234567890123456789012345678901abcde")},
		{"empty", []byte{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewEncryptor(tc.key)
			if err == nil {
				t.Errorf("expected error for key length %d, got nil", len(tc.key))
			}
		})
	}
}

// Ensures Encrypt followed by Decrypt returns the original plaintext
func TestEncryptorDecrypt_RoundTrip(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	inputs := []string{
		"hello world",
		"",
		"damn-bruh-thisguystinks123",
		"emoji: 🥞😅",
		strings.Repeat("x", 10000),
	}
	for _, plaintext := range inputs {
		ciphertext, err := enc.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encrypt(%q) failed: %v", plaintext, err)
		}
		if len(plaintext) > 0 && ciphertext == plaintext {
			t.Errorf("ciphertext should differ from plaintext")
		}
		decrypted, err := enc.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("decrypt fialed: %v", err)
		}
		if decrypted != plaintext {
			t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
		}
	}
}

// Ensures ciphertext encrypted with one key cant be decrypted with another key
func TestDecrypt_WrongKey(t *testing.T) {
	key1 := []byte("01234567890123456789012345678901")
	key2 := []byte("012345678901234567890123456789ab")

	enc1, _ := NewEncryptor(key1)
	enc2, _ := NewEncryptor(key2)
	ciphertext, _ := enc1.Encrypt("yo pierre")
	_, err := enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected decryption error with wrong key, got nil")
	}
}
