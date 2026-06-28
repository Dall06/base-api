package auth_test

import (
	"github.com/diegoaleon/test-app/pkg/auth"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name          string
		secretKey     string
		tokenDuration time.Duration
		wantNil       bool
	}{
		{
			name:          "success - creates manager with valid params",
			secretKey:     "my-secret-key",
			tokenDuration: time.Hour,
			wantNil:       false,
		},
		{
			name:          "success - creates manager with short secret",
			secretKey:     "short",
			tokenDuration: time.Minute,
			wantNil:       false,
		},
		{
			name:          "success - creates manager with long expiration",
			secretKey:     "secret-key-for-testing",
			tokenDuration: 24 * time.Hour,
			wantNil:       false,
		},
		{
			name:          "success - creates manager with empty secret",
			secretKey:     "",
			tokenDuration: time.Hour,
			wantNil:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := auth.NewJWTManager(tt.secretKey, tt.tokenDuration)

			if tt.wantNil && manager != nil {
				t.Error("expected nil manager, got non-nil")
			}
			if !tt.wantNil && manager == nil {
				t.Error("expected non-nil manager, got nil")
			}
		})
	}
}

func TestJWTManager_Generate(t *testing.T) {
	tests := []struct {
		name      string
		staffID   string
		companyID string
		email     string
		role      string
		duration  time.Duration
		wantErr   bool
	}{
		{
			name:      "success - generates valid token with all fields",
			staffID:   "staff-123",
			companyID: "company-456",
			email:     "user@example.com",
			role:      "admin",
			duration:  time.Hour,
			wantErr:   false,
		},
		{
			name:      "success - generates token with empty role",
			staffID:   "staff-789",
			companyID: "company-012",
			email:     "test@test.com",
			role:      "",
			duration:  time.Hour,
			wantErr:   false,
		},
		{
			name:      "success - generates token with empty email",
			staffID:   "staff-abc",
			companyID: "company-def",
			email:     "",
			role:      "user",
			duration:  time.Hour,
			wantErr:   false,
		},
		{
			name:      "success - generates token with short expiration",
			staffID:   "staff-111",
			companyID: "company-222",
			email:     "short@test.com",
			role:      "user",
			duration:  time.Second,
			wantErr:   false,
		},
		{
			name:      "success - generates token with long expiration",
			staffID:   "staff-333",
			companyID: "company-444",
			email:     "long@test.com",
			role:      "admin",
			duration:  365 * 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "success - generates token with special characters in email",
			staffID:   "staff-555",
			companyID: "company-666",
			email:     "user+tag@example.co.uk",
			role:      "manager",
			duration:  time.Hour,
			wantErr:   false,
		},
		{
			name:      "success - generates token with unicode in role",
			staffID:   "staff-777",
			companyID: "company-888",
			email:     "unicode@test.com",
			role:      "administrador",
			duration:  time.Hour,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := auth.NewJWTManager("test-secret-key", tt.duration)

			token, expiresAt, err := manager.Generate(tt.staffID, tt.companyID, tt.email, tt.role)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				// Verify token is not empty
				if token == "" {
					t.Error("expected token, got empty string")
				}

				// Verify token has 3 parts (header.payload.signature)
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("token should have 3 parts, got %d", len(parts))
				}

				// Verify expiration time is in the future
				if expiresAt.IsZero() {
					t.Error("expected valid expiration time, got zero time")
				}
				if !expiresAt.After(time.Now()) {
					t.Error("expiration time should be in the future")
				}

				// Verify expiration is roughly correct (within 1 second tolerance)
				expectedExpiration := time.Now().Add(tt.duration)
				diff := expiresAt.Sub(expectedExpiration).Abs()
				if diff > time.Second {
					t.Errorf("expiration time differs by %v, expected ~%v", diff, tt.duration)
				}
			}
		})
	}
}

func TestJWTManager_Validate(t *testing.T) {
	secret := "test-secret-key-for-validation"
	duration := time.Hour
	manager := auth.NewJWTManager(secret, duration)

	// Generate a valid token for testing
	validToken, _, _ := manager.Generate("staff-123", "company-456", "test@example.com", "admin")

	// Generate an expired token
	expiredManager := auth.NewJWTManager(secret, -time.Hour)
	expiredToken, _, _ := expiredManager.Generate("staff-expired", "company-exp", "expired@test.com", "user")

	// Generate token with different secret
	wrongManager := auth.NewJWTManager("different-secret-key", duration)
	wrongSecretToken, _, _ := wrongManager.Generate("staff-wrong", "company-wrong", "wrong@test.com", "user")

	tests := []struct {
		name          string
		token         string
		wantErr       bool
		wantErrType   error
		wantStaffID   string
		wantCompanyID string
		wantEmail     string
		wantRole      string
	}{
		{
			name:          "success - validates valid token",
			token:         validToken,
			wantErr:       false,
			wantStaffID:   "staff-123",
			wantCompanyID: "company-456",
			wantEmail:     "test@example.com",
			wantRole:      "admin",
		},
		{
			name:        "error - expired token returns ErrExpiredToken",
			token:       expiredToken,
			wantErr:     true,
			wantErrType: auth.ErrExpiredToken,
		},
		{
			name:        "error - invalid signature returns ErrInvalidToken",
			token:       wrongSecretToken,
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
		{
			name:        "error - malformed token returns ErrInvalidToken",
			token:       "not.a.valid.token.structure",
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
		{
			name:        "error - empty token returns ErrInvalidToken",
			token:       "",
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
		{
			name:        "error - random string returns ErrInvalidToken",
			token:       "completely-random-invalid-string",
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
		{
			name:        "error - token with only header returns ErrInvalidToken",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
		{
			name:        "error - token with two parts returns ErrInvalidToken",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			wantErr:     true,
			wantErrType: auth.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.Validate(tt.token)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantErr && tt.wantErrType != nil {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
				}
			}

			if !tt.wantErr {
				if claims == nil {
					t.Fatal("expected claims, got nil")
				}

				if claims.StaffID != tt.wantStaffID {
					t.Errorf("StaffID = %v, want %v", claims.StaffID, tt.wantStaffID)
				}
				if claims.CompanyID != tt.wantCompanyID {
					t.Errorf("CompanyID = %v, want %v", claims.CompanyID, tt.wantCompanyID)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("Email = %v, want %v", claims.Email, tt.wantEmail)
				}
				if claims.Role != tt.wantRole {
					t.Errorf("Role = %v, want %v", claims.Role, tt.wantRole)
				}

				// Verify RegisteredClaims
				if claims.Subject != tt.wantStaffID {
					t.Errorf("Subject = %v, want %v", claims.Subject, tt.wantStaffID)
				}
				if claims.ExpiresAt == nil {
					t.Error("expected ExpiresAt, got nil")
				}
				if claims.IssuedAt == nil {
					t.Error("expected IssuedAt, got nil")
				}
			}
		})
	}
}

func TestJWTManager_GenerateAndValidateRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		expiration time.Duration
		staffID    string
		companyID  string
		email      string
		role       string
	}{
		{
			name:       "round trip - admin role",
			secret:     "secret-for-admin",
			expiration: time.Hour,
			staffID:    "staff-admin",
			companyID:  "company-001",
			email:      "admin@example.com",
			role:       "admin",
		},
		{
			name:       "round trip - user role",
			secret:     "secret-for-user",
			expiration: 30 * time.Minute,
			staffID:    "staff-user",
			companyID:  "company-002",
			email:      "user@example.com",
			role:       "user",
		},
		{
			name:       "round trip - empty role",
			secret:     "secret-for-empty",
			expiration: 2 * time.Hour,
			staffID:    "staff-empty",
			companyID:  "company-003",
			email:      "empty@example.com",
			role:       "",
		},
		{
			name:       "round trip - long expiration",
			secret:     "secret-long-exp",
			expiration: 24 * time.Hour,
			staffID:    "staff-long",
			companyID:  "company-long",
			email:      "long@test.com",
			role:       "manager",
		},
		{
			name:       "round trip - unicode characters",
			secret:     "secret-unicode",
			expiration: time.Hour,
			staffID:    "staff-日本",
			companyID:  "company-中国",
			email:      "unicode@テスト.com",
			role:       "administrador",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := auth.NewJWTManager(tt.secret, tt.expiration)

			// Generate token
			token, expiresAt, err := manager.Generate(tt.staffID, tt.companyID, tt.email, tt.role)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Validate token immediately
			claims, err := manager.Validate(token)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Verify all claims match
			if claims.StaffID != tt.staffID {
				t.Errorf("StaffID = %v, want %v", claims.StaffID, tt.staffID)
			}
			if claims.CompanyID != tt.companyID {
				t.Errorf("CompanyID = %v, want %v", claims.CompanyID, tt.companyID)
			}
			if claims.Email != tt.email {
				t.Errorf("Email = %v, want %v", claims.Email, tt.email)
			}
			if claims.Role != tt.role {
				t.Errorf("Role = %v, want %v", claims.Role, tt.role)
			}

			// Verify expiration matches
			claimsExpiry := claims.ExpiresAt.Time
			diff := claimsExpiry.Sub(expiresAt).Abs()
			if diff > time.Second {
				t.Errorf("expiration mismatch: diff = %v", diff)
			}
		})
	}
}

func TestJWTManager_InvalidSigningMethod(t *testing.T) {
	// Create a token with RS256 (RSA) instead of HS256 (HMAC)
	// This should fail validation because JWTManager expects HMAC
	claims := auth.Claims{
		StaffID:   "staff-test",
		CompanyID: "company-test",
		Email:     "test@example.com",
		Role:      "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "staff-test",
		},
	}

	// Create token with None algorithm (which should fail)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	manager := auth.NewJWTManager("test-secret", time.Hour)
	_, err = manager.Validate(tokenString)
	if err == nil {
		t.Error("expected error for invalid signing method, got nil")
	}
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTManager_DifferentSecrets(t *testing.T) {
	// Test that tokens signed with different secrets cannot be validated
	secret1 := "first-secret-key"
	secret2 := "second-secret-key"

	manager1 := auth.NewJWTManager(secret1, time.Hour)
	manager2 := auth.NewJWTManager(secret2, time.Hour)

	token1, _, err := manager1.Generate("staff-1", "company-1", "test1@example.com", "admin")
	if err != nil {
		t.Fatalf("failed to generate token with manager1: %v", err)
	}

	token2, _, err := manager2.Generate("staff-2", "company-2", "test2@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate token with manager2: %v", err)
	}

	// manager1 should not validate token2
	_, err = manager1.Validate(token2)
	if err == nil {
		t.Error("manager1 should not validate token from manager2")
	}
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// manager2 should not validate token1
	_, err = manager2.Validate(token1)
	if err == nil {
		t.Error("manager2 should not validate token from manager1")
	}
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// Each manager should validate its own token
	_, err = manager1.Validate(token1)
	if err != nil {
		t.Errorf("manager1 should validate its own token: %v", err)
	}

	_, err = manager2.Validate(token2)
	if err != nil {
		t.Errorf("manager2 should validate its own token: %v", err)
	}
}

func TestJWTManager_ClaimsStructure(t *testing.T) {
	// Test that generated tokens have the correct structure
	manager := auth.NewJWTManager("test-secret", time.Hour)

	token, _, err := manager.Generate("staff-123", "company-456", "test@example.com", "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := manager.Validate(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	// Verify all custom claims are present
	if claims.StaffID == "" {
		t.Error("expected StaffID to be set")
	}
	if claims.CompanyID == "" {
		t.Error("expected CompanyID to be set")
	}
	if claims.Email == "" {
		t.Error("expected Email to be set")
	}
	// Role can be empty, so we don't check it

	// Verify registered claims
	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}
	if claims.Subject == "" {
		t.Error("expected Subject to be set")
	}
	if claims.Subject != claims.StaffID {
		t.Error("expected Subject to match StaffID")
	}
}

func TestJWTManager_ErrorTypes(t *testing.T) {
	// Verify that error types are correctly defined
	if auth.ErrInvalidToken == nil {
		t.Error("ErrInvalidToken should not be nil")
	}
	if auth.ErrExpiredToken == nil {
		t.Error("ErrExpiredToken should not be nil")
	}

	// Verify error messages
	if auth.ErrInvalidToken.Error() != "invalid token" {
		t.Errorf("ErrInvalidToken message = %v, want 'invalid token'", auth.ErrInvalidToken.Error())
	}
	if auth.ErrExpiredToken.Error() != "token has expired" {
		t.Errorf("ErrExpiredToken message = %v, want 'token has expired'", auth.ErrExpiredToken.Error())
	}

	// Verify errors are distinct
	if errors.Is(auth.ErrInvalidToken, auth.ErrExpiredToken) {
		t.Error("ErrInvalidToken and ErrExpiredToken should be distinct")
	}
}
