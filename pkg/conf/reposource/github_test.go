package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitHub_cloneURLToRepoName(t *testing.T) {
	var tests = []struct {
		conn schema.GitHubConnection
		urls []urlToRepoName
	}{{
		conn: schema.GitHubConnection{
			Url: "https://github.com",
		},
		urls: []urlToRepoName{
			{"git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
			{"git@github.com:/gorilla/mux.git", "github.com/gorilla/mux"},
			{"git+https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://www.github.com/gorilla/mux.git", "github.com/gorilla/mux"},
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
		urls: []urlToRepoName{
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
			repoName, err := GitHub{&test.conn}.CloneURLToRepoName(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoName != string(repoName) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoName, repoName, u.cloneURL, test.conn)
			}
		}
	}
}
