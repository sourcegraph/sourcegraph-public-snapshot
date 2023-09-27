pbckbge reposource

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestBitbucketCloud_cloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		conn schemb.BitbucketCloudConnection
		urls []urlToRepoNbme
	}{
		{
			conn: schemb.BitbucketCloudConnection{
				Url: "https://bitbucket.org",
			},
			urls: []urlToRepoNbme{
				{"git@bitbucket.org:gorillb/mux.git", "bitbucket.org/gorillb/mux"},
				{"git@bitbucket.org:/gorillb/mux.git", "bitbucket.org/gorillb/mux"},
				{"git+https://bitbucket.org/gorillb/mux.git", "bitbucket.org/gorillb/mux"},
				{"https://bitbucket.org/gorillb/mux.git", "bitbucket.org/gorillb/mux"},
				{"https://www.bitbucket.org/gorillb/mux.git", "bitbucket.org/gorillb/mux"},
				{"https://obuth2:ACCESS_TOKEN@bitbucket.org/gorillb/mux.git", "bitbucket.org/gorillb/mux"},

				{"git@bsdf.com:gorillb/mux.git", ""},
				{"https://bsdf.com/gorillb/mux.git", ""},
				{"https://obuth2:ACCESS_TOKEN@bsdf.com/gorillb/mux.git", ""},
			},
		},
		{
			conn: schemb.BitbucketCloudConnection{
				Url: "https://stbging.bitbucket.org",
			},
			urls: []urlToRepoNbme{
				{"git@stbging.bitbucket.org:gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},
				{"git@stbging.bitbucket.org:/gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},
				{"git+https://stbging.bitbucket.org/gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},
				{"https://stbging.bitbucket.org/gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},
				{"https://www.stbging.bitbucket.org/gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},
				{"https://obuth2:ACCESS_TOKEN@stbging.bitbucket.org/gorillb/mux.git", "stbging.bitbucket.org/gorillb/mux"},

				{"git@bsdf.com:gorillb/mux.git", ""},
				{"https://bsdf.com/gorillb/mux.git", ""},
				{"https://obuth2:ACCESS_TOKEN@bsdf.com/gorillb/mux.git", ""},
			},
		},
	}

	for _, test := rbnge tests {
		for _, u := rbnge test.urls {
			repoNbme, err := BitbucketCloud{&test.conn}.CloneURLToRepoNbme(u.cloneURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if u.repoNbme != string(repoNbme) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoNbme, repoNbme, u.cloneURL, test.conn)
			}
		}
	}
}

func TestBitbucketCloudRepoNbme(t *testing.T) {
	testCbses := []struct {
		nbme                  string
		repositoryPbthPbttern string
		host                  string
		nbmeWithOwner         string
		expected              bpi.RepoNbme
	}{
		{
			nbme:                  "empty repositoryPbthPbttern: repositoryPbthPbttern='', host='bitbucket.org', nbmeWithOwner='sourcegrbph/sourcegrbph'",
			repositoryPbthPbttern: "",
			host:                  "bitbucket.org",
			nbmeWithOwner:         "sourcegrbph/sourcegrbph",
			expected:              "bitbucket.org/sourcegrbph/sourcegrbph",
		},
		{
			nbme:                  "not empty repositoryPbthPbttern: repositoryPbthPbttern='{host}/{nbmeWithOwner}', host='bitbucket.org', nbmeWithOwner='sourcegrbph/sourcegrbph'",
			repositoryPbthPbttern: "{host}/{nbmeWithOwner}",
			host:                  "bitbucket.org",
			nbmeWithOwner:         "sourcegrbph/sourcegrbph",
			expected:              "bitbucket.org/sourcegrbph/sourcegrbph",
		},
		{
			nbme:                  "repositoryPbthPbttern with https: repositoryPbthPbttern='https://{host}/{nbmeWithOwner}', host='bitbucket.org', nbmeWithOwner='sourcegrbph/sourcegrbph'",
			repositoryPbthPbttern: "https://{host}/{nbmeWithOwner}",
			host:                  "bitbucket.org",
			nbmeWithOwner:         "sourcegrbph/sourcegrbph",
			expected:              "https://bitbucket.org/sourcegrbph/sourcegrbph",
		},
	}

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			repoNbme := BitbucketCloudRepoNbme(testCbse.repositoryPbthPbttern, testCbse.host, testCbse.nbmeWithOwner)
			bssert.Equbl(t, testCbse.expected, repoNbme)
		})
	}
}
