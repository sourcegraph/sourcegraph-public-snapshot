package app

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGuessRepoNameFromRemoteURL(t *testing.T) {
	cases := []struct {
		url     string
		pattern string
		expName api.RepoName
	}{
		{"github.com:a/b", "", "github.com/a/b"},
		{"github.com:a/b.git", "", "github.com/a/b"},
		{"git@github.com:a/b", "", "github.com/a/b"},
		{"git@github.com:a/b.git", "", "github.com/a/b"},
		{"ssh://git@github.com/a/b.git", "", "github.com/a/b"},
		{"ssh://github.com/a/b.git", "", "github.com/a/b"},
		{"ssh://github.com:1234/a/b.git", "", "github.com/a/b"},
		{"https://github.com:1234/a/b.git", "", "github.com/a/b"},
		{"http://alice@foo.com:1234/a/b", "", "foo.com/a/b"},
		{"github.com:a/b", "{hostname}/{path}", "github.com/a/b"},
		{"github.com:a/b", "{hostname}-{path}", "github.com-a/b"},
		{"github.com:a/b", "{path}", "a/b"},
		{"github.com:a/b", "{hostname}", "github.com"},
	}
	for _, c := range cases {
		if got, want := guessRepoNameFromRemoteURL(c.url, c.pattern), c.expName; got != want {
			t.Errorf("%+v: got %q, want %q", c, got, want)
		}
	}
}
