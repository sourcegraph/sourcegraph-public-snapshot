package errors

import (
	"testing"

	"github.com/cockroachdb/errors"
)

func TestWarningError(t *testing.T) {
	err := New("foo")
	var ref Warning

	// Ensure that all errors are not a warning type error.
	if errors.As(err, &ref) {
		t.Error(`Expected error "err" to NOT be of type warning`)
	}

	// Ensure that all warning type errors are indeed a Warning type error.
	w := NewWarningError(err)
	if !errors.As(w, &ref) {
		t.Error(`Expected error "w" to be of type warning`)
	}
}
