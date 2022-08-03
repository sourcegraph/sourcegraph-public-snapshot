package errors

import (
	"testing"
)

func TestWarningError(t *testing.T) {
	var ref Warning
	var err error

	// Ensure that a nil error is not a warning type error.
	if As(err, &ref) {
		t.Error(`Expected nil error to NOT be of type warning`)
	}

	err = New("foo")

	t.Run("all errors are not a warning type error", func(t *testing.T) {
		if As(err, &ref) {
			t.Error(`Expected error "err" to NOT be of type warning`)
		}
	})

	w := NewWarningError(err)

	t.Run("all warning type errors are indeed a Warning type error", func(t *testing.T) {
		if !As(w, &ref) {
			t.Error(`Expected error "w" to be of type warning`)
		}
	})

	// Test that IsWarning always returns true.
	if !w.IsWarning() {
		t.Error("Expecting warning.IsWarning to return true but got false")
	}

	if !IsWarning(w) {
		t.Error("Expecting IsWarning to return true but got false")
	}

	t.Run("test the warning.As method against the Warning interface", func(t *testing.T) {
		if !w.As(ref) {
			t.Error("Expected warning.As to return true but got false")
		}
	})

	t.Run("test the warning.As method against the error interface.", func(t *testing.T) {
		var e error
		if w.As(e) {
			t.Error("Expected warning.As to return false but got true")
		}
	})

	t.Run("test that IsWarning always returns true.", func(t *testing.T) {
		if !w.IsWarning() {
			t.Error("Expecting warning.IsWarning to return true but got false")
		}
	})
}
