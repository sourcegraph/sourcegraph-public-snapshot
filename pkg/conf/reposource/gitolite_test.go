package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitolite_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		conn schema.GitoliteConnection
		urls []urlURI
	}{{
		conn: schema.GitoliteConnection{
			Host:   "git@gitolite.sgdev.org",
			Prefix: "gitolite.sgdev.org/",
		},
		urls: []urlURI{
			{"git@gitolite.sgdev.org:bl/go/app.git", "gitolite.sgdev.org/bl/go/app"},
			{"git@gitolite.sgdev.org:bl/go/app", "gitolite.sgdev.org/bl/go/app"},

			{"git@asdf.org:bl/go/app.git", ""},
			{"git@asdf.org:bl/go/app", ""},
		},
	}, {
		conn: schema.GitoliteConnection{
			Host: "git@gitolite.sgdev.org",
		},
		urls: []urlURI{
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
		urls: []urlURI{
			{"git@gitolite.sgdev.org:bl/go/app.git", "git/bl/go/app"},
			{"git@gitolite.sgdev.org:bl/go/app", "git/bl/go/app"},

			{"git@asdf.org:bl/go/app.git", ""},
			{"git@asdf.org:bl/go/app", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := Gitolite{&test.conn}.cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoURI != string(repoURI) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoURI, repoURI, u.cloneURL, test.conn)
			}
		}
	}
}
