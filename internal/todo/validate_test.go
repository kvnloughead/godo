package todo

import (
	"fmt"
	"testing"
)

func TestIsValidRFC3339(t *testing.T) {
	tests := []struct {
		description string
		time        string
		expected    bool
	}{
		{description: "Valid (UTC)", time: "2026-10-25T16:55:00Z", expected: true},
		{description: "Valid (With offset)", time: "2026-10-25T16:55:00-04:00", expected: true},
		{description: "Valid (With fractional seconds)", time: "2026-10-25T16:55:00.123456Z", expected: true},
		{description: "Invalid format", time: "2026/10/25 16:55:00", expected: false},
		{description: "Invalid string", time: "invalid-string", expected: false},
		{description: "Invalid Empty string", time: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ok := IsValidRFC3339(fmt.Sprintf(tt.time))
			if tt.expected && !ok {
				t.Fatalf("Expected %t, got %t", tt.expected, ok)
			} else if ok && !tt.expected {
				t.Fatalf("Expected %t, got %t", tt.expected, ok)
			}
		})
	}
}
