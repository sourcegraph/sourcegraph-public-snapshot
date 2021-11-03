package graphqlbackend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var (
	rawCursor    = database.Cursor{Column: "foo", Value: "bar", Direction: "next"}
	opaqueCursor = "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6ImZvbyIsIlZhbHVlIjoiYmFyIiwiRGlyZWN0aW9uIjoibmV4dCJ9"
)

func TestMarshalRepositoryCursor(t *testing.T) {
	if got, want := marshalRepositoryCursor(&rawCursor), opaqueCursor; got != want {
		t.Errorf("got opaque cursor %q, want %q", got, want)
	}
}

func TestUnmarshalRepositoryCursor(t *testing.T) {
	cursor, err := unmarshalRepositoryCursor(&opaqueCursor)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(cursor, &rawCursor); diff != "" {
		t.Fatal(diff)
	}
}
