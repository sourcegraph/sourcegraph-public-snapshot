package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitHub struct {
	*schema.GitHubConnection
}

var _ repoSource = GitHub{}

func (c GitHub) cloneURLToRepoURI(cloneURL string) (repoURI api.RepoURI, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}
	return GitHubRepoURI(c.RepositoryPathPattern, baseURL.Hostname(), strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

func GitHubRepoURI(repositoryPathPattern, host, nameWithOwner string) api.RepoURI {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{nameWithOwner}"
	}

	return api.RepoURI(strings.NewReplacer(
		"{host}", host,
		"{nameWithOwner}", nameWithOwner,
	).Replace(repositoryPathPattern))
}
