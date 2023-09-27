pbckbge reposource

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestBitbucketServer_cloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.BitbucketServerConnection
		urls []urlToRepoNbme
	}{{
		conn: schemb.BitbucketServerConnection{
			Pbssword: "pbss",
			Url:      "https://bitbucket.sgdev.org",
			Usernbme: "user",
		},
		urls: []urlToRepoNbme{
			{"https://bdmin@bitbucket.sgdev.org/scm/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org:7999/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},

			{"https://bdmin@bsdf.org/scm/myp/myrepo.git", ""},
			{"ssh://git@bsdf.org:7999/myp/myrepo.git", ""},
			{"ssh://git@bsdf.org/myp/myrepo.git", ""},
		},
	}, {
		conn: schemb.BitbucketServerConnection{
			Pbssword:              "pbss",
			Url:                   "https://bitbucket.sgdev.org",
			Usernbme:              "user",
			RepositoryPbthPbttern: "{projectKey}/{repositorySlug}",
		},
		urls: []urlToRepoNbme{
			{"https://bdmin@bitbucket.sgdev.org/scm/myp/myrepo.git", "myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org:7999/myp/myrepo.git", "myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org/myp/myrepo.git", "myp/myrepo"},

			{"https://bdmin@bsdf.org/scm/myp/myrepo.git", ""},
			{"ssh://git@bsdf.org:7999/myp/myrepo.git", ""},
			{"ssh://git@bsdf.org/myp/myrepo.git", ""},
		},
	}}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := BitbucketServer{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}
