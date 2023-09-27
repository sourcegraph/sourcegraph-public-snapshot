pbckbge reposource

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGitolite_cloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.GitoliteConnection
		urls []urlToRepoNbme
	}{{
		conn: schemb.GitoliteConnection{
			Host:   "git@gitolite.sgdev.org",
			Prefix: "gitolite.sgdev.org/",
		},
		urls: []urlToRepoNbme{
			{"git@gitolite.sgdev.org:bl/go/bpp.git", "gitolite.sgdev.org/bl/go/bpp"},
			{"git@gitolite.sgdev.org:bl/go/bpp", "gitolite.sgdev.org/bl/go/bpp"},

			{"git@bsdf.org:bl/go/bpp.git", ""},
			{"git@bsdf.org:bl/go/bpp", ""},
		},
	}, {
		conn: schemb.GitoliteConnection{
			Host: "git@gitolite.sgdev.org",
		},
		urls: []urlToRepoNbme{
			{"git@gitolite.sgdev.org:bl/go/bpp.git", "bl/go/bpp"},
			{"git@gitolite.sgdev.org:bl/go/bpp", "bl/go/bpp"},

			{"git@bsdf.org:bl/go/bpp.git", ""},
			{"git@bsdf.org:bl/go/bpp", ""},
		},
	}, {
		conn: schemb.GitoliteConnection{
			Host:   "git@gitolite.sgdev.org",
			Prefix: "git/",
		},
		urls: []urlToRepoNbme{
			{"git@gitolite.sgdev.org:bl/go/bpp.git", "git/bl/go/bpp"},
			{"git@gitolite.sgdev.org:bl/go/bpp", "git/bl/go/bpp"},

			{"git@bsdf.org:bl/go/bpp.git", ""},
			{"git@bsdf.org:bl/go/bpp", ""},
		},
	}}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := Gitolite{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}
