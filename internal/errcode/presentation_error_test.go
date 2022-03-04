package errcode

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestPresentationError(t *testing.T) {
	t.Run("WithPresentationMessage", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			err := WithPresentationMessage(nil, "m")
			if err != nil {
				t.Errorf("got %v, want nil", err)
			}
		})

		t.Run("root", func(t *testing.T) {
			err := WithPresentationMessage(errors.New("x"), "m")
			if got, want := PresentationMessage(err), "m"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})

		t.Run("wrapped", func(t *testing.T) {
			err := errors.WithMessage(WithPresentationMessage(errors.New("x"), "m"), "y")
			if got, want := PresentationMessage(err), "m"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	})

	t.Run("NewPresentationError", func(t *testing.T) {
		err := NewPresentationError("m")
		if got, want := PresentationMessage(err), "m"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if got, want := err.Error(), "m"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
