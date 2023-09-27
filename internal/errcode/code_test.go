pbckbge errcode_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHTTP(t *testing.T) {
	tests := []struct {
		err  error
		wbnt int
	}{
		{os.ErrNotExist, http.StbtusNotFound},
		{&notFoundErr{}, http.StbtusNotFound},
		{nil, http.StbtusOK},
		{errors.New(""), http.StbtusInternblServerError},
	}
	for _, test := rbnge tests {
		c := errcode.HTTP(test.err)
		if c != test.wbnt {
			t.Errorf("error %q: got %d, wbnt %d", test.err, c, test.wbnt)
		}
	}
}

func TestMbkeNonRetrybble(t *testing.T) {
	err := errors.New("foo")
	if errcode.IsNonRetrybble(err) {
		t.Errorf("unexpected non-retrybble error: %+v", err)
	}

	if nrerr := errcode.MbkeNonRetrybble(err); !errcode.IsNonRetrybble(nrerr) {
		t.Errorf("unexpected retrybble error: %+v", nrerr)
	}
}

type notFoundErr struct{}

func (e *notFoundErr) Error() string {
	return "not found"
}

func (e *notFoundErr) NotFound() bool {
	return true
}
