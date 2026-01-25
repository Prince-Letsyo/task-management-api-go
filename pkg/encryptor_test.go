package pkg

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestEncryptor(t *testing.T) {
	key := "6368616e676520746869732070617373776f726420746f206120736563726574" // 32 bytes hex
	// 32 bytes = 64 hex chars
	// "change this password to a secret" in hex

	e, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	t.Run("Encrypt and Decrypt successfully", func(t *testing.T) {
		payload := "user@example.com"
		token, err := e.Encrypt(payload, time.Minute)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		decrypted, err := e.Decrypt(token)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}

		if decrypted != payload {
			t.Errorf("Expected payload %s, got %s", payload, decrypted)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		payload := "expired@example.com"
		// Encrypt with negative duration to force expiration
		token, err := e.Encrypt(payload, -time.Minute)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		_, err = e.Decrypt(token)
		if err == nil {
			t.Error("Expected error for expired token, got nil")
		}
		if err.Error() != "token has expired" {
			t.Errorf("Expected 'token has expired' error, got: %v", err)
		}
	})

	t.Run("Invalid Key Hex", func(t *testing.T) {
		_, err := NewEncryptor("invalid-hex")
		if err == nil {
			t.Error("Expected error for invalid key hex, got nil")
		}
	})

	t.Run("Invalid Token Hex", func(t *testing.T) {
		_, err := e.Decrypt("invalid-token-hex")
		if err == nil {
			t.Error("Expected error for invalid token hex, got nil")
		}
	})

	t.Run("Tampered Token", func(t *testing.T) {
		payload := "tamper@example.com"
		token, err := e.Encrypt(payload, time.Minute)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		// Decode, tamper, encode
		tokenBytes, _ := hex.DecodeString(token)
		tokenBytes[len(tokenBytes)-1] ^= 0x01 // Flip last bit
		tamperedToken := hex.EncodeToString(tokenBytes)

		_, err = e.Decrypt(tamperedToken)
		if err == nil {
			t.Error("Expected error for tampered token, got nil")
		}
	})
}
