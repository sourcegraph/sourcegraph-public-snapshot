package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitHub_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		conn schema.GitHubConnection
		urls []urlURI
	}{{
		conn: schema.GitHubConnection{
			Url: "https://github.com",
		},
		urls: []urlURI{
			{"git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
			{"git@github.com:/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://oauth2:ACCESS_TOKEN@github.com/gorilla/mux.git", "github.com/gorilla/mux"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}, {
		conn: schema.GitHubConnection{
			Url:                   "https://github.mycompany.com",
			RepositoryPathPattern: "{nameWithOwner}",
		},
		urls: []urlURI{
			{"git@github.mycompany.com:foo/bar/baz.git", "foo/bar/baz"},
			{"https://github.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},
			{"https://oauth2:ACCESS_TOKEN@github.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := GitHub{&test.conn}.cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoURI != string(repoURI) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoURI, repoURI, u.cloneURL, test.conn)
			}
		}
	}
}
