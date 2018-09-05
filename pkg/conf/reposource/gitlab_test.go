package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitLab_cloneURLToRepoURI(t *testing.T) {
	var tests = []struct {
		conn schema.GitLabConnection
		urls []urlURI
	}{{
		conn: schema.GitLabConnection{
			Url: "https://gitlab.com",
		},
		urls: []urlURI{
			{"git@gitlab.com:beyang/public-repo.git", "gitlab.com/beyang/public-repo"},
			{"git@gitlab.com:/beyang/public-repo.git", "gitlab.com/beyang/public-repo"},
			{"https://gitlab.com/beyang/public-repo.git", "gitlab.com/beyang/public-repo"},
			{"https://oauth2:ACCESS_TOKEN@gitlab.com/beyang/public-repo.git", "gitlab.com/beyang/public-repo"},

			{"git@asdf.com:beyang/public-repo.git", ""},
			{"https://asdf.com/beyang/public-repo.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/beyang/public-repo.git", ""},
		},
	}, {
		conn: schema.GitLabConnection{
			Url:                   "https://gitlab.mycompany.com",
			RepositoryPathPattern: "{pathWithNamespace}",
		},
		urls: []urlURI{
			{"git@gitlab.mycompany.com:foo/bar/baz.git", "foo/bar/baz"},
			{"https://gitlab.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},
			{"https://oauth2:ACCESS_TOKEN@gitlab.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},

			{"git@asdf.com:beyang/public-repo.git", ""},
			{"https://asdf.com/beyang/public-repo.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/beyang/public-repo.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoURI, err := GitLab{&test.conn}.cloneURLToRepoURI(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoURI != string(repoURI) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoURI, repoURI, u.cloneURL, test.conn)
			}
		}
	}
}
