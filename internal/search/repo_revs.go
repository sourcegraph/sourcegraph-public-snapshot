package search

import (
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// RepositoryRevisions specifies a repository and 0 or more revspecs and ref
// globs.  If no revspecs and no ref globs are specified, then the
// repository's default branch is used.
type RepositoryRevisions struct {
	Repo types.MinimalRepo
	Revs []string
}

func (r *RepositoryRevisions) Copy() *RepositoryRevisions {
	revs := make([]string, len(r.Revs))
	copy(revs, r.Revs)
	return &RepositoryRevisions{
		Repo: r.Repo,
		Revs: revs,
	}
}

// Equal provides custom comparison which is used by go-cmp
func (r *RepositoryRevisions) Equal(other *RepositoryRevisions) bool {
	return reflect.DeepEqual(r.Repo, other.Repo) && reflect.DeepEqual(r.Revs, other.Revs)
}

// GitserverRepo is a convenience function to return the api.RepoName for
// r.Repo. The returned Repo will not have the URL set, only the name.
func (r *RepositoryRevisions) GitserverRepo() api.RepoName {
	return r.Repo.Name
}

func (r *RepositoryRevisions) String() string {
	if len(r.Revs) == 0 {
		return string(r.Repo.Name)
	}

	return string(r.Repo.Name) + "@" + strings.Join(r.Revs, ":")
}
