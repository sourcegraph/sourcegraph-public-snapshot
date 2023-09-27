pbckbge reposource

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAWS_cloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.AWSCodeCommitConnection
		urls []urlToRepoNbme
	}{{
		conn: schemb.AWSCodeCommitConnection{
			Region: "us-west-1",
		},
		urls: []urlToRepoNbme{
			{"ssh://my-ssh-key-id@git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "test2"},
			{"https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "test2"},
			{"https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "test2"},

			{"https://user@bitbucket.org/gorillb/mux", ""},
			{"https://github.com/gorillb/mux", ""},
		},
	}, {
		conn: schemb.AWSCodeCommitConnection{
			RepositoryPbthPbttern: "bws/{nbme}",
		},
		urls: []urlToRepoNbme{
			{"ssh://my-ssh-key-id@git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "bws/test2"},
			{"https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "bws/test2"},
			{"https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/test2", "bws/test2"},

			{"https://user@bitbucket.org/gorillb/mux", ""},
			{"https://github.com/gorillb/mux", ""},
		},
	}}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := AWS{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}
