package graphqlbackend

import (
	"testing"
)

func TestEscapePathForURL(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Example repo names
		{"sourcegraph/sourcegraph", "sourcegraph/sourcegraph"},
		{"sourcegraph.visualstudio.com/Test Repo With Spaces", "sourcegraph.visualstudio.com/Test%20Repo%20With%20Spaces"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got := escapePathForURL(test.path)
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
		})
	}
}
