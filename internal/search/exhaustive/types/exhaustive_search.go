package types

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RevisionSpecifiers is something that still needs to be resolved on
// gitserver. This is a serialized version of []query.RevisionSpecifier.
//
// We need to store a list since specifiers interact. For example a glob
// pattern can be coupled with a negative glob pattern.
type RevisionSpecifiers string

func (s RevisionSpecifiers) String() string {
	return string(s)
}

// Get returns the marshalled version of []query.RevisionSpecifier
func (s RevisionSpecifiers) Get() []string {
	// : is the same seperator we use in our query language.
	return strings.Split(string(s), ":")
}

// RevisionSpecifierJoin is the inverse of RevisionSpecifiers.Get(). It can be
// used to convert a []query.RevisionSpecifier into a RevisionSpecifiers.
func RevisionSpecifierJoin(s []string) RevisionSpecifiers {
	return RevisionSpecifiers(strings.Join(s, ":"))
}

// RepositoryRevSpecs represents zero or more revisions we need to search in a
// repository for a revision specifier. This can be inferred relatively
// cheaply from parsing a query and the repos table.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
//
// Note: this is serializable version of search/repos.RepoRevSpecs.
type RepositoryRevSpecs struct {
	// Repository is the repository to search.
	Repository api.RepoID

	// RevisionSpecifiers is something that still needs to be resolved on
	// gitserver. This is a serialized version of query.RevisionSpecifier.
	RevisionSpecifiers RevisionSpecifiers
}

func (r RepositoryRevSpecs) String() string {
	return fmt.Sprintf("RepositoryRevSpec{%d@%s}", r.Repository, r.RevisionSpecifiers)
}

// RepositoryRevision represents the smallest unit we can search over, a
// specific repository and revision.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
type RepositoryRevision struct {
	// RepositoryRevSpecs is where this RepositoryRevision got resolved from.
	RepositoryRevSpecs

	// Revision is a resolved revision specifier. eg HEAD, branch-name,
	// commit-hash, etc.
	Revision string
}

func (r RepositoryRevision) String() string {
	return fmt.Sprintf("RepositoryRevision{%d@%s}", r.Repository, r.Revision)
}

type RepoRevJobStats struct {
	Total      int32
	Completed  int32
	Failed     int32
	InProgress int32
}
