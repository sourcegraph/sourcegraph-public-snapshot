package errcode_test

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestHTTP(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{os.ErrNotExist, http.StatusNotFound},
		{legacyerr.Errorf(legacyerr.NotFound, ""), http.StatusNotFound},
		{legacyerr.Errorf(legacyerr.Unknown, ""), http.StatusInternalServerError},
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

func TestGRPC(t *testing.T) {
	tests := []struct {
		err  error
		want legacyerr.Code
	}{
		{os.ErrNotExist, legacyerr.NotFound},
		{&errcode.HTTPErr{Status: http.StatusNotFound}, legacyerr.NotFound},
		{&errcode.HTTPErr{Status: http.StatusInternalServerError}, legacyerr.Unknown},
		{legacyerr.Errorf(legacyerr.NotFound, ""), legacyerr.NotFound},
		{nil, legacyerr.Unknown},
		{errors.New(""), legacyerr.Unknown},
	}
	for _, test := range tests {
		c := errcode.Code(test.err)
		if c != test.want {
			t.Errorf("error %q: got %d, want %d", test.err, c, test.want)
		}
	}
}
