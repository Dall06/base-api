package validation

import "github.com/diegoaleon/test-app/pkg/errs"

// ValidatePlanRequest validates the common fields for plan creation and update
func ValidatePlanRequest(name string, price float64, inscriptionFee *float64, durationDays int) error {
	if name == "" {
		return errs.ValueError("plan name is required")
	}

	if price <= 0 {
		return errs.ValueError("invalid membership cost")
	}

	if inscriptionFee != nil && *inscriptionFee < 0 {
		return errs.ValueError("invalid inscription cost")
	}

	if durationDays <= 0 {
		return errs.ValueError("invalid membership period")
	}

	return nil
}
