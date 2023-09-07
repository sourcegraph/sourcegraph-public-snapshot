package types

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RepositoryRevSpec represents zero or more revisions we need to search in a
// repository for a revision specifier. This can be inferred relatively
// cheaply from parsing a query and the repos table.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
//
// Note: this is like a search/repos.RepoRevSpecs except we store 1 revision
// specifier per repository. It may be worth updating this to instead store a
// slice of RevisionSpecifiers.
type RepositoryRevSpec struct {
	// Repository is the repository to search.
	Repository api.RepoID

	// RevisionSpecifier is something that still needs to be resolved on gitserver. This
	// is a serialiazed version of query.RevisionSpecifier.
	RevisionSpecifier string
}

func (r RepositoryRevSpec) String() string {
	return fmt.Sprintf("RepositoryRevSpec{%d@%s}", r.Repository, r.RevisionSpecifier)
}

// RepositoryRevision represents the smallest unit we can search over, a
// specific repository and revision.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
type RepositoryRevision struct {
	// RepositoryRevSpec is where this RepositoryRevision got resolved from.
	RepositoryRevSpec

	// Revision is a resolved revision specifier. eg HEAD, branch-name,
	// commit-hash, etc.
	Revision string
}

func (r RepositoryRevision) String() string {
	return fmt.Sprintf("RepositoryRevision{%d@%s}", r.Repository, r.Revision)
}
