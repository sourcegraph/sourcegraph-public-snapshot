package search

import (
	"regexp"
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
func (r RepositoryRevisions) GitserverRepo() gitserver.Repo {
	return gitserver.Repo{Name: r.Repo.Name}
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

// RepoRevisionsQuery evaulates ref specifiers in q to find out which
// revisions need to be searched for each repository.
func RepoRevisionsQuery(q query.Q, repos []*types.Repo) ([]RepositoryRevisions, error) {
	// If we have no ref specifiers, then we can shortcut this evaluation and
	// just search default branch.
	hasRef := false
	hasNotRef := false
	notCount := 0
	query.Map(q, func(q query.Q) query.Q {
		switch q.(type) {
		case *query.Not:
			notCount++
		case *query.Ref:
			hasRef = true
			if notCount > 0 {
				hasNotRef = true
			}
		}
		return q
	}, func(q query.Q) query.Q {
		switch q.(type) {
		case *query.Not:
			notCount--
		}
		return q
	})
	if hasNotRef {
		return nil, errors.Errorf("search clauses that filter git refs cannot be negated: %s", q)
	}
	if !hasRef {
		rr := make([]RepositoryRevisions, len(repos))
		for i, r := range repos {
			rr[i].Repo = r
			rr[i].Revs = []RevisionSpecifier{{RevSpec: ""}}
		}
		return rr, nil
	}

	// We need to evaluate the repo atoms as regular expressions. This should
	// happen inside the query package, for now we do it here.
	var reErr error
	re := map[string]*regexp.Regexp{}
	query.VisitAtoms(q, func(q query.Q) {
		if r, ok := q.(*query.Repo); ok {
			m, err := regexp.Compile(r.Pattern)
			if err != nil {
				reErr = err
				return
			}
			re[r.Pattern] = m
		}
	})
	if reErr != nil {
		return nil, reErr
	}

	// Now for each repository we evaluate each Repo* atom. Whatever ref
	// specifiers are left apply to the repository.
	rr := make([]RepositoryRevisions, 0, len(repos))
	for _, r := range repos {
		eval := query.Map(q, nil, func(q query.Q) query.Q {
			switch a := q.(type) {
			case *query.Repo:
				return &query.Const{Value: re[a.Pattern].MatchString(string(r.Name))}
			case *query.RepoSet:
				_, matches := a.Set[string(r.Name)]
				return &query.Const{Value: matches}
			}
			return q
		})
		eval = query.Simplify(eval)

		if c, ok := eval.(*query.Const); ok && !c.Value {
			// We don't expect this to happen, since we expect repos to be
			// populated via RepoQuery. But checking just in case.
			continue
		}

		// Collect refs. Also set them to false to see if we have to
		// additionally evaluate the default branch.
		var revs []RevisionSpecifier
		eval = query.Map(eval, nil, func(q query.Q) query.Q {
			if b, ok := q.(*query.Ref); ok {
				revs = append(revs, parseRev(b.Pattern))
				return &query.Const{Value: false}
			}
			return q
		})
		eval = query.Simplify(eval)

		if c, ok := eval.(*query.Const); !ok || c.Value {
			// We have to also check the default branch.
			revs = append(revs, parseRev(""))
		}

		rr = append(rr, RepositoryRevisions{Repo: r, Revs: revs})
	}
	return rr, nil
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

	// Replace all atoms (except constants and repo) with a constant related
	// to the ancestor not count. We track the not count by incrementing
	// notCount when we visit a Not node, and decrementing it when we leave
	// it.
	var err error
	notCount := 0
	q = query.Map(q, func(q query.Q) query.Q {
		// pre
		switch q.(type) {
		case *query.Not:
			notCount++
		}
		return q
	}, func(q query.Q) query.Q {
		// post
		switch c := q.(type) {
		case *query.Not:
			notCount--

		case *query.Repo:
			// Preserve Repo atom.
			return q
		case *query.Const:
			// Preserve Const atom.
			return q

		case *query.Type:
			// Remove Type from expression, doesn't affect repos searched.
			return c.Child

		case *query.RepoSet:
			err = errors.Errorf("unsupported RepoSet in RepoQuery %s", q.String())
		}

		if query.IsAtom(q) {
			// If this gets constant evaluated all the way to the root, it
			// will be true => doesn't reduce the set of repositories the
			// expression will return.
			return &query.Const{Value: notCount%2 == 0}
		}

		return q
	})
	if err != nil {
		return nil, err
	}

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
