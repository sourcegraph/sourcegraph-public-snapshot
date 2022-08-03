package errors

import (
	"testing"
)

func TestWarningError(t *testing.T) {
	err := New("foo")
	var ref Warning

	// Ensure that all errors are not a warning type error.
	if As(err, &ref) {
		t.Error(`Expected error "err" to NOT be of type warning`)
	}

	// Ensure that all warning type errors are indeed a Warning type error.
	w := NewWarningError(err)
	if !As(w, &ref) {
		t.Error(`Expected error "w" to be of type warning`)
	}

	// Test the warning.As method.
	if !w.As(ref) {
		t.Error("Expected warning.As to return true but got false")
	}

	// Test that IsWarning always returns true.
	if !w.IsWarning() {
		t.Error("Expecting warning.IsWarning to return true but got false")
	}
}
