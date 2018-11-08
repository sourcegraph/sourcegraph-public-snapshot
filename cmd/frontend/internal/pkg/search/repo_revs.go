package search

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	dbquery "github.com/sourcegraph/sourcegraph/cmd/frontend/db/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
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

// RepoQuery takes a search query q and translates into a query for the
// repository database. If a repository could match q, it will be returned via
// Repos.List on the query.
func RepoQuery(q query.Q) (dbquery.Q, error) {
	// Given an expression we can convert it to dbquery with the following
	// observations:
	//
	//   RepoQuery((and A B)) == RepoQuery(A) AND RepoQuery(B)
	//   RepoQuery((or A B))  == RepoQuery(A) OR  RepoQuery(B)
	//   RepoQuery((type A))  == RepoQuery(A)
	//   RepoQuery((repo))    == repo.pattern
	//   RepoQuery((bool))    == bool
	//
	// However, query.Q can contain not nodes and atom nodes (such as
	// content:"foo"). We can't include atoms in our dbquery, since that
	// requires actually searching the code. So we want them to essentially
	// act in a way which never reduces the set of repositories that could be
	// found.
	//
	// For a simple query like (and "foo" r:bar) we can see that if we
	// substitute TRUE for "foo" the query simplifies to r:bar which is the
	// set of repos we should search. However, for (and (not "foo") r:bar) if
	// we substitute in TRUE we get:
	//
	//   (and (not TRUE) r:bar) == (and FALSE r:bar)
	//                          == FALSE
	//
	// But what we want is r:bar, since bar could contain matches which don't
	// contain "foo" so should be searched.
	//
	// The key insight here is the boolean to substitute in for an atom is
	// whatever simplifies to TRUE. Given not is the only node which can flip
	// a boolean, the boolean to substitute in is related to the number of not
	// nodes which are ancestors of the atom.
	//
	//   v         v			=> TRUE
	//   (not v)   v			=> FALSE
	//   (not (not v)) v		=> TRUE
	//   (not (not (not v))) v	=> FALSE
	//
	// This generalizes to substituting in AncestorNotCount % 2 == 0

	// Replace all atoms (except constants and repo) with nothing. We use
	// nothing to count the number of ancestor not nodes.
	var err error
	q = query.Map(q, nil, func(q query.Q) query.Q {
		switch c := q.(type) {
		case *query.Repo:
		case *query.Const:
		case *query.And:
		case *query.Or:

		case *query.RepoSet:
			err = errors.Errorf("unsupported RepoSet in RepoQuery %s", q.String())

		case *query.Type:
			return c.Child

		case *query.Not:
			return query.Map(q, nil, func(q query.Q) query.Q {
				if c, ok := q.(*nothing); ok {
					c.NotCount++
				}
				return q
			})

		default:
			return &nothing{}
		}
		return q
	})
	if err != nil {
		return nil, err
	}

	// Convert the nothing atoms into constants (since they have the correct
	// NotCount now).
	q = query.Map(q, nil, func(q query.Q) query.Q {
		if c, ok := q.(*nothing); ok {
			return &query.Const{Value: c.NotCount%2 == 0}
		}
		return q
	})

	// Constant fold
	q = query.Simplify(q)

	// We now have a query with only And, Or, Not, boolean constants and Repo
	// queries. This can now be translated into a Repos.List dbquery.
	return convertQuery(q), nil
}

func convertQuery(q query.Q) dbquery.Q {
	switch c := q.(type) {
	case *query.And:
		return dbquery.And(convertQueries(c.Children)...)

	case *query.Or:
		return dbquery.Or(convertQueries(c.Children)...)

	case *query.Not:
		return dbquery.Not(convertQuery(c.Child))

	case *query.Const:
		return c.Value

	case *query.Repo:
		return c.Pattern

	default:
		// We control the queries passed into convertQuery, so this shouldn't
		// happen.
		panic("unexpected: " + q.String())
	}
}

func convertQueries(qs []query.Q) []dbquery.Q {
	x := make([]dbquery.Q, 0, len(qs))
	for _, q := range qs {
		x = append(x, convertQuery(q))
	}
	return x
}

// nothing is a type which should not affect the outcome of a query. In
// RepoQuery it is used to replace all non-repo atoms. NotCount is stored
// since it is used to determine which value has no effect on the query.
type nothing struct {
	// NotCount is the number of ancestors which are a Not node.
	NotCount int
}

func (q *nothing) String() string {
	return fmt.Sprintf("N(%d)", q.NotCount)
}
