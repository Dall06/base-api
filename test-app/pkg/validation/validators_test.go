package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid simple", "test@example.com", true},
		{"valid with subdomain", "test@mail.example.com", true},
		{"valid with plus", "test+tag@example.com", true},
		{"valid with dots", "test.name@example.com", true},
		{"valid with numbers", "user123@domain456.com", true},
		{"valid with hyphen in domain", "test@my-domain.com", true},
		{"valid with underscore", "test_user@example.com", true},
		{"invalid no @", "testexample.com", false},
		{"invalid double @", "test@@example.com", false},
		{"invalid no domain", "test@", false},
		{"invalid no local", "@example.com", false},
		{"invalid no extension", "test@example", false},
		{"invalid spaces", "test @example.com", false},
		{"invalid special chars", "test!user@example.com", false},
		{"empty string", "", false},
		{"only spaces", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			assert.Equal(t, tt.want, got, "IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
		})
	}
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want bool
	}{
		{"valid simple", "my-gym", true},
		{"valid min length 3", "abc", true},
		{"valid max length 50", strings.Repeat("a", 50), true},
		{"valid numbers", "gym123", true},
		{"valid hyphen middle", "my-gym-2024", true},
		{"valid multiple hyphens", "my-super-cool-gym", true},
		{"valid starts with number", "123gym", true},
		{"valid all lowercase", "mygym", true},
		{"too short 2 chars", "ab", false},
		{"too short 1 char", "a", false},
		{"too long 51 chars", strings.Repeat("a", 51), false},
		{"too long 100 chars", strings.Repeat("a", 100), false},
		{"invalid uppercase", "My-Gym", false},
		{"invalid all uppercase", "MYGYM", false},
		{"invalid underscore", "my_gym", false},
		{"invalid special chars", "my-gym!", false},
		{"invalid space", "my gym", false},
		{"invalid dot", "my.gym", false},
		{"invalid starts with hyphen", "-my-gym", false},
		{"invalid ends with hyphen", "my-gym-", false},
		{"invalid consecutive hyphens", "my--gym", false},
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"only hyphen", "-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidSlug(tt.slug)
			assert.Equal(t, tt.want, got, "IsValidSlug(%q) = %v, want %v", tt.slug, got, tt.want)
		})
	}
}

// TestIsValidSlugEdgeCases tests additional edge cases for slug validation
func TestIsValidSlugEdgeCases(t *testing.T) {
	t.Run("exact min length boundary", func(t *testing.T) {
		assert.True(t, IsValidSlug("abc"), "3 characters should be valid")
		assert.False(t, IsValidSlug("ab"), "2 characters should be invalid")
	})

	t.Run("exact max length boundary", func(t *testing.T) {
		slug50 := strings.Repeat("a", 50)
		slug51 := strings.Repeat("a", 51)
		assert.True(t, IsValidSlug(slug50), "50 characters should be valid")
		assert.False(t, IsValidSlug(slug51), "51 characters should be invalid")
	})

	t.Run("complex valid slugs", func(t *testing.T) {
		validSlugs := []string{
			"gym-24-7",
			"cross-fit-2024",
			"a1b2c3",
			"the-best-gym-in-town",
		}
		for _, slug := range validSlugs {
			assert.True(t, IsValidSlug(slug), "slug %q should be valid", slug)
		}
	})

	t.Run("complex invalid slugs", func(t *testing.T) {
		invalidSlugs := []string{
			"Gym-24-7",  // uppercase
			"gym_24_7",  // underscore
			"-gym-24-7", // starts with hyphen
			"gym-24-7-", // ends with hyphen
			"gym--24",   // consecutive hyphens
			"gym@24",    // special char
			"gym 24",    // space
			"gy",        // too short
		}
		for _, slug := range invalidSlugs {
			assert.False(t, IsValidSlug(slug), "slug %q should be invalid", slug)
		}
	})
}

// TestIsValidEmailEdgeCases tests additional edge cases for email validation
func TestIsValidEmailEdgeCases(t *testing.T) {
	t.Run("complex valid emails", func(t *testing.T) {
		validEmails := []string{
			"user+tag@example.com",
			"user.name@example.com",
			"user_name@example.com",
			"user123@example456.com",
			"a@b.co",
			"test@subdomain.example.com",
		}
		for _, email := range validEmails {
			assert.True(t, IsValidEmail(email), "email %q should be valid", email)
		}
	})

	t.Run("complex invalid emails", func(t *testing.T) {
		invalidEmails := []string{
			"plaintext",
			"@example.com",
			"user@",
			"user name@example.com",
			"user@domain",
			"user@@example.com",
			"",
		}
		for _, email := range invalidEmails {
			assert.False(t, IsValidEmail(email), "email %q should be invalid", email)
		}
	})
}
