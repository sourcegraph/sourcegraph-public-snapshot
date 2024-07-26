package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGerrit_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.GerritConnection
		urls []urlToRepoName
	}{{
		conn: schema.GerritConnection{
			Url: "https://gerrit.example.com",
		},
		urls: []urlToRepoName{
			{"git@gerrit.example.com:gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			{"git@gerrit.example.com:/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			{"git+https://gerrit.example.com/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			{"https://gerrit.example.com/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			{"https://www.gerrit.example.com/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			// Authenticated clone URL.
			{"https://www.gerrit.example.com/a/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},
			{"https://oauth2:ACCESS_TOKEN@gerrit.example.com/gorilla/mux.git", "gerrit.example.com/gorilla/mux"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}, {
		conn: schema.GerritConnection{
			Url:                   "https://gerrit.example.com",
			RepositoryPathPattern: "prefix/{name}",
		},
		urls: []urlToRepoName{
			{"git@gerrit.example.com:gorilla/mux.git", "prefix/gorilla/mux"},
			{"git@gerrit.example.com:/gorilla/mux.git", "prefix/gorilla/mux"},
			{"git+https://gerrit.example.com/gorilla/mux.git", "prefix/gorilla/mux"},
			{"https://gerrit.example.com/gorilla/mux.git", "prefix/gorilla/mux"},
			{"https://www.gerrit.example.com/gorilla/mux.git", "prefix/gorilla/mux"},
			// Authenticated clone URL.
			{"https://www.gerrit.example.com/a/gorilla/mux.git", "prefix/gorilla/mux"},
			{"https://oauth2:ACCESS_TOKEN@gerrit.example.com/gorilla/mux.git", "prefix/gorilla/mux"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoName, err := Gerrit{&test.conn}.CloneURLToRepoName(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoName != string(repoName) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoName, repoName, u.cloneURL, test.conn)
			}
		}
	}
}
