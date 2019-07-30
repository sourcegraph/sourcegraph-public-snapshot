package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GeneralProtocol struct {
	*schema.GeneralProtocolConnection
}

var _ RepoSource = GeneralProtocol{}

func (c GeneralProtocol) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}
	return GeneralProtocolRepoName(c.RepositoryPathPattern, baseURL.Hostname(), strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

func GeneralProtocolRepoName(repositoryPathPattern, host, nameWithOwner string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{nameWithOwner}"
	}

	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{nameWithOwner}", nameWithOwner,
	).Replace(repositoryPathPattern))
}
