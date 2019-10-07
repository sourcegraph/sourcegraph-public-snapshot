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
