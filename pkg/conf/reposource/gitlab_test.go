package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitLab_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.GitLabConnection
		urls []urlToRepoName
	}{{
		conn: schema.GitLabConnection{
			Url: "https://gitlab.com",
		},
		urls: []urlToRepoName{
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
		urls: []urlToRepoName{
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
			repoName, err := GitLab{&test.conn}.CloneURLToRepoName(u.cloneURL)
			if err != nil {
				t.Fatal(err)
			}
			if u.repoName != string(repoName) {
				t.Errorf("expected %q but got %q for clone URL %q (connection: %+v)", u.repoName, repoName, u.cloneURL, test.conn)
			}
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_742(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
