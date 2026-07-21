package crypto_test

import (
	"testing"

	"base-api/pkg/crypto"
)

func TestPasswordHashing(t *testing.T) {
	password := "my-secret-password"

	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to not be empty")
	}

	if !crypto.CheckPassword(hash, password) {
		t.Fatal("expected password check to pass")
	}

	if crypto.CheckPassword(hash, "wrong-password") {
		t.Fatal("expected password check to fail with wrong password")
	}
}
