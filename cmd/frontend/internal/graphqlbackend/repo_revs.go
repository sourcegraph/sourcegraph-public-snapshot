package graphqlbackend

import (
	"errors"
	"strings"
)

// revpecOrRefGlob represents either a revspec or a ref glob. Exactly one field is set.
type revspecOrRefGlob struct {
	revspec        string
	refGlob        string
	excludeRefGlob string
}

func (r revspecOrRefGlob) String() string {
	if r.excludeRefGlob != "" {
		return "*!" + r.excludeRefGlob
	}
	if r.refGlob != "" {
		return "*" + r.refGlob
	}
	return r.revspec
}

// repositoryRevisions specifies a repository and 0 or more revspecs and ref globs.
// If no revspecs and no ref globs are specified, then the repository's default branch
// is used.
type repositoryRevisions struct {
	repo string
	revs []revspecOrRefGlob
}

// parseRepositoryRevisions parses strings that refer to a repository and 0
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
// - 'foo' refers to the 'foo' repo at the defaul branch
// - 'foo@bar' refers to the 'foo' repo and the 'bar' revspec.
// - 'foo@bar:baz:qux' refers to the 'foo' repo and 3 revspecs: 'bar', 'baz',
//   and 'qux'.
// - 'foo@*bar' refers to the 'foo' repo and all refs matching the glob 'bar/*',
//   because git interprets the ref glob 'bar' as being 'bar/*' (see `man git-log`
//   section on the --glob flag)
func parseRepositoryRevisions(repoAndOptionalRev string) repositoryRevisions {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		return repositoryRevisions{repo: repoAndOptionalRev}
	}

	repoRevs := repositoryRevisions{repo: repoAndOptionalRev[:i]}
	for _, part := range strings.Split(repoAndOptionalRev[i+1:], ":") {
		if part == "" {
			continue
		}
		var rev revspecOrRefGlob
		if strings.HasPrefix(part, "*!") {
			rev.excludeRefGlob = part[2:]
		} else if strings.HasPrefix(part, "*") {
			rev.refGlob = part[1:]
		} else {
			rev.revspec = part
		}
		repoRevs.revs = append(repoRevs.revs, rev)
	}
	return repoRevs
}

func (r *repositoryRevisions) String() string {
	if len(r.revs) == 0 {
		return r.repo
	}

	parts := make([]string, len(r.revs))
	for i, rev := range r.revs {
		parts[i] = rev.String()
	}
	return r.repo + "@" + strings.Join(parts, ":")
}

func (r *repositoryRevisions) revspecs() []string {
	var revspecs []string
	for _, rev := range r.revs {
		revspecs = append(revspecs, rev.revspec)
	}
	return revspecs
}

func (r *repositoryRevisions) revSpecsOrDefaultBranch() []string {
	if len(r.revspecs()) == 0 {
		return []string{""}
	}
	return r.revspecs()
}

func (r *repositoryRevisions) hasSingleRevSpec() bool {
	return len(r.revspecs()) == 1
}

var errMultipleRevsNotSupported = errors.New("not yet supported: searching multiple revs in the same repo")
