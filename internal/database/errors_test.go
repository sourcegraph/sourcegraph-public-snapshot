package database

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func TestResourceNotFoundError(t *testing.T) {
	err := resourceNotFoundError{"foo"}
	if want := "foo not found"; err.Error() != want {
		t.Errorf("got %q, want %q", err, want)
	}

	if !errcode.IsNotFound(resourceNotFoundError{"foo"}) {
		t.Fatal()
	}
	if !errcode.IsNotFound(&resourceNotFoundError{"foo"}) {
		t.Fatal()
	}

}
