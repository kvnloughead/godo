package assert

import "testing"

func TestSliceEqual(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := []int{4, 3, 2, 1}
	c := []int{1, 2, 3, 5}

	if !SlicesAreEqual(t, a, b) {
		t.Errorf("expected SlicesAreEqual to pass, but it failed")
	}

	if SlicesAreEqual(t, a, c) {
		t.Errorf("expected SlicesAreEqual to fail, but it passed")
	}
}
