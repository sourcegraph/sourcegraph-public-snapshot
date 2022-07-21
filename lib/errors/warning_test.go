package errors

import (
	"testing"

	"github.com/cockroachdb/errors"
)

func TestWarningError(t *testing.T) {
	err := errors.New("foo")
	w := NewWarningError(err)
	if _, ok := w.(Warning); !ok {
		t.Error(`Expected variable "w" to be of type Warning`)
	}

	if errors.Is(err, &warning{}) {
		t.Error(`Expected variable "err" to not be of type Warning`)
	}
}
