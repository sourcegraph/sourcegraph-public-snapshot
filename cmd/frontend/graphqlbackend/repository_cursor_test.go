package graphqlbackend

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	rawCursor    = types.Cursor{Column: "foo", Value: "bar", Direction: "next"}
	opaqueCursor = "UmVwb3NpdG9yeUN1cnNvcjp7ImMiOiJmb28iLCJ2IjoiYmFyIiwiZCI6Im5leHQifQ=="
)

func TestMarshalRepositoryCursor(t *testing.T) {
	if got, want := MarshalRepositoryCursor(&rawCursor), opaqueCursor; got != want {
		t.Errorf("got opaque cursor %q, want %q", got, want)
	}
}

func TestUnmarshalRepositoryCursor(t *testing.T) {
	cursor, err := UnmarshalRepositoryCursor(&opaqueCursor)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(cursor, &rawCursor); diff != "" {
		t.Fatal(diff)
	}
}
