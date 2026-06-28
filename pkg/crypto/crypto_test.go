package crypto

import (
	"testing"
)

func TestNewFieldEncryptor(t *testing.T) {
	tests := []struct {
		name          string
		encryptionKey []byte
		hmacKey       []byte
		wantErr       error
	}{
		{
			name:          "valid keys",
			encryptionKey: make([]byte, 32),
			hmacKey:       make([]byte, 32),
			wantErr:       nil,
		},
		{
			name:          "encryption key too short",
			encryptionKey: make([]byte, 16),
			hmacKey:       make([]byte, 32),
			wantErr:       ErrInvalidKey,
		},
		{
			name:          "encryption key too long",
			encryptionKey: make([]byte, 64),
			hmacKey:       make([]byte, 32),
			wantErr:       ErrInvalidKey,
		},
		{
			name:          "hmac key too short",
			encryptionKey: make([]byte, 32),
			hmacKey:       make([]byte, 16),
			wantErr:       ErrInvalidHMACKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFieldEncryptor(tt.encryptionKey, tt.hmacKey)
			if err != tt.wantErr {
				t.Errorf("NewFieldEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"email", "test@example.com"},
		{"phone", "+52 555 123 4567"},
		{"unicode", "Juan Pérez García"},
		{"special chars", "test!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Encrypted should be different from plaintext (unless empty)
			if tt.plaintext != "" && encrypted == tt.plaintext {
				t.Error("Encrypt() returned plaintext unchanged")
			}

			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	encryptor := mustCreateEncryptor(t)
	plaintext := "same input"

	encrypted1, _ := encryptor.Encrypt(plaintext)
	encrypted2, _ := encryptor.Encrypt(plaintext)

	// Due to random nonce, same plaintext should produce different ciphertexts
	if encrypted1 == encrypted2 {
		t.Error("Encrypt() should produce different outputs for same input (due to random nonce)")
	}

	// But both should decrypt to the same value
	decrypted1, _ := encryptor.Decrypt(encrypted1)
	decrypted2, _ := encryptor.Decrypt(encrypted2)

	if decrypted1 != decrypted2 || decrypted1 != plaintext {
		t.Error("Both ciphertexts should decrypt to the original plaintext")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	tests := []struct {
		name       string
		ciphertext string
		wantErr    error
	}{
		{"invalid base64", "not-valid-base64!!!", ErrInvalidCipher},
		{"too short", "YWJj", ErrInvalidCipher}, // "abc" in base64, too short for nonce
		{"tampered", "dGFtcGVyZWRkYXRhdGhhdGlzbm90dmFsaWQ=", ErrDecryptionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if err != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlindIndex(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"email", "test@example.com"},
		{"phone", "+52 555 123 4567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index1 := encryptor.BlindIndex(tt.value)
			index2 := encryptor.BlindIndex(tt.value)

			// Same input should produce same index (deterministic)
			if index1 != index2 {
				t.Error("BlindIndex() should be deterministic")
			}

			// Empty value should return empty string
			if tt.value == "" && index1 != "" {
				t.Error("BlindIndex() should return empty string for empty input")
			}
		})
	}
}

func TestBlindIndexDifferentInputs(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	index1 := encryptor.BlindIndex("email1@test.com")
	index2 := encryptor.BlindIndex("email2@test.com")

	if index1 == index2 {
		t.Error("BlindIndex() should produce different outputs for different inputs")
	}
}

func TestBlindIndexDifferentKeys(t *testing.T) {
	key1 := []byte("key1key1key1key1key1key1key1key1")
	key2 := []byte("key2key2key2key2key2key2key2key2")
	aesKey := make([]byte, 32)

	enc1, _ := NewFieldEncryptor(aesKey, key1)
	enc2, _ := NewFieldEncryptor(aesKey, key2)

	input := "test@example.com"
	index1 := enc1.BlindIndex(input)
	index2 := enc2.BlindIndex(input)

	if index1 == index2 {
		t.Error("BlindIndex() with different keys should produce different outputs")
	}
}

func TestEncryptWithIndex(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	plaintext := "test@example.com"
	encrypted, index, err := encryptor.EncryptWithIndex(plaintext)
	if err != nil {
		t.Fatalf("EncryptWithIndex() error = %v", err)
	}

	// Verify encryption
	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}

	// Verify index matches standalone BlindIndex
	expectedIndex := encryptor.BlindIndex(plaintext)
	if index != expectedIndex {
		t.Errorf("EncryptWithIndex() index = %v, want %v", index, expectedIndex)
	}
}

func TestEncryptWithIndexEmpty(t *testing.T) {
	encryptor := mustCreateEncryptor(t)

	encrypted, index, err := encryptor.EncryptWithIndex("")
	if err != nil {
		t.Fatalf("EncryptWithIndex() error = %v", err)
	}

	if encrypted != "" || index != "" {
		t.Error("EncryptWithIndex() should return empty strings for empty input")
	}
}

func mustCreateEncryptor(t *testing.T) *FieldEncryptor {
	t.Helper()
	encryptionKey := []byte("12345678901234567890123456789012") // 32 bytes
	hmacKey := []byte("hmackey-hmackey-hmackey-hmackey!") // 32 bytes

	enc, err := NewFieldEncryptor(encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}
	return enc
}
