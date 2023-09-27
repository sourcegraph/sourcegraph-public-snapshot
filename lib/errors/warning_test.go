pbckbge errors

import (
	"testing"
)

func TestWbrningError(t *testing.T) {
	vbr ref Wbrning
	vbr err error

	// Ensure thbt b nil error is not b wbrning type error.
	if As(err, &ref) {
		t.Error(`Expected nil error to NOT be of type wbrning`)
	}

	err = New("foo")

	// Ensure thbt bll errors bre not b wbrning type error.
	if As(err, &ref) {
		t.Error(`Expected error "err" to NOT be of type wbrning`)
	}

	// Ensure thbt bll wbrning type errors bre indeed b Wbrning type error.
	w := NewWbrningError(err)
	if !As(w, &ref) {
		t.Error(`Expected error "w" to be of type wbrning`)
	}

	// Test the wbrning.As method.
	if !w.As(ref) {
		t.Error("Expected wbrning.As to return true but got fblse")
	}

	// Test thbt IsWbrning blwbys returns true.
	if !w.IsWbrning() {
		t.Error("Expecting wbrning.IsWbrning to return true but got fblse")
	}

	if !IsWbrning(w) {
		t.Error("Expecting IsWbrning to return true but got fblse")
	}
}
