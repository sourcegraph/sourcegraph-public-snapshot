package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type AWS struct {
	*schema.AWSCodeCommitConnection
}

var _ RepoSource = AWS{}

func (c AWS) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, _, _, err := parseURLs(cloneURL, "")
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(parsedCloneURL.Hostname(), ".amazonaws.com") {
		return "", nil
	}

	return AWSRepoName(c.RepositoryPathPattern, strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/v1/repos/")), nil
}

func AWSRepoName(repositoryPathPattern, name string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{name}"
	}
	return api.RepoName(strings.NewReplacer(
		"{name}", name,
	).Replace(repositoryPathPattern))
}
