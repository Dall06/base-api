package auth_test

import (
	"github.com/diegoaleon/test-app/pkg/auth"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "success - generates valid bcrypt hash for normal password",
			password: "MySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "success - generates hash for short password",
			password: "pass",
			wantErr:  false,
		},
		{
			name:     "success - generates hash for long password",
			password: strings.Repeat("a", 70),
			wantErr:  false,
		},
		{
			name:     "error - password exceeds bcrypt limit",
			password: strings.Repeat("a", 100),
			wantErr:  true,
		},
		{
			name:     "success - generates hash for password with special characters",
			password: "P@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "success - generates hash for empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "success - generates hash for unicode password",
			password: "パスワード123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := auth.HashPassword(tt.password)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				// Verify hash is not empty
				if hash == "" {
					t.Error("expected hash, got empty string")
				}

				// Verify it's a valid bcrypt hash (starts with $2a$ or $2b$)
				if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
					t.Errorf("hash doesn't appear to be bcrypt format: %s", hash)
				}

				// Verify the hash can be used to compare against the original password
				err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
				if err != nil {
					t.Errorf("generated hash doesn't match original password: %v", err)
				}

				// Verify that generating the same password twice produces different hashes (salt test)
				hash2, err := auth.HashPassword(tt.password)
				if err != nil {
					t.Fatalf("failed to generate second hash: %v", err)
				}
				if hash == hash2 {
					t.Error("expected different hashes for same password (salt should differ)")
				}
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	// Pre-generate some hashes for testing
	validPassword := "MySecurePassword123"
	validHash, _ := auth.HashPassword(validPassword)

	emptyPassword := ""
	emptyHash, _ := auth.HashPassword(emptyPassword)

	specialPassword := "P@ssw0rd!#$%"
	specialHash, _ := auth.HashPassword(specialPassword)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "success - correct password returns true",
			password: validPassword,
			hash:     validHash,
			want:     true,
		},
		{
			name:     "failure - incorrect password returns false",
			password: "WrongPassword",
			hash:     validHash,
			want:     false,
		},
		{
			name:     "failure - empty password against valid hash returns false",
			password: "",
			hash:     validHash,
			want:     false,
		},
		{
			name:     "success - empty password against empty hash returns true",
			password: emptyPassword,
			hash:     emptyHash,
			want:     true,
		},
		{
			name:     "failure - valid password against empty hash returns false",
			password: validPassword,
			hash:     "",
			want:     false,
		},
		{
			name:     "failure - password against invalid hash returns false",
			password: "password",
			hash:     "not-a-valid-bcrypt-hash",
			want:     false,
		},
		{
			name:     "success - password with special characters",
			password: specialPassword,
			hash:     specialHash,
			want:     true,
		},
		{
			name:     "failure - similar but different password",
			password: "MySecurePassword124", // Last char different
			hash:     validHash,
			want:     false,
		},
		{
			name:     "failure - case sensitive password check",
			password: "mysecurepassword123", // Different case
			hash:     validHash,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.CheckPassword(tt.password, tt.hash)

			if got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	// Test that the same password produces verifiable but different hashes
	password := "TestPassword123"

	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password (1): %v", err)
	}

	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password (2): %v", err)
	}

	// Hashes should be different (due to different salts)
	if hash1 == hash2 {
		t.Error("expected different hashes, got identical hashes")
	}

	// Both hashes should validate the original password
	if !auth.CheckPassword(password, hash1) {
		t.Error("hash1 failed to validate original password")
	}
	if !auth.CheckPassword(password, hash2) {
		t.Error("hash2 failed to validate original password")
	}
}

func TestCheckPasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "empty password and empty hash",
			password: "",
			hash:     "",
			want:     false,
		},
		{
			name:     "whitespace password",
			password: "   ",
			hash:     "",
			want:     false,
		},
		{
			name:     "null bytes in password",
			password: "pass\x00word",
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.CheckPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPasswordCost(t *testing.T) {
	// Verify that the hash uses the expected bcrypt cost
	password := "TestPassword"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Extract cost from hash
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("failed to extract cost from hash: %v", err)
	}

	expectedCost := 12 // From the source code
	if cost != expectedCost {
		t.Errorf("bcrypt cost = %d, want %d", cost, expectedCost)
	}
}
