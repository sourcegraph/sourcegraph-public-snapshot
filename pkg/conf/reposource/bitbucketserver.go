package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketServer struct {
	*schema.BitbucketServerConnection
}

var _ RepoSource = BitbucketServer{}

func (c BitbucketServer) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}

	var projAndRepo string
	if parsedCloneURL.Scheme == "ssh" {
		projAndRepo = strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")
	} else if strings.HasPrefix(parsedCloneURL.Scheme, "http") {
		projAndRepo = strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/scm/")
	}
	idx := strings.Index(projAndRepo, "/")
	if idx < 0 || len(projAndRepo)-1 == idx { // Not a Bitbucket Server clone URL
		return "", nil
	}
	proj, rp := projAndRepo[:idx], projAndRepo[idx+1:]

	return BitbucketServerRepoName(c.RepositoryPathPattern, baseURL.Hostname(), proj, rp), nil
}

func BitbucketServerRepoName(repositoryPathPattern, host, projectKey, repoSlug string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{projectKey}/{repositorySlug}"
	}
	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{projectKey}", projectKey,
		"{repositorySlug}", repoSlug,
	).Replace(repositoryPathPattern))
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_733(size int) error {
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
