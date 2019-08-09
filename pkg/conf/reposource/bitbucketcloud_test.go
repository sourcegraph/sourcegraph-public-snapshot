package reposource

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketCloud_cloneURLToRepoName(t *testing.T) {
	tests := []struct {
		conn schema.BitbucketCloudConnection
		urls []urlToRepoName
	}{{
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
	}}

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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_732(size int) error {
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
