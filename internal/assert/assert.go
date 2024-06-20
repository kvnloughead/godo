package assert

import (
	"regexp"
	"strings"
	"testing"
)

// A test helper that calls t.Errorf() if actual and expected are not equal.
func Equal[T comparable](t *testing.T, actual, expected T) {
	// Test helpers won't be cited in test output.
	t.Helper()

	if actual != expected {
		t.Errorf("got %v; want %v", actual, expected)
	}
}

// A test helper that calls t.Errorf() if actual doesn't contain expectedSubstr.
func StringContains(t *testing.T, actual, expectedSubstr string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstr) {
		t.Errorf("got %q; expected to contain %q", actual, expectedSubstr)
	}
}

// A test helper that calls t.Errorf() if there are no matches of expectedMatch found in actual.
func StringContainsMatch(t *testing.T, actual string, expectedMatch *regexp.Regexp) {
	t.Helper()

	matches := expectedMatch.FindStringSubmatch(actual)
	if len(matches) < 1 {
		t.Errorf("got %q; expected to match %q", actual, expectedMatch)
	}
}

// A test helper that calls t.Errorf() if actual is not nil.
func IsNil(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got %q; expected: nil", actual)
	}
}

// SliceEqual returns true if the actual and expected slice arguments contain
// the same elements.
func SlicesAreEqual[T comparable](t *testing.T, a []T, b []T) bool {
	t.Helper()

	if len(a) != len(b) {
		t.Errorf("slices have different lengths: %d and %d", len(a), len(b))
	}

	countA := make(map[T]int)
	countB := make(map[T]int)

	for _, v := range a {
		countA[v]++
	}

	for _, v := range b {
		countB[v]++
	}

	for key, count := range countA {
		if countB[key] != count {
			return false
		}
	}

	return true
}
