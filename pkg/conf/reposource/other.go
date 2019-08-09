package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type urlMismatchErr struct {
	cloneURL string
	hostURL  string
}

func (e urlMismatchErr) Error() string {
	return fmt.Sprintf("cloneURL %q did not match git host %q", e.cloneURL, e.hostURL)
}

type Other struct {
	*schema.OtherExternalServiceConnection
}

var _ RepoSource = Other{}

const DefaultRepositoryPathPattern = "{base}/{repo}"

func (c Other) CloneURLToRepoURI(cloneURL string) (string, error) {
	return cloneURLToRepoName(cloneURL, c.Url, DefaultRepositoryPathPattern)
}

func (c Other) CloneURLToRepoName(cloneURL string) (api.RepoName, error) {
	repoName, err := cloneURLToRepoName(cloneURL, c.Url, c.RepositoryPathPattern)
	return api.RepoName(repoName), err
}

func cloneURLToRepoName(cloneURL, baseURL, repositoryPathPattern string) (string, error) {
	parsedCloneURL, parsedBaseURL, match, err := parseURLs(cloneURL, baseURL)
	if err != nil {
		return "", err
	}
	if !match {
		return "", urlMismatchErr{cloneURL: cloneURL, hostURL: baseURL}
	}

	basePrefix := parsedBaseURL.Path
	if !strings.HasSuffix(basePrefix, "/") {
		basePrefix = basePrefix + "/"
	}
	if parsedCloneURL.Path != parsedBaseURL.Path && !strings.HasPrefix(parsedCloneURL.Path, basePrefix) {
		return "", urlMismatchErr{cloneURL: cloneURL, hostURL: baseURL}
	}
	relativeRepoPath := strings.TrimPrefix(parsedCloneURL.Path, basePrefix)

	base := url.URL{
		Host: parsedBaseURL.Host,
		Path: parsedBaseURL.Path,
	}
	return OtherRepoName(repositoryPathPattern, base.String(), relativeRepoPath), nil
}

var otherRepoNameReplacer = strings.NewReplacer(":", "-", "@", "-", "//", "")

func OtherRepoName(repositoryPathPattern, base, relativeRepoPath string) string {
	if repositoryPathPattern == "" {
		repositoryPathPattern = DefaultRepositoryPathPattern
	}
	return strings.NewReplacer(
		"{base}", otherRepoNameReplacer.Replace(strings.TrimSuffix(base, "/")),
		"{repo}", otherRepoNameReplacer.Replace(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(relativeRepoPath, "/"), ".git"), "/")),
	).Replace(repositoryPathPattern)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_745(size int) error {
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
