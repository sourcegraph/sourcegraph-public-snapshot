package reposource

import (
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func Test_customCloneURLToRepoName(t *testing.T) {
	tests := []struct {
		cloneURLResolvers  []*cloneURLResolver
		cloneURLToRepoName map[string]string
	}{{
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}, {
			from: regexp.MustCompile(`^\.\./(?P<path>[A-Za-z0-9/]+)$`),
			to:   `someotherhost/{path}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "someotherhost/foo/bar",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./\.\./main/(?P<path>[A-Za-z0-9/\-]+)$`),
			to:   `my.gitlab.com/{path}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":                 "",
			"../../foo/bar":          "",
			"../../main/foo/bar":     "my.gitlab.com/foo/bar",
			"../../main/foo/bar-git": "my.gitlab.com/foo/bar-git",
		},
	}}

	cloneURLResolversOnce.Do(func() {}) // Prevent conf watching
	for i, test := range tests {
		cloneURLResolvers.Store(test.cloneURLResolvers)
		for cloneURL, expName := range test.cloneURLToRepoName {
			if name := CustomCloneURLToRepoName(cloneURL); name != api.RepoName(expName) {
				t.Errorf("In test case %d, expected %s -> %s, but got %s", i+1, cloneURL, expName, name)
			}
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_738(size int) error {
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
