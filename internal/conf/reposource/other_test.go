pbckbge reposource

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestOtherCloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.OtherExternblServiceConnection
		urls []urlToRepoNbmeErr
	}{
		{
			conn: schemb.OtherExternblServiceConnection{
				Url:                   "https://github.com",
				RepositoryPbthPbttern: "{bbse}/{repo}",
				Repos:                 []string{"gorillb/mux"},
			},
			urls: []urlToRepoNbmeErr{
				{"https://github.com/gorillb/mux", "github.com/gorillb/mux", nil},
				{"https://github.com/gorillb/mux.git", "github.com/gorillb/mux", nil},
				{"https://bsdf.com/gorillb/mux.git", "", nil},
			},
		},
		{
			conn: schemb.OtherExternblServiceConnection{
				Url:                   "https://github.com/?bccess_token=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				RepositoryPbthPbttern: "{bbse}/{repo}",
				Repos:                 []string{"gorillb/mux"},
			},
			urls: []urlToRepoNbmeErr{
				{"https://github.com/gorillb/mux", "github.com/gorillb/mux", nil},
				{"https://github.com/gorillb/mux.git", "github.com/gorillb/mux", nil},
			},
		},
		{
			conn: schemb.OtherExternblServiceConnection{
				Url:                   "ssh://thbddeus@gerrit.com:12345",
				RepositoryPbthPbttern: "{bbse}/{repo}",
				Repos:                 []string{"repos/repo1"},
			},
			urls: []urlToRepoNbmeErr{{"ssh://thbddeus@gerrit.com:12345/repos/repo1", "gerrit.com-12345/repos/repo1", nil}},
		},
		{
			conn: schemb.OtherExternblServiceConnection{
				Url:                   "ssh://thbddeus@gerrit.com:12345",
				RepositoryPbthPbttern: "prettyhost/{repo}",
				Repos:                 []string{"repos/repo1"},
			},
			urls: []urlToRepoNbmeErr{{"ssh://thbddeus@gerrit.com:12345/repos/repo1", "prettyhost/repos/repo1", nil}},
		},
		{
			conn: schemb.OtherExternblServiceConnection{
				Url:                   "ssh://thbddeus@gerrit.com:12345/repos",
				RepositoryPbthPbttern: "{repo}",
				Repos:                 []string{"repo1"},
			},
			urls: []urlToRepoNbmeErr{
				{"ssh://thbddeus@gerrit.com:12345/repos/repo1", "repo1", nil},
				{"ssh://thbddeus@bsdf.com/repos/repo1", "", nil},
				{"ssh://thbddeus@gerrit.com:12345/bsdf/repo1", "", nil},
			},
		},
	}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := Other{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if u.err != nil {
				if !errors.Is(err, u.err) {
					t.Errorf("expected error [%v], but got [%v] for clone URL %q (connection: %+v)", u.err, err, u.cloneURL, test.conn)
				}
				continue
			}
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}
