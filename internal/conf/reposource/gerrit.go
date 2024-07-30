package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Gerrit struct {
	*schema.GerritConnection
}

var _ RepoSource = Gerrit{}

func (c Gerrit) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}
	name := strings.TrimSuffix(parsedCloneURL.Path, ".git")
	name = strings.TrimPrefix(name, "/")
	// Gerrit clone URLs can contain `/a/` prefixes which means "authenticated clone".
	name = strings.TrimPrefix(name, "a/")
	return GerritRepoName(c.RepositoryPathPattern, baseURL.Hostname(), name), nil
}

func GerritRepoName(repositoryPathPattern, host, name string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{name}"
	}

	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{name}", name,
	).Replace(repositoryPathPattern))
}
