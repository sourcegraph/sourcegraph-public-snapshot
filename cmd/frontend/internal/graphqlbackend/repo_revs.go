package graphqlbackend

import (
	"errors"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// revpecOrRefGlob represents either a revspec or a ref glob. At most one field is set. The default
// branch is represented by all fields being empty.
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
	repo          *types.Repo
	gitserverRepo gitserver.Repo // URL field is optional (see (gitserver.ExecRequest).URL field for behavior)
	revs          []revspecOrRefGlob
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
// - 'foo' refers to the 'foo' repo at the default branch
// - 'foo@bar' refers to the 'foo' repo and the 'bar' revspec.
// - 'foo@bar:baz:qux' refers to the 'foo' repo and 3 revspecs: 'bar', 'baz',
//   and 'qux'.
// - 'foo@*bar' refers to the 'foo' repo and all refs matching the glob 'bar/*',
//   because git interprets the ref glob 'bar' as being 'bar/*' (see `man git-log`
//   section on the --glob flag)
func parseRepositoryRevisions(repoAndOptionalRev string) (api.RepoURI, []revspecOrRefGlob) {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		return api.RepoURI(repoAndOptionalRev), []revspecOrRefGlob{{revspec: ""}} // default branch
	}

	repo := api.RepoURI(repoAndOptionalRev[:i])
	var revs []revspecOrRefGlob
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
		revs = append(revs, rev)
	}
	if len(revs) == 0 {
		revs = []revspecOrRefGlob{{revspec: ""}} // default branch
	}
	return repo, revs
}

func (r repositoryRevisions) String() string {
	if len(r.revs) == 0 {
		return string(r.repo.URI)
	}

	parts := make([]string, len(r.revs))
	for i, rev := range r.revs {
		parts[i] = rev.String()
	}
	return string(r.repo.URI) + "@" + strings.Join(parts, ":")
}

func (r *repositoryRevisions) revspecs() []string {
	var revspecs []string
	for _, rev := range r.revs {
		revspecs = append(revspecs, rev.revspec)
	}
	return revspecs
}

func (r *repositoryRevisions) hasSingleRevSpec() bool {
	return len(r.revspecs()) == 1
}

var errMultipleRevsNotSupported = errors.New("not yet supported: searching multiple revs in the same repo")
