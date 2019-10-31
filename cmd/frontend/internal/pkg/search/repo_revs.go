package search

import (
	"reflect"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func (r1 RevisionSpecifier) String() string {
	if r1.ExcludeRefGlob != "" {
		return "*!" + r1.ExcludeRefGlob
	}
	if r1.RefGlob != "" {
		return "*" + r1.RefGlob
	}
	return r1.RevSpec
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
	// IndexedHEADCommit contains the HEAD Git commit indexed by Zoekt.
	// It is written to by zoektIndexedRepos and read later by zoektSearchHEAD.
	// See https://github.com/sourcegraph/sourcegraph/pull/4702 for the performance
	// rationale.
	mu                sync.Mutex
	indexedHEADCommit api.CommitID
}

func (r *RepositoryRevisions) Equal(other *RepositoryRevisions) bool {
	return reflect.DeepEqual(r.Repo, other.Repo) && reflect.DeepEqual(r.Revs, other.Revs)
}

func (r *RepositoryRevisions) IndexedHEADCommit() api.CommitID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.indexedHEADCommit
}

func (r *RepositoryRevisions) SetIndexedHEADCommit(ihc api.CommitID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.indexedHEADCommit = ihc
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
func ParseRepositoryRevisions(repoAndOptionalRev string) (api.RepoName, []RevisionSpecifier) {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		// return an empty slice to indicate that there's no revisions; callers
		// have to distinguish between "none specified" and "default" to handle
		// cases where two repo specs both match the same repository, and only one
		// specifies a revspec, which normally implies "master" but in that case
		// really means "didn't specify"
		return api.RepoName(repoAndOptionalRev), []RevisionSpecifier{}
	}

	repo := api.RepoName(repoAndOptionalRev[:i])
	var revs []RevisionSpecifier
	for _, part := range strings.Split(repoAndOptionalRev[i+1:], ":") {
		if part == "" {
			continue
		}
		revs = append(revs, parseRev(part))
	}
	if len(revs) == 0 {
		revs = []RevisionSpecifier{{RevSpec: ""}} // default branch
	}
	return repo, revs
}

func parseRev(spec string) RevisionSpecifier {
	if strings.HasPrefix(spec, "*!") {
		return RevisionSpecifier{ExcludeRefGlob: spec[2:]}
	} else if strings.HasPrefix(spec, "*") {
		return RevisionSpecifier{RefGlob: spec[1:]}
	}
	return RevisionSpecifier{RevSpec: spec}
}

// GitserverRepo is a convenience function to return the gitserver.Repo for
// r.Repo. The returned Repo will not have the URL set, only the name.
func (r *RepositoryRevisions) GitserverRepo() gitserver.Repo {
	return gitserver.Repo{Name: r.Repo.Name}
}

func (r *RepositoryRevisions) String() string {
	if len(r.Revs) == 0 {
		return string(r.Repo.Name)
	}

	parts := make([]string, len(r.Revs))
	for i, rev := range r.Revs {
		parts[i] = rev.String()
	}
	return string(r.Repo.Name) + "@" + strings.Join(parts, ":")
}

func (r *RepositoryRevisions) RevSpecs() []string {
	var revspecs []string
	for _, rev := range r.Revs {
		revspecs = append(revspecs, rev.RevSpec)
	}
	return revspecs
}
