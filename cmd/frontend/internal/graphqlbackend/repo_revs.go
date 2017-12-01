package graphqlbackend

import (
	"errors"
	"strings"
)

// repositoryRevisions specifies a repository and 0 or more revspecs.
// If no revspec is specified, then the repository's default branch is used.
type repositoryRevisions struct {
	repo     string
	revspecs []string
}

// parseRepositoryRevisions parses strings that refer to a repository and 0
// or more revspecs. The format is:
//
//   repo@revs
//
// where repo is a repository path and revs is a ':'-separated list of revspecs.
// The '@' and revs may be omitted to refer to the default branch.
//
// For example:
//
// - 'foo' refers to the 'foo' repo at the defaul branch
// - 'foo@bar' refers to the 'foo' repo and the 'bar' revspec.
// - 'foo@bar:baz:qux' refers to the 'foo' repo and 3 revspecs: 'bar', 'baz',
//    and 'qux'.
func parseRepositoryRevisions(repoAndOptionalRev string) repositoryRevisions {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		return repositoryRevisions{repo: repoAndOptionalRev}
	}
	return repositoryRevisions{
		repo:     repoAndOptionalRev[:i],
		revspecs: strings.Split(repoAndOptionalRev[i+1:], ":"),
	}
}

func (r *repositoryRevisions) String() string {
	if len(r.revspecs) > 0 {
		return r.repo + "@" + strings.Join(r.revspecs, ":")
	}
	return r.repo
}

func (r *repositoryRevisions) revSpecsOrDefaultBranch() []string {
	if len(r.revspecs) == 0 {
		return []string{""}
	}
	return r.revspecs
}

func (r *repositoryRevisions) hasSingleRevSpec() bool {
	return len(r.revspecs) == 1
}

var errMultipleRevSpecsNotSupported = errors.New("not yet supported: searching multiple revspecs in the same repo")
