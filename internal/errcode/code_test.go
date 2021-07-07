package errcode_test

import (
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func TestIsNotFound(t *testing.T) {
	if !errcode.IsNotFound(&notFoundErr{}) {
		t.Errorf("unexpectedly found")
	}
	if errcode.IsNotFound(&errcode.Mock{IsNotFound: false}) {
		t.Errorf("unexpectedly not found")
	}
	if errcode.IsNotFound(errors.New("test")) {
		t.Errorf("unexpectedly not found")
	}
}

type notFoundErr struct{}

func (e *notFoundErr) Error() string {
	return "not found"
}

func (e *notFoundErr) NotFound() bool {
	return true
}
