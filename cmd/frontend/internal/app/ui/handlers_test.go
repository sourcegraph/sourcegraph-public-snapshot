package ui

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestRepoShortName(t *testing.T) {
	tests := []struct {
		input api.RepoURI
		want  string
	}{
		{input: "repo", want: "repo"},
		{input: "github.com/foo/bar", want: "foo/bar"},
		{input: "mycompany.com/foo", want: "foo"},
	}
	for _, tst := range tests {
		t.Run(string(tst.input), func(t *testing.T) {
			got := repoShortName(tst.input)
			if got != tst.want {
				t.Fatalf("input %q got %q want %q", tst.input, got, tst.want)
			}
		})
	}
}
