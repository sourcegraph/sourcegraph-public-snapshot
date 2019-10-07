package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitHub struct {
	*schema.GitHubConnection
}

var _ RepoSource = GitHub{}

func (c GitHub) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}
	return GitHubRepoName(c.RepositoryPathPattern, baseURL.Hostname(), strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

func GitHubRepoName(repositoryPathPattern, host, nameWithOwner string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{nameWithOwner}"
	}

	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{nameWithOwner}", nameWithOwner,
	).Replace(repositoryPathPattern))
}
