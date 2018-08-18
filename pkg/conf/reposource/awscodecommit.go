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

func (c AWS) CloneURLToRepoURI(cloneURL string) (repoURI api.RepoURI, err error) {
	parsedCloneURL, _, _, err := parseURLs(cloneURL, "")
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(parsedCloneURL.Hostname(), ".amazonaws.com") {
		return "", nil
	}

	return AWSRepoURI(c.RepositoryPathPattern, strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/v1/repos/")), nil
}

func AWSRepoURI(repositoryPathPattern, name string) api.RepoURI {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{name}"
	}
	return api.RepoURI(strings.NewReplacer(
		"{name}", name,
	).Replace(repositoryPathPattern))
}
