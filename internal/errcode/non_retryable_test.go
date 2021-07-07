package errcode_test

import (
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func TestMakeNonRetryable(t *testing.T) {
	err := errors.New("foo")
	if errcode.IsNonRetryable(err) {
		t.Errorf("unexpected non-retryable error: %+v", err)
	}

	if nrerr := errcode.MakeNonRetryable(err); !errcode.IsNonRetryable(nrerr) {
		t.Errorf("unexpected retryable error: %+v", nrerr)
	}
}
