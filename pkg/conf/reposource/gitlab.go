package reposource

import (
	"strings"

	"sourcegraph.com/pkg/api"
	"sourcegraph.com/schema"
)

type GitLab struct {
	*schema.GitLabConnection
}

var _ RepoSource = GitLab{}

func (c GitLab) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}

	pathWithNamespace := strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")
	return GitLabRepoName(c.RepositoryPathPattern, baseURL.Hostname(), pathWithNamespace), nil
}

func GitLabRepoName(repositoryPathPattern, host, pathWithNamespace string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{pathWithNamespace}"
	}

	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{pathWithNamespace}", pathWithNamespace,
	).Replace(repositoryPathPattern))
}
