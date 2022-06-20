package zoekt

import (
	zoektquery "github.com/google/zoekt/query"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultGlobalQueryScope returns a Zoekt query that applies a default
// repository scope based on RepoOptions. This default scope is determined
// statically from a query, and does not depend on runtime values.
func DefaultGlobalQueryScope(repoOptions search.RepoOptions) (zoektquery.Q, error) {
	if !(repoOptions.Visibility == query.Public || repoOptions.Visibility == query.Any) {
		// If "Public" or "Any" repos are excluded, the default scope
		// static scope is empty, and implies a different RepoVisibility
		// (e.g., "Private")
		return nil, nil
	}
	rc := zoektquery.RcOnlyPublic
	apply := func(f zoektquery.RawConfig, b bool) {
		if !b {
			return
		}
		rc |= f
	}
	apply(zoektquery.RcOnlyArchived, repoOptions.OnlyArchived)
	apply(zoektquery.RcNoArchived, repoOptions.NoArchived)
	apply(zoektquery.RcOnlyForks, repoOptions.OnlyForks)
	apply(zoektquery.RcNoForks, repoOptions.NoForks)

	children := []zoektquery.Q{&zoektquery.Branch{Pattern: "HEAD", Exact: true}, rc}
	for _, pat := range repoOptions.MinusRepoFilters {
		re, err := regexp.Compile(`(?i)` + pat)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid regex for -repo filter %q", pat)
		}
		children = append(children, &zoektquery.Not{Child: &zoektquery.RepoRegexp{Regexp: re}})
	}
	return zoektquery.NewAnd(children...), nil
}

// GlobalZoektQuery is an object that represents a query to Zoekt across all its
// replicas. It exposes methods to modify the scope of such a query, (e.g., to
// ensure repo privacy filters). A `Generate` method converts the object to a
// Zoekt query that ensures appropriate repo privacy scopes.
type GlobalZoektQuery struct {
	Query          zoektquery.Q
	RepoScope      []zoektquery.Q
	IncludePrivate bool
}

func NewGlobalZoektQuery(query zoektquery.Q, scope zoektquery.Q, includePrivate bool) *GlobalZoektQuery {
	repoScope := []zoektquery.Q{}
	if scope != nil {
		repoScope = append(repoScope, scope)
	}
	return &GlobalZoektQuery{
		Query:          query,
		RepoScope:      repoScope,
		IncludePrivate: includePrivate,
	}
}

// ApplyPrivateFilter ensures that the argument, a set of user private
// repositories, are included in a Global Zoekt search scope. Note that this
// method only adds a set of private repositories to the scope, if the
// construction of a GlobalZoektQuery was permitted to includePrivate
// repositories.
func (q *GlobalZoektQuery) ApplyPrivateFilter(userPrivateRepos []types.MinimalRepo) {
	if q.IncludePrivate && len(userPrivateRepos) > 0 {
		ids := make([]uint32, 0, len(userPrivateRepos))
		for _, r := range userPrivateRepos {
			ids = append(ids, uint32(r.ID))
		}
		q.RepoScope = append(q.RepoScope, zoektquery.NewSingleBranchesRepos("HEAD", ids...))
	}
}

// Generate generates a Global Zoekt query that ensures the appropriate repo
// scope (i.e., whether to either exclusively public, exclusively private, or
// either public or private repositories)
func (q *GlobalZoektQuery) Generate() zoektquery.Q {
	return zoektquery.Simplify(zoektquery.NewAnd(q.Query, zoektquery.NewOr(q.RepoScope...)))
}
