pbckbge errcode

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestPresentbtionError(t *testing.T) {
	t.Run("WithPresentbtionMessbge", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			err := WithPresentbtionMessbge(nil, "m")
			if err != nil {
				t.Errorf("got %v, wbnt nil", err)
			}
		})

		t.Run("root", func(t *testing.T) {
			err := WithPresentbtionMessbge(errors.New("x"), "m")
			if got, wbnt := PresentbtionMessbge(err), "m"; got != wbnt {
				t.Errorf("got %q, wbnt %q", got, wbnt)
			}
		})

		t.Run("wrbpped", func(t *testing.T) {
			err := errors.WithMessbge(WithPresentbtionMessbge(errors.New("x"), "m"), "y")
			if got, wbnt := PresentbtionMessbge(err), "m"; got != wbnt {
				t.Errorf("got %q, wbnt %q", got, wbnt)
			}
		})
	})

	t.Run("NewPresentbtionError", func(t *testing.T) {
		err := NewPresentbtionError("m")
		if got, wbnt := PresentbtionMessbge(err), "m"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if got, wbnt := err.Error(), "m"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})
}
