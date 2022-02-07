package errcode_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func TestMakeNonRetryable(t *testing.T) {
	err := errors.New("foo")
	if errcode.IsNonRetryable(err) {
		t.Errorf("unexpected non-retryable error: %+v", err)
	}

	if nrerr := errcode.MakeNonRetryable(err); !errcode.IsNonRetryable(nrerr) {
		t.Errorf("unexpected retryable error: %+v", nrerr)
	}
}

type notFoundErr struct{}

func (e *notFoundErr) Error() string {
	return "not found"
}

func (e *notFoundErr) NotFound() bool {
	return true
}
