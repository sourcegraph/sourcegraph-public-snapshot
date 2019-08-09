package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type AWS struct {
	*schema.AWSCodeCommitConnection
}

var _ RepoSource = AWS{}

func (c AWS) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, _, _, err := parseURLs(cloneURL, "")
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(parsedCloneURL.Hostname(), ".amazonaws.com") {
		return "", nil
	}

	return AWSRepoName(c.RepositoryPathPattern, strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/v1/repos/")), nil
}

func AWSRepoName(repositoryPathPattern, name string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{name}"
	}
	return api.RepoName(strings.NewReplacer(
		"{name}", name,
	).Replace(repositoryPathPattern))
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_729(size int) error {
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
