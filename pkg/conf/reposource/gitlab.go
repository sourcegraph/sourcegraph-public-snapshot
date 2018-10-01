package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitLab struct {
	*schema.GitLabConnection
}

var _ repoSource = GitLab{}

func (c GitLab) cloneURLToRepoURI(cloneURL string) (repoURI api.RepoURI, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}

	pathWithNamespace := strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")
	return GitLabRepoURI(c.RepositoryPathPattern, baseURL.Hostname(), pathWithNamespace), nil
}

func GitLabRepoURI(repositoryPathPattern, host, pathWithNamespace string) api.RepoURI {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{pathWithNamespace}"
	}

	return api.RepoURI(strings.NewReplacer(
		"{host}", host,
		"{pathWithNamespace}", pathWithNamespace,
	).Replace(repositoryPathPattern))
}
