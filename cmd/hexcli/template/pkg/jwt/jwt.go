package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the standard JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Config holds JWT configuration
type Config struct {
	Secret     string
	Expiration time.Duration
}

// Generator handles JWT token generation
type Generator struct {
	config Config
}

// NewGenerator creates a new JWT generator
func NewGenerator(config Config) *Generator {
	return &Generator{config: config}
}

// GenerateInput contains the input for token generation
type GenerateInput struct {
	UserID string
	Email  string
	Role   string
}

// GenerateOutput contains the result of token generation
type GenerateOutput struct {
	Token     string
	JTI       string
	ExpiresAt time.Time
}

// Generate creates a new JWT token
func (g *Generator) Generate(input GenerateInput) (*GenerateOutput, error) {
	now := time.Now()
	expiresAt := now.Add(g.config.Expiration)
	jti := uuid.New().String()

	claims := Claims{
		UserID: input.UserID,
		Email:  input.Email,
		Role:   input.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.config.Secret))
	if err != nil {
		return nil, err
	}

	return &GenerateOutput{
		Token:     tokenString,
		JTI:       jti,
		ExpiresAt: expiresAt,
	}, nil
}

// Validate validates a JWT token and returns the claims
func (g *Generator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(g.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
