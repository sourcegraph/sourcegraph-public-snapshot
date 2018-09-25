package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketServer struct {
	*schema.BitbucketServerConnection
}

var _ repoSource = BitbucketServer{}

func (c BitbucketServer) cloneURLToRepoURI(cloneURL string) (repoURI api.RepoURI, err error) {
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

	return BitbucketServerRepoURI(c.RepositoryPathPattern, baseURL.Hostname(), proj, rp), nil
}

func BitbucketServerRepoURI(repositoryPathPattern, host, projectKey, repoSlug string) api.RepoURI {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{projectKey}/{repositorySlug}"
	}
	return api.RepoURI(strings.NewReplacer(
		"{host}", host,
		"{projectKey}", projectKey,
		"{repositorySlug}", repoSlug,
	).Replace(repositoryPathPattern))
}
