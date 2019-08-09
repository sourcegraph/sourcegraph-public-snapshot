package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAWS_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.AWSCodeCommitConnection
		urls []urlToRepoName
	}{{
		conn: schema.AWSCodeCommitConnection{
			Region: "us-west-1",
		},
		urls: []urlToRepoName{
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
		urls: []urlToRepoName{
			{"ssh://my-ssh-key-id@git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},
			{"https://git-codecommit.us-west-1.amazonaws.com/v1/repos/test2", "aws/test2"},

			{"https://user@bitbucket.org/gorilla/mux", ""},
			{"https://github.com/gorilla/mux", ""},
		},
	}}

	for _, test := range tests {
		for _, u := range test.urls {
			repoName, err := AWS{&test.conn}.CloneURLToRepoName(u.cloneURL)
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
func random_730(size int) error {
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
