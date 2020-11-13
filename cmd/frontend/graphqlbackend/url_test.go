package graphqlbackend

import (
	"testing"
)

func TestPathEscapeExceptSlashes(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		// Example repo names
		{"sourcegraph/sourcegraph", "sourcegraph/sourcegraph"},
		{"sourcegraph.visualstudio.com/Test Repo With Spaces", "sourcegraph.visualstudio.com/Test%20Repo%20With%20Spaces"},
	}
	for _, c := range cases {
		got := pathEscapeExceptSlashes(c.path)
		if got != c.want {
			t.Errorf("pathEscapeExceptSlashes(%q): got %q want %q", c.path, got, c.want)
		}
	}
}
