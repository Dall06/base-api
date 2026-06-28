// Package colombus provides common parsing utilities.
package colombus

import (
	"fmt"
	"time"
)

// ParseDate parses an optional date string (YYYY-MM-DD) into a time.Time.
// Returns fallback if the pointer is nil or empty.
func ParseDate(s *string, fallback time.Time) (time.Time, error) {
	if s == nil || *s == "" {
		return fallback, nil
	}
	parsed, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return parsed, nil
}

// ParseDateRequired parses a required date string (YYYY-MM-DD).
// Returns an error if the string is nil or empty.
func ParseDateRequired(s *string) (time.Time, error) {
	if s == nil || *s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	parsed, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return parsed, nil
}

// AddDuration adds a duration in days to a date, converting to calendar
// months when possible so "pay on the 4th, expires on the 4th" works.
// 30d=1mo, 60d=2mo, 90d=3mo, 180d=6mo, 365d=1yr.
func AddDuration(start time.Time, days int) time.Time {
	switch days {
	case 30:
		return start.AddDate(0, 1, 0)
	case 60:
		return start.AddDate(0, 2, 0)
	case 90:
		return start.AddDate(0, 3, 0)
	case 180:
		return start.AddDate(0, 6, 0)
	case 365:
		return start.AddDate(1, 0, 0)
	}
	return start.AddDate(0, 0, days)
}
