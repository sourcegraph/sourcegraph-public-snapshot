package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Gitolite struct {
	*schema.GitoliteConnection
}

var _ repoSource = Gitolite{}

func (c Gitolite) cloneURLToRepoURI(cloneURL string) (repoURI api.RepoName, err error) {
	parsedCloneURL, err := parseCloneURL(cloneURL)
	if err != nil {
		return "", err
	}
	parsedHostURL, err := parseCloneURL(c.Host)
	if err != nil {
		return "", err
	}
	if parsedHostURL.Hostname() != parsedCloneURL.Hostname() {
		return "", nil
	}
	return GitoliteRepoURI(c.Prefix, strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")), nil
}

func GitoliteRepoURI(prefix, path string) api.RepoName {
	return api.RepoName(strings.NewReplacer(
		"{prefix}", prefix,
		"{path}", path,
	).Replace("{prefix}{path}"))
}
