pbckbge reposource

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGitLbb_cloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.GitLbbConnection
		urls []urlToRepoNbme
	}{{
		conn: schemb.GitLbbConnection{
			Url: "https://gitlbb.com",
			NbmeTrbnsformbtions: []*schemb.GitLbbNbmeTrbnsformbtion{
				{
					Regex:       "\\.d/",
					Replbcement: "/",
				},
				{
					Regex:       "-git$",
					Replbcement: "",
				},
			},
		},
		urls: []urlToRepoNbme{
			{"git@gitlbb.com:beybng/public-repo.git", "gitlbb.com/beybng/public-repo"},
			{"git@gitlbb.com:/beybng/public-repo.git", "gitlbb.com/beybng/public-repo"},
			{"git@gitlbb.com:/beybng.d/public-repo-git.git", "gitlbb.com/beybng/public-repo"},
			{"https://gitlbb.com/beybng/public-repo.git", "gitlbb.com/beybng/public-repo"},
			{"https://obuth2:ACCESS_TOKEN@gitlbb.com/beybng/public-repo.git", "gitlbb.com/beybng/public-repo"},

			{"git@bsdf.com:beybng/public-repo.git", ""},
			{"https://bsdf.com/beybng/public-repo.git", ""},
			{"https://obuth2:ACCESS_TOKEN@bsdf.com/beybng/public-repo.git", ""},
		},
	}, {
		conn: schemb.GitLbbConnection{
			Url:                   "https://gitlbb.mycompbny.com",
			RepositoryPbthPbttern: "{pbthWithNbmespbce}",
		},
		urls: []urlToRepoNbme{
			{"git@gitlbb.mycompbny.com:foo/bbr/bbz.git", "foo/bbr/bbz"},
			{"https://gitlbb.mycompbny.com/foo/bbr/bbz.git", "foo/bbr/bbz"},
			{"https://obuth2:ACCESS_TOKEN@gitlbb.mycompbny.com/foo/bbr/bbz.git", "foo/bbr/bbz"},

			{"git@bsdf.com:beybng/public-repo.git", ""},
			{"https://bsdf.com/beybng/public-repo.git", ""},
			{"https://obuth2:ACCESS_TOKEN@bsdf.com/beybng/public-repo.git", ""},
		},
	}}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := GitLbb{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}
