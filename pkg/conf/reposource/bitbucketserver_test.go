package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketServer_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		conn schema.BitbucketServerConnection
		urls []urlURI
	}{{
		conn: schema.BitbucketServerConnection{
			Password: "pass",
			Url:      "https://bitbucket.sgdev.org",
			Username: "user",
		},
		urls: []urlURI{
			{"https://admin@bitbucket.sgdev.org/scm/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org:7999/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org/myp/myrepo.git", "bitbucket.sgdev.org/myp/myrepo"},

			{"https://admin@asdf.org/scm/myp/myrepo.git", ""},
			{"ssh://git@asdf.org:7999/myp/myrepo.git", ""},
			{"ssh://git@asdf.org/myp/myrepo.git", ""},
		},
	}, {
		conn: schema.BitbucketServerConnection{
			Password:              "pass",
			Url:                   "https://bitbucket.sgdev.org",
			Username:              "user",
			RepositoryPathPattern: "{projectKey}/{repositorySlug}",
		},
		urls: []urlURI{
			{"https://admin@bitbucket.sgdev.org/scm/myp/myrepo.git", "myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org:7999/myp/myrepo.git", "myp/myrepo"},
			{"ssh://git@bitbucket.sgdev.org/myp/myrepo.git", "myp/myrepo"},

			{"https://admin@asdf.org/scm/myp/myrepo.git", ""},
			{"ssh://git@asdf.org:7999/myp/myrepo.git", ""},
			{"ssh://git@asdf.org/myp/myrepo.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := BitbucketServer{&test.conn}.cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoURI != string(repoURI) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoURI, repoURI, u.cloneURL, test.conn)
			}
		}
	}
}
