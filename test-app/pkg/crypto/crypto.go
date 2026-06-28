// Package crypto provides AES-GCM encryption and HMAC-based blind indexing
// for sensitive data at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKey       = errors.New("crypto: invalid key length, must be 32 bytes for AES-256")
	ErrInvalidHMACKey   = errors.New("crypto: invalid HMAC key length, must be at least 32 bytes")
	ErrDecryptionFailed = errors.New("crypto: decryption failed")
	ErrInvalidCipher    = errors.New("crypto: invalid ciphertext")
)

// FieldEncryptor handles encryption/decryption and blind indexing for sensitive fields.
type FieldEncryptor struct {
	aesKey  []byte // 32 bytes for AES-256
	hmacKey []byte // for blind index generation
}

// NewFieldEncryptor creates a new encryptor with the given keys.
// encryptionKey must be 32 bytes for AES-256.
// hmacKey should be at least 32 bytes for secure blind indexing.
func NewFieldEncryptor(encryptionKey, hmacKey []byte) (*FieldEncryptor, error) {
	if len(encryptionKey) != 32 {
		return nil, ErrInvalidKey
	}
	if len(hmacKey) < 32 {
		return nil, ErrInvalidHMACKey
	}
	return &FieldEncryptor{
		aesKey:  encryptionKey,
		hmacKey: hmacKey,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns base64-encoded ciphertext (nonce prepended).
func (e *FieldEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM.
func (e *FieldEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCipher
	}

	block, err := aes.NewCipher(e.aesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCipher
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// BlindIndex generates a deterministic HMAC-SHA256 hash for searchable encryption.
// The hash is base64-encoded and can be used for exact-match queries.
func (e *FieldEncryptor) BlindIndex(value string) string {
	if value == "" {
		return ""
	}

	h := hmac.New(sha256.New, e.hmacKey)
	h.Write([]byte(value))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// EncryptWithIndex encrypts a value and generates its blind index.
// Useful for fields that need both encryption and searchability.
func (e *FieldEncryptor) EncryptWithIndex(plaintext string) (encrypted, index string, err error) {
	encrypted, err = e.Encrypt(plaintext)
	if err != nil {
		return "", "", err
	}
	index = e.BlindIndex(plaintext)
	return encrypted, index, nil
}
