package errcode

import (
	"testing"

	"github.com/pkg/errors"
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_776(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
