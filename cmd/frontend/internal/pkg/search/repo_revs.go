package search

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// RevisionSpecifier represents either a revspec or a ref glob. At most one
// field is set. The default branch is represented by all fields being empty.
type RevisionSpecifier struct {
	// RevSpec is a revision range specifier suitable for passing to git. See
	// the manpage gitrevisions(7).
	RevSpec string

	// RefGlob is a reference glob to pass to git. See the documentation for
	// "--glob" in git-log.
	RefGlob string

	// ExcludeRefGlob is a glob for references to exclude. See the
	// documentation for "--exclude" in git-log.
	ExcludeRefGlob string
}

func (r RevisionSpecifier) String() string {
	if r.ExcludeRefGlob != "" {
		return "*!" + r.ExcludeRefGlob
	}
	if r.RefGlob != "" {
		return "*" + r.RefGlob
	}
	return r.RevSpec
}

// Less compares two revspecOrRefGlob entities, suitable for use
// with sort.Slice()
//
// possibly-undesired: this results in treating an entity with
// no revspec, but a refGlob, as "earlier" than any revspec.
func (r1 RevisionSpecifier) Less(r2 RevisionSpecifier) bool {
	if r1.RevSpec != r2.RevSpec {
		return r1.RevSpec < r2.RevSpec
	}
	if r1.RefGlob != r2.RefGlob {
		return r1.RefGlob < r2.RefGlob
	}
	return r1.ExcludeRefGlob < r2.ExcludeRefGlob
}

// RepositoryRevisions specifies a repository and 0 or more revspecs and ref
// globs.  If no revspecs and no ref globs are specified, then the
// repository's default branch is used.
type RepositoryRevisions struct {
	Repo *types.Repo
	Revs []RevisionSpecifier

	GitserverRepo gitserver.Repo // URL field is optional (see (gitserver.ExecRequest).URL field for behavior)
}

// ParseRepositoryRevisions parses strings that refer to a repository and 0
// or more revspecs. The format is:
//
//   repo@revs
//
// where repo is a repository path and revs is a ':'-separated list of revspecs
// and/or ref globs. A ref glob is a revspec prefixed with '*' (which is not a
// valid revspec or ref itself; see `man git-check-ref-format`). The '@' and revs
// may be omitted to refer to the default branch.
//
// For example:
//
// - 'foo' refers to the 'foo' repo at the default branch
// - 'foo@bar' refers to the 'foo' repo and the 'bar' revspec.
// - 'foo@bar:baz:qux' refers to the 'foo' repo and 3 revspecs: 'bar', 'baz',
//   and 'qux'.
// - 'foo@*bar' refers to the 'foo' repo and all refs matching the glob 'bar/*',
//   because git interprets the ref glob 'bar' as being 'bar/*' (see `man git-log`
//   section on the --glob flag)
func ParseRepositoryRevisions(repoAndOptionalRev string) (api.RepoURI, []RevisionSpecifier) {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		// return an empty slice to indicate that there's no revisions; callers
		// have to distinguish between "none specified" and "default" to handle
		// cases where two repo specs both match the same repository, and only one
		// specifies a revspec, which normally implies "master" but in that case
		// really means "didn't specify"
		return api.RepoURI(repoAndOptionalRev), []RevisionSpecifier{}
	}

	repo := api.RepoURI(repoAndOptionalRev[:i])
	var revs []RevisionSpecifier
	for _, part := range strings.Split(repoAndOptionalRev[i+1:], ":") {
		if part == "" {
			continue
		}
		var rev RevisionSpecifier
		if strings.HasPrefix(part, "*!") {
			rev.ExcludeRefGlob = part[2:]
		} else if strings.HasPrefix(part, "*") {
			rev.RefGlob = part[1:]
		} else {
			rev.RevSpec = part
		}
		revs = append(revs, rev)
	}
	if len(revs) == 0 {
		revs = []RevisionSpecifier{{RevSpec: ""}} // default branch
	}
	return repo, revs
}

func (r RepositoryRevisions) String() string {
	if len(r.Revs) == 0 {
		return string(r.Repo.URI)
	}

	parts := make([]string, len(r.Revs))
	for i, rev := range r.Revs {
		parts[i] = rev.String()
	}
	return string(r.Repo.URI) + "@" + strings.Join(parts, ":")
}

func (r *RepositoryRevisions) RevSpecs() []string {
	var revspecs []string
	for _, rev := range r.Revs {
		revspecs = append(revspecs, rev.RevSpec)
	}
	return revspecs
}
