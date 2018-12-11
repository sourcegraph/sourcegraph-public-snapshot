package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitolite_cloneURLToRepoName(t *testing.T) {
	var tests = []struct {
		conn schema.GitoliteConnection
		urls []urlToRepoName
	}{{
		conn: schema.GitoliteConnection{
			Host:   "git@gitolite.sgdev.org",
			Prefix: "gitolite.sgdev.org/",
		},
		urls: []urlToRepoName{
			{"git@gitolite.sgdev.org:bl/go/app.git", "gitolite.sgdev.org/bl/go/app"},
			{"git@gitolite.sgdev.org:bl/go/app", "gitolite.sgdev.org/bl/go/app"},

			{"git@asdf.org:bl/go/app.git", ""},
			{"git@asdf.org:bl/go/app", ""},
		},
	}, {
		conn: schema.GitoliteConnection{
			Host: "git@gitolite.sgdev.org",
		},
		urls: []urlToRepoName{
			{"git@gitolite.sgdev.org:bl/go/app.git", "bl/go/app"},
			{"git@gitolite.sgdev.org:bl/go/app", "bl/go/app"},

			{"git@asdf.org:bl/go/app.git", ""},
			{"git@asdf.org:bl/go/app", ""},
		},
	}, {
		conn: schema.GitoliteConnection{
			Host:   "git@gitolite.sgdev.org",
			Prefix: "git/",
		},
		urls: []urlToRepoName{
			{"git@gitolite.sgdev.org:bl/go/app.git", "git/bl/go/app"},
			{"git@gitolite.sgdev.org:bl/go/app", "git/bl/go/app"},

			{"git@asdf.org:bl/go/app.git", ""},
			{"git@asdf.org:bl/go/app", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoName, err := Gitolite{&test.conn}.CloneURLToRepoName(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoName != string(repoName) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoName, repoName, u.cloneURL, test.conn)
			}
		}
	}
}
