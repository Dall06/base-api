package validation

import (
	"testing"
)

func ptr(f float64) *float64 { return &f }

func TestValidatePlanRequest(t *testing.T) {
	tests := []struct {
		name           string
		planName       string
		price          float64
		inscriptionFee *float64
		durationDays   int
		wantErr        bool
	}{
		{"valid plan", "Premium", 99.99, nil, 30, false},
		{"valid with inscription fee", "Gold", 149.99, ptr(50.0), 30, false},
		{"valid zero inscription fee", "Basic", 99.99, ptr(0), 30, false},
		{"empty name", "", 99.99, nil, 30, true},
		{"zero price", "Basic", 0, nil, 30, true},
		{"negative price", "Basic", -10, nil, 30, true},
		{"negative inscription fee", "Basic", 99, ptr(-5.0), 30, true},
		{"zero duration", "Basic", 99, nil, 0, true},
		{"negative duration", "Basic", 99, nil, -1, true},
		{"valid min duration", "Daily", 10, nil, 1, false},
		{"valid large price", "VIP", 9999.99, ptr(500.0), 365, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlanRequest(tt.planName, tt.price, tt.inscriptionFee, tt.durationDays)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlanRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
