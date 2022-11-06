package reposource

import (
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Local struct {
	*schema.LocalExternalServiceConnection
}

var _ RepoSource = Local{}

const DefaultLocalRepositoryPathPattern = "{root}/{repo}"

func (c Local) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	rel, err := filepath.Rel(c.LocalExternalServiceConnection.Root, cloneURL)
	if err != nil {
		return "", errors.Wrap(err, "cloneURL was not a subdirectory of root directory")
	}

	var repoPathPattern = c.RepositoryPathPattern
	if repoPathPattern == "" {
		repoPathPattern = DefaultLocalRepositoryPathPattern
	}
	return api.RepoName(strings.NewReplacer(
		"{root}", strings.TrimPrefix(c.Root, "/"),
		"{repo}", rel,
	).Replace(repoPathPattern)), nil
}
