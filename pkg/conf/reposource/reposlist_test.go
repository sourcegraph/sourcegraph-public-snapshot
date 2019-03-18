package reposource

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestReposList_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		repos []*schema.Repository
		urls  []urlURI
	}{{
		repos: []*schema.Repository{
			{Path: "github.com/gorilla/mux", Url: "https://github.com/gorilla/mux"},
		},
		urls: []urlURI{
			{"https://github.com/gorilla/mux", "github.com/gorilla/mux"},
			{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"git@github.com:gorilla/mux", "github.com/gorilla/mux"},
			{"git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
			{"ssh://git@github.com:1234/gorilla/mux.git", "github.com/gorilla/mux"},

			{"https://asdf.com/gorilla/mux", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"git@asdf.com:gorilla/mux", ""},
			{"git@asdf.com:gorilla/mux.git", ""},
			{"ssh://git@asdf.com:1234/gorilla/mux.git", ""},

			{"https://github.com/gorilla/pat", ""},
			{"https://github.com/gorilla/pat.git", ""},
			{"git@github.com:gorilla/pat", ""},
			{"git@github.com:gorilla/pat.git", ""},
			{"ssh://git@github.com:1234/gorilla/pat.git", ""},

			{"https://github.com/asdf/mux", ""},
			{"https://github.com/asdf/mux.git", ""},
			{"git@github.com:asdf/mux", ""},
			{"git@github.com:asdf/mux.git", ""},
			{"ssh://git@github.com:1234/asdf/mux.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := newReposList(test.repos).cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}

			if u.repoURI != string(repoURI) {
				b, _ := json.Marshal(test.repos)
				t.Errorf("expected %q but got %q for clone URL %q (repos: %s)", u.repoURI, repoURI, u.cloneURL, string(b))
			}
		}
	}
}
