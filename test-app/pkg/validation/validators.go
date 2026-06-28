package validation

import "regexp"

var (
	// emailRegex is a compiled regex for email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// slugRegex is a compiled regex for slug validation
	// Must be lowercase alphanumeric with hyphens, cannot start/end with hyphen
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// IsValidEmail validates email format using a basic regex pattern.
// Returns true if the email matches the expected format.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidSlug validates slug format.
// A valid slug must be 3-50 characters long, contain only lowercase
// alphanumeric characters and hyphens, and cannot start or end with a hyphen.
func IsValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}
	return slugRegex.MatchString(slug)
}
