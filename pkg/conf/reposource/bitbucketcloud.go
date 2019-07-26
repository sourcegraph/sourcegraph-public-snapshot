package reposource

import (
	"strings"

	"sourcegraph.com/pkg/api"
	"sourcegraph.com/schema"
)

type BitbucketCloud struct {
	*schema.BitbucketCloudConnection
}

var _ RepoSource = BitbucketCloud{}

func (c BitbucketCloud) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}
	return BitbucketCloudRepoName(c.RepositoryPathPattern, baseURL.Hostname(), strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

func BitbucketCloudRepoName(repositoryPathPattern, host, nameWithOwner string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{nameWithOwner}"
	}

	return api.RepoName(strings.NewReplacer(
		"{host}", host,
		"{nameWithOwner}", nameWithOwner,
	).Replace(repositoryPathPattern))
}
