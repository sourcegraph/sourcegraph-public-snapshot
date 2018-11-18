package app

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGuessRepoNameFromRemoteURL(t *testing.T) {
	type URLAndMapping struct {
		url     string
		mapping string
	}
	tests := map[URLAndMapping]api.RepoName{
		URLAndMapping{"github.com:a/b", "broken_json}"}:                       "github.com/a/b",
		URLAndMapping{"github.com:a/b.git", "{}"}:                             "github.com/a/b",
		URLAndMapping{"git@github.com:a/b", ""}:                               "github.com/a/b",
		URLAndMapping{"git@github.com:a/b.git", ""}:                           "github.com/a/b",
		URLAndMapping{"ssh://git@github.com/a/b.git", ""}:                     "github.com/a/b",
		URLAndMapping{"ssh://github.com/a/b.git", ""}:                         "github.com/a/b",
		URLAndMapping{"ssh://github.com:1234/a/b.git", ""}:                    "github.com/a/b",
		URLAndMapping{"https://github.com:1234/a/b.git", ""}:                  "github.com/a/b",
		URLAndMapping{"https://github.com:1234/a/b.git", "{\"x\": \"y\"}"}:    "github.com/a/b",
		URLAndMapping{"http://alice@foo.com:1234/a/b", ""}:                    "foo.com/a/b",
		URLAndMapping{"http://alice@foo.com:1234/a/b", "{\"foo.com\": \"\"}"}: "/a/b",
	}
	for input, want := range tests {
		got := guessRepoNameFromRemoteURL(input.url, input.mapping)
		if got != want {
			t.Errorf("%s: got %q, want %q", input, got, want)
		}
	}
}
