package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketServer_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.BitbucketServerConnection
		urls []urlToRepoName
	}{{
		conn: schema.BitbucketServerConnection{
			Password: "pass",
			Url:      "https://bitbucket.sgdev.org",
			Username: "user",
		},
		urls: []urlToRepoName{
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
		urls: []urlToRepoName{
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
			repoName, err := BitbucketServer{&test.conn}.CloneURLToRepoName(u.cloneURL)
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
func random_734(size int) error {
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
