package todo

import "time"

// IsValidRFC3339 checks if a string matches the RFC 3339 format.
func IsValidRFC3339(input string) bool {
	_, err := time.Parse(time.RFC3339, input)
	return err == nil
}
