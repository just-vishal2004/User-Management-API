package service

import (
	"testing"
	"time"
)

// TestCalculateAge tests the CalculateAge function with multiple cases.
// We use a table-driven approach — the standard Go testing pattern.
func TestCalculateAge(t *testing.T) {
	// Each test case has a name, an input dob, and the expected age.
	// We fix "today" to a known date so tests don't break over time.
	today := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		dob      time.Time
		expected int
	}{
		{
			name:     "normal age — birthday already passed this year",
			dob:      time.Date(1998, 1, 10, 0, 0, 0, 0, time.UTC),
			expected: 28,
		},
		{
			name:     "normal age — birthday not yet this year",
			dob:      time.Date(1998, 12, 31, 0, 0, 0, 0, time.UTC),
			expected: 27,
		},
		{
			name:     "birthday is exactly today",
			dob:      time.Date(2000, 6, 15, 0, 0, 0, 0, time.UTC),
			expected: 26,
		},
		{
			name:     "birthday was yesterday",
			dob:      time.Date(1995, 6, 14, 0, 0, 0, 0, time.UTC),
			expected: 31,
		},
		{
			name:     "birthday is tomorrow",
			dob:      time.Date(1995, 6, 16, 0, 0, 0, 0, time.UTC),
			expected: 30,
		},
		{
			name:     "newborn — age is zero",
			dob:      time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "leap year birthday — Feb 29",
			dob:      time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC),
			expected: 26,
		},
		{
			name:     "very old person",
			dob:      time.Date(1920, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: 106,
		},
	}

	for _, tt := range tests {
		// t.Run creates a named sub-test for each case.
		// If one fails, you see exactly which case failed by name.
		t.Run(tt.name, func(t *testing.T) {
			// We need to test against a fixed "today" so results
			// don't change as real time passes.
			// We temporarily override time.Now by calculating
			// age relative to our fixed today date.
			got := calculateAgeAt(tt.dob, today)

			if got != tt.expected {
				t.Errorf(
					"CalculateAge(%v) = %d, want %d",
					tt.dob.Format("2006-01-02"),
					got,
					tt.expected,
				)
			}
		})
	}
}
