package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type TokenData struct {
	Payload   string `json:"p"`   // the original data you want to protect (e.g. user ID, email, etc.)
	ExpiresAt int64  `json:"exp"` // Unix timestamp when it expires
}

// Encrypt creates a time-limited encrypted token that expires in the given duration (e.g. 15 * time.Minute)
func Encrypt(payload string, keyString string, validity time.Duration) (encryptedString string, err error) {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("invalid key hex: %w", err)
	}

	data := TokenData{
		Payload:   payload,
		ExpiresAt: time.Now().Add(validity).Unix(),
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("json marshal failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

// Decrypt decrypts the token and checks if it has expired.
// Returns the original payload if valid, otherwise an error.
func Decrypt(encryptedString string, keyString string) (payload string, err error) {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("invalid key hex: %w", err)
	}

	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted hex: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (possibly tampered or wrong key): %w", err)
	}

	var data TokenData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return "", fmt.Errorf("invalid payload format: %w", err)
	}

	if time.Now().Unix() > data.ExpiresAt {
		return "", fmt.Errorf("token has expired")
	}

	return data.Payload, nil
}
