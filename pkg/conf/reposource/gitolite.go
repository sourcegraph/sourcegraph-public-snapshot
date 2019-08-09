package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Gitolite struct {
	*schema.GitoliteConnection
}

var _ RepoSource = Gitolite{}

func (c Gitolite) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, err := parseCloneURL(cloneURL)
	if err != nil {
		return "", err
	}
	parsedHostURL, err := parseCloneURL(c.Host)
	if err != nil {
		return "", err
	}
	if parsedHostURL.Hostname() != parsedCloneURL.Hostname() {
		return "", nil
	}
	return GitoliteRepoName(c.Prefix, strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

// GitoliteRepoName returns the Sourcegraph name for a repository given the Gitolite prefix (defined
// in the Gitolite external service config) and the Gitolite repository name. This is normally just
// the prefix concatenated with the Gitolite name. Gitolite permits the "@" character, but
// Sourcegraph does not, so "@" characters are rewritten to be "-".
func GitoliteRepoName(prefix, gitoliteName string) api.RepoName {
	gitoliteNameWithNoIllegalChars := strings.Replace(gitoliteName, "@", "-", -1)
	return api.RepoName(strings.NewReplacer(
		"{prefix}", prefix,
		"{gitoliteName}", gitoliteNameWithNoIllegalChars,
	).Replace("{prefix}{gitoliteName}"))
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_743(size int) error {
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
