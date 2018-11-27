package app

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGuessRepoURIFromRemoteURL(t *testing.T) {
	cases := []struct {
		url               string
		hostnameToPattern map[string]string
		expName           api.RepoURI
	}{
		{"github.com:a/b", nil, "github.com/a/b"},
		{"github.com:a/b.git", nil, "github.com/a/b"},
		{"git@github.com:a/b", nil, "github.com/a/b"},
		{"git@github.com:a/b.git", nil, "github.com/a/b"},
		{"ssh://git@github.com/a/b.git", nil, "github.com/a/b"},
		{"ssh://github.com/a/b.git", nil, "github.com/a/b"},
		{"ssh://github.com:1234/a/b.git", nil, "github.com/a/b"},
		{"https://github.com:1234/a/b.git", nil, "github.com/a/b"},
		{"http://alice@foo.com:1234/a/b", nil, "foo.com/a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}/{path}"}, "github.com/a/b"},
		{"github.com:a/b", map[string]string{"asdf.com": "{hostname}-----{path}"}, "github.com/a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}-{path}"}, "github.com-a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{path}"}, "a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}"}, "github.com"},
		{"github.com:a/b", map[string]string{"github.com": "github/{path}", "asdf.com": "asdf/{path}"}, "github/a/b"},
		{"asdf.com:a/b", map[string]string{"github.com": "github/{path}", "asdf.com": "asdf/{path}"}, "asdf/a/b"},
	}
	for _, c := range cases {
		if got, want := guessRepoURIFromRemoteURL(c.url, c.hostnameToPattern), c.expName; got != want {
			t.Errorf("%+v: got %q, want %q", c, got, want)
		}
	}
}
