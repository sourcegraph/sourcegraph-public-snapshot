package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitHub_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.GitHubConnection
		urls []urlToRepoName
	}{{
		conn: schema.GitHubConnection{
			Url: "https://github.com",
		},
		urls: []urlToRepoName{
			{"git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
			{"git@github.com:/gorilla/mux.git", "github.com/gorilla/mux"},
			{"git+https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://www.github.com/gorilla/mux.git", "github.com/gorilla/mux"},
			{"https://oauth2:ACCESS_TOKEN@github.com/gorilla/mux.git", "github.com/gorilla/mux"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}, {
		conn: schema.GitHubConnection{
			Url:                   "https://github.mycompany.com",
			RepositoryPathPattern: "{nameWithOwner}",
		},
		urls: []urlToRepoName{
			{"git@github.mycompany.com:foo/bar/baz.git", "foo/bar/baz"},
			{"https://github.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},
			{"https://oauth2:ACCESS_TOKEN@github.mycompany.com/foo/bar/baz.git", "foo/bar/baz"},

			{"git@asdf.com:gorilla/mux.git", ""},
			{"https://asdf.com/gorilla/mux.git", ""},
			{"https://oauth2:ACCESS_TOKEN@asdf.com/gorilla/mux.git", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoName, err := GitHub{&test.conn}.CloneURLToRepoName(u.cloneURL)
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
func random_740(size int) error {
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
