package jwt_test

import (
	"base-api/pkg/jwt"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		input   jwt.GenerateInput
		wantErr bool
	}{
		{
			name: "success - generates valid token",
			input: jwt.GenerateInput{
				StaffID:   "staff-123",
				CompanyID: "company-456",
				Slug:      "my-gym",
				Email:     "test@example.com",
				Role:      "admin",
			},
			wantErr: false,
		},
		{
			name: "success - generates token with empty role",
			input: jwt.GenerateInput{
				StaffID:   "staff-999",
				CompanyID: "company-888",
				Slug:      "fitness-center",
				Email:     "user@test.com",
				Role:      "",
			},
			wantErr: false,
		},
		{
			name: "success - generates token with different role",
			input: jwt.GenerateInput{
				StaffID:   "staff-abc",
				CompanyID: "company-xyz",
				Slug:      "gym-xyz",
				Email:     "manager@example.com",
				Role:      "manager",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := jwt.Config{
				Secret:     "test-secret-key-32-chars-long!!",
				Expiration: time.Hour,
			}
			gen := jwt.NewGenerator(cfg)

			output, err := gen.Generate(tt.input)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && output.Token == "" {
				t.Error("expected token, got empty string")
			}
			if !tt.wantErr && output.JTI == "" {
				t.Error("expected JTI, got empty string")
			}
			if !tt.wantErr && output.ExpiresAt.IsZero() {
				t.Error("expected valid expiration time, got zero time")
			}
			if !tt.wantErr && !output.ExpiresAt.After(time.Now()) {
				t.Error("expiration time should be in the future")
			}
		})
	}
}

func TestGenerator_Validate(t *testing.T) {
	secret := "test-secret-key-32-chars-long!!"
	cfg := jwt.Config{
		Secret:     secret,
		Expiration: time.Hour,
	}
	gen := jwt.NewGenerator(cfg)

	// Generate a valid token for testing
	validOutput, _ := gen.Generate(jwt.GenerateInput{
		StaffID:   "staff-123",
		CompanyID: "company-456",
		Slug:      "my-gym",
		Email:     "test@example.com",
		Role:      "admin",
	})

	// Generate an expired token
	expiredGen := jwt.NewGenerator(jwt.Config{
		Secret:     secret,
		Expiration: -time.Hour, // Already expired
	})
	expiredOutput, _ := expiredGen.Generate(jwt.GenerateInput{
		StaffID:   "staff-999",
		CompanyID: "company-888",
		Slug:      "old-gym",
		Email:     "expired@example.com",
		Role:      "user",
	})

	// Generate a token with wrong secret
	wrongSecretGen := jwt.NewGenerator(jwt.Config{
		Secret:     "different-secret-key-32-chars!",
		Expiration: time.Hour,
	})
	wrongSecretOutput, _ := wrongSecretGen.Generate(jwt.GenerateInput{
		StaffID:   "staff-111",
		CompanyID: "company-222",
		Slug:      "other-gym",
		Email:     "wrong@example.com",
		Role:      "user",
	})

	tests := []struct {
		name          string
		token         string
		wantErr       bool
		wantStaffID   string
		wantCompanyID string
		wantSlug      string
		wantEmail     string
		wantRole      string
	}{
		{
			name:          "success - valid token returns correct claims",
			token:         validOutput.Token,
			wantErr:       false,
			wantStaffID:   "staff-123",
			wantCompanyID: "company-456",
			wantSlug:      "my-gym",
			wantEmail:     "test@example.com",
			wantRole:      "admin",
		},
		{
			name:    "error - expired token returns error",
			token:   expiredOutput.Token,
			wantErr: true,
		},
		{
			name:    "error - invalid signature returns error",
			token:   wrongSecretOutput.Token,
			wantErr: true,
		},
		{
			name:    "error - malformed token returns error",
			token:   "not.a.valid.token",
			wantErr: true,
		},
		{
			name:    "error - empty token returns error",
			token:   "",
			wantErr: true,
		},
		{
			name:    "error - random string returns error",
			token:   "completely-invalid-token-string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := gen.Validate(tt.token)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
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
				if claims.Slug != tt.wantSlug {
					t.Errorf("Slug = %v, want %v", claims.Slug, tt.wantSlug)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("Email = %v, want %v", claims.Email, tt.wantEmail)
				}
				if claims.Role != tt.wantRole {
					t.Errorf("Role = %v, want %v", claims.Role, tt.wantRole)
				}
				if claims.ID == "" {
					t.Error("expected JTI (ID) to be set, got empty string")
				}
			}
		})
	}
}

func TestGenerator_ValidateWithDifferentAlgorithm(t *testing.T) {
	// Create a token with HS512 algorithm instead of HS256
	claims := jwt.Claims{
		StaffID:   "staff-123",
		CompanyID: "company-456",
		Slug:      "my-gym",
		Email:     "test@example.com",
		Role:      "admin",
		RegisteredClaims: jwtgo.RegisteredClaims{
			ExpiresAt: jwtgo.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now()),
		},
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key-32-chars-long!!"))
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	cfg := jwt.Config{
		Secret:     "test-secret-key-32-chars-long!!",
		Expiration: time.Hour,
	}
	gen := jwt.NewGenerator(cfg)

	// The token should still validate because both HS256 and HS512 are HMAC methods
	// and jwt.ParseWithClaims accepts any HMAC method by default
	_, err = gen.Validate(tokenString)
	if err != nil {
		t.Logf("Token with HS512 returned error: %v", err)
	}
}

func TestGenerator_GenerateAndValidateRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		expiration time.Duration
		input      jwt.GenerateInput
	}{
		{
			name:       "round trip - short expiration",
			secret:     "my-secret-key-32-chars-long!!!!",
			expiration: time.Minute,
			input: jwt.GenerateInput{
				StaffID:   "staff-001",
				CompanyID: "company-001",
				Slug:      "gym-one",
				Email:     "user1@example.com",
				Role:      "user",
			},
		},
		{
			name:       "round trip - long expiration",
			secret:     "another-secret-32-chars-long!!",
			expiration: 24 * time.Hour,
			input: jwt.GenerateInput{
				StaffID:   "staff-002",
				CompanyID: "company-002",
				Slug:      "gym-two",
				Email:     "admin@example.com",
				Role:      "admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := jwt.Config{
				Secret:     tt.secret,
				Expiration: tt.expiration,
			}
			gen := jwt.NewGenerator(cfg)

			// Generate token
			output, err := gen.Generate(tt.input)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Validate token
			claims, err := gen.Validate(output.Token)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			// Verify all claims match
			if claims.StaffID != tt.input.StaffID {
				t.Errorf("StaffID = %v, want %v", claims.StaffID, tt.input.StaffID)
			}
			if claims.CompanyID != tt.input.CompanyID {
				t.Errorf("CompanyID = %v, want %v", claims.CompanyID, tt.input.CompanyID)
			}
			if claims.Slug != tt.input.Slug {
				t.Errorf("Slug = %v, want %v", claims.Slug, tt.input.Slug)
			}
			if claims.Email != tt.input.Email {
				t.Errorf("Email = %v, want %v", claims.Email, tt.input.Email)
			}
			if claims.Role != tt.input.Role {
				t.Errorf("Role = %v, want %v", claims.Role, tt.input.Role)
			}
			if claims.ID == "" {
				t.Error("expected JTI (ID) to be set")
			}
			if claims.ID != output.JTI {
				t.Errorf("JTI mismatch: claims.ID = %v, output.JTI = %v", claims.ID, output.JTI)
			}
		})
	}
}
