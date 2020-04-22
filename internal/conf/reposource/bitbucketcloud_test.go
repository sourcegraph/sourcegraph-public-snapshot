package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketCloud_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.BitbucketCloudConnection
		urls []urlToRepoName
	}{
		{
			conn: schema.BitbucketCloudConnection{
				Url: "https://bitbucket.org",
			},
			urls: []urlToRepoName{
				{"git@bitbucket.org:gorilla/mux.git", "bitbucket.org/gorilla/mux"},
				{"git@bitbucket.org:/gorilla/mux.git", "bitbucket.org/gorilla/mux"},
				{"git+https://bitbucket.org/gorilla/mux.git", "bitbucket.org/gorilla/mux"},
				{"https://bitbucket.org/gorilla/mux.git", "bitbucket.org/gorilla/mux"},
				{"https://www.bitbucket.org/gorilla/mux.git", "bitbucket.org/gorilla/mux"},
				{"https://oauth2:ACCESS_TOKEN@bitbucket.org/gorilla/mux.git", "bitbucket.org/gorilla/mux"},

				{"git@asdf.com:gorilla/mux.git", ""},
				{"https://asdf.com/gorilla/mux.git", ""},
				{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
			},
		},
		{
			conn: schema.BitbucketCloudConnection{
				Url: "https://staging.bitbucket.org",
			},
			urls: []urlToRepoName{
				{"git@staging.bitbucket.org:gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},
				{"git@staging.bitbucket.org:/gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},
				{"git+https://staging.bitbucket.org/gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},
				{"https://staging.bitbucket.org/gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},
				{"https://www.staging.bitbucket.org/gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},
				{"https://oauth2:ACCESS_TOKEN@staging.bitbucket.org/gorilla/mux.git", "staging.bitbucket.org/gorilla/mux"},

				{"git@asdf.com:gorilla/mux.git", ""},
				{"https://asdf.com/gorilla/mux.git", ""},
				{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
			},
		},
	}

	for _, test := range tests {
		for _, u := range test.urls {
			repoName, err := BitbucketCloud{&test.conn}.CloneURLToRepoName(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoName != string(repoName) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoName, repoName, u.cloneURL, test.conn)
			}
		}
	}
}
