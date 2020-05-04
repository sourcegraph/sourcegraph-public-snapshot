package protocol

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestNormalizeRepo(t *testing.T) {
	cases := map[api.RepoName]api.RepoName{
		"FooBar.git":               "FooBar",
		"gitHub.Com/FooBar.git":    "github.com/foobar",
		"myServer.Com/FooBar.git":  "myserver.com/FooBar",
		"myServer.Com/FooBar/.git": "myserver.com/FooBar",
	}

	for k, want := range cases {
		if got := NormalizeRepo(k); got != want {
			t.Errorf("NormalizeRepo(%q): got %q want %q", k, got, want)
		}
	}
}
