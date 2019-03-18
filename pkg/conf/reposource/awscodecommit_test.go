package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAWS_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		conn schema.AWSCodeCommitConnection
		urls []urlURI
	}{{
		conn: schema.AWSCodeCommitConnection{
			Region: "us-west-1",
		},
		urls: []urlURI{
			{"ssh://my-ssh-key-id@git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "test2"},

			{"https://user@bitbucket.org/gorilla/mux", ""},
			{"https://github.com/gorilla/mux", ""},
		},
	}, {
		conn: schema.AWSCodeCommitConnection{
			RepositoryPathPattern: "aws/{name}",
		},
		urls: []urlURI{
			{"ssh://my-ssh-key-id@git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},

			{"https://user@bitbucket.org/gorilla/mux", ""},
			{"https://github.com/gorilla/mux", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := AWS{&test.conn}.cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoURI != string(repoURI) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoURI, repoURI, u.cloneURL, test.conn)
			}
		}
	}
}
