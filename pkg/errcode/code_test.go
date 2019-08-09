package errcode_test

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestHTTP(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{os.ErrNotExist, http.StatusNotFound},
		{&notFoundErr{}, http.StatusNotFound},
		{nil, http.StatusOK},
		{errors.New(""), http.StatusInternalServerError},
	}
	for _, test := range tests {
		c := errcode.HTTP(test.err)
		if c != test.want {
			t.Errorf("error %q: got %d, want %d", test.err, c, test.want)
		}
	}
}

type notFoundErr struct{}

func (e *notFoundErr) Error() string {
	return "not found"
}

func (e *notFoundErr) NotFound() bool {
	return true
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_774(size int) error {
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
