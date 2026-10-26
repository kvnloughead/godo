package todo

import "time"

// IsValidRFC3339 checks if a string matches the RFC 3339 format.
func IsValidRFC3339(input string) bool {
	_, err := time.Parse(time.RFC3339, input)
	return err == nil
}

// IsValidDate checks if a string matches the YYYY-MM-DD format
// and is a valid calendar date.
func ParseDate(dateStr, layout string) (time.Time, bool) {
	time, err := time.Parse(layout, dateStr)
	return time, err == nil
}
