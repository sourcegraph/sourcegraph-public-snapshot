package graphqlbackend

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/syntax"
)

type searchAlert struct {
	title           string
	description     string
	proposedQueries []*searchQueryDescription
}

func (a searchAlert) Title() string { return a.title }

func (a searchAlert) Description() *string {
	if a.description == "" {
		return nil
	}
	return &a.description
}

func (a searchAlert) ProposedQueries() *[]*searchQueryDescription {
	if len(a.proposedQueries) == 0 {
		return nil
	}
	return &a.proposedQueries
}

func (r *searchResolver) alertForNoResolvedRepos(ctx context.Context) (*searchAlert, error) {
	repoFilters, minusRepoFilters := r.combinedQuery.RegexpPatterns(searchquery.FieldRepo)
	repoGroupFilters, _ := r.combinedQuery.StringValues(searchquery.FieldRepoGroup)

	// Handle repogroup-only scenarios.
	if len(repoFilters) == 0 && len(repoGroupFilters) == 0 {
		return &searchAlert{
			title:       "Add repositories or connect repository hosts",
			description: "There are no repositories to search. See the documentation for setup instructions.",
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) == 1 {
		return &searchAlert{
			title:       fmt.Sprintf("Add repositories to repogroup:%s to see results", repoGroupFilters[0]),
			description: fmt.Sprintf("The repository group %q is empty. See the documentation for configuration and troubleshooting.", repoGroupFilters[0]),
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) > 1 {
		return &searchAlert{
			title:       fmt.Sprintf("Repository groups have no repositories in common"),
			description: fmt.Sprintf("No repository exists in all of the specified repository groups."),
		}, nil
	}

	// TODO(sqs): handle -repo:foo fields.

	var a searchAlert
	switch {
	case len(repoGroupFilters) > 1:
		// This is a rare case, so don't bother proposing queries.
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repository exists in all specified groups and satisfies all of your repo: filters.")

	case len(repoGroupFilters) == 1 && len(repoFilters) > 1:
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repositories in repogroup:%s satisfied all of your repo: filters.", repoGroupFilters[0])

		repos1, _, _, _, err := resolveRepositories(ctx, repoFilters, minusRepoFilters, nil)
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, searchquery.FieldRepoGroup),
			})
		}

		unionRepoFilter := unionRegExps(repoFilters)
		repos2, _, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters, repoGroupFilters)
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchquery.FieldRepo)
			query.query += fmt.Sprintf(" repo:%s", unionRepoFilter)
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		} else {
			// Fall back to removing repo filters.
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: "remove repo: filters",
				query:       omitQueryFields(r, searchquery.FieldRepo),
			})
		}

	case len(repoGroupFilters) == 1 && len(repoFilters) == 1:
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repositories in repogroup:%s satisfied your repo: filter.", repoGroupFilters[0])

		repos1, _, _, _, err := resolveRepositories(ctx, repoFilters, minusRepoFilters, nil)
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, searchquery.FieldRepoGroup),
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       omitQueryFields(r, searchquery.FieldRepo),
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) > 1:
		a.title = "Expand your repo: filters to see results"
		a.description = fmt.Sprintf("No repositories satisfied all of your repo: filters.")

		unionRepoFilter := unionRegExps(repoFilters)
		repos2, _, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters, repoGroupFilters)
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchquery.FieldRepo)
			query.query += fmt.Sprintf(" repo:%s", unionRepoFilter)
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       omitQueryFields(r, searchquery.FieldRepo),
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) == 1:
		a.title = "Change your repo: filter to see results"
		a.description = fmt.Sprintf("No repositories satisfied your repo: filter.")

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filter",
			query:       omitQueryFields(r, searchquery.FieldRepo),
		})
	}

	return &a, nil
}

func (r *searchResolver) alertForOverRepoLimit(ctx context.Context) (*searchAlert, error) {
	alert := &searchAlert{
		title:       "Too many matching repositories",
		description: "Narrow your search with a repo: filter to see results.",
	}

	// Try to suggest the most helpful repo: filters to narrow the query.
	//
	// For example, suppose the query contains "repo:kubern" and it matches > 30
	// repositories, and each one of the (clipped result set of) 30 repos has
	// "kubernetes" in their path. Then it's likely that the user would want to
	// search for "repo:kubernetes". If that still matches > 30 repositories,
	// then try to narrow it further using "/kubernetes/", etc.
	//
	// (In the above sample paragraph, we assume MAX_REPOS_TO_SEARCH is 30.)
	//
	// TODO(sqs): this logic can be significantly improved, but it's better than
	// nothing for now.
	repos, _, _, _, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(repos))
	pathPatterns := make([]string, len(repos))
	for i, repo := range repos {
		paths[i] = repo.repo
		pathPatterns[i] = "^" + regexp.QuoteMeta(repo.repo) + "$"
	}

	// See if we can narrow it down by using filters like
	// repo:github.com/myorg/.
	const maxParentsToPropose = 4
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
outer:
	for i, repoParent := range pathParentsByFrequency(paths) {
		if i >= maxParentsToPropose {
			break
		}
		repoParentPattern := "^" + regexp.QuoteMeta(repoParent) + "/"
		repoFieldValues, _ := r.combinedQuery.RegexpPatterns(searchquery.FieldRepo)

		for _, v := range repoFieldValues {
			if strings.HasPrefix(v, strings.TrimSuffix(repoParentPattern, "/")) {
				continue outer // this repo: filter is already applied
			}
		}

		repoFieldValues = append(repoFieldValues, repoParentPattern)
		_, _, _, overLimit, err := r.resolveRepositories(ctx, repoFieldValues)
		if err == context.DeadlineExceeded {
			break
		} else if err != nil {
			return nil, err
		}

		var more string
		if overLimit {
			more = " (further filtering required)"
		}

		// We found a more specific repo: filter that may be narrow enough. Now
		// add it to the user's query, but be smart. For example, if the user's
		// query was "repo:foo" and the parent is "foobar/", then propose "repo:foobar/"
		// not "repo:foo repo:foobar/" (which are equivalent, but shorter is better).
		newExpr := addQueryRegexpField(&r.query, searchquery.FieldRepo, repoParentPattern)
		alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
			description: "in repositories under " + repoParent + more,
			query: searchQuery{
				query:      syntax.ExprString(newExpr),
				scopeQuery: r.scopeQuery.Syntax.Input,
			},
		})
	}
	if len(alert.proposedQueries) == 0 || ctx.Err() == context.DeadlineExceeded {
		// Propose specific repos' paths if we aren't able to propose
		// anything else.
		const maxReposToPropose = 4
		shortest := append([]string{}, paths...) // prefer shorter repo names
		sort.Slice(shortest, func(i, j int) bool {
			return len(shortest[i]) < len(shortest[j]) || (len(shortest[i]) == len(shortest[j]) && shortest[i] < shortest[j])
		})
		for i, pathToPropose := range shortest {
			if i >= maxReposToPropose {
				break
			}
			newExpr := addQueryRegexpField(&r.query, searchquery.FieldRepo, "^"+regexp.QuoteMeta(pathToPropose)+"$")
			alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
				description: "in the repository " + strings.TrimPrefix(pathToPropose, "github.com/"),
				query: searchQuery{
					query:      syntax.ExprString(newExpr),
					scopeQuery: r.scopeQuery.Syntax.Input,
				},
			})
		}
	}

	return alert, nil
}

func (r *searchResolver) alertForMissingRepoRevs(missingRepoRevs []*repositoryRevisions) *searchAlert {
	var description string
	if len(missingRepoRevs) == 1 {
		description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it does not contain the revision %q.", missingRepoRevs[0].repo, missingRepoRevs[0].revSpecsOrDefaultBranch()[0])
	} else {
		revs := make([]string, 0, len(missingRepoRevs))
		for _, r := range missingRepoRevs {
			revs = append(revs, r.revSpecsOrDefaultBranch()...)
		}
		description = fmt.Sprintf("%d repositories matched by your repo: filter could not be searched because they do not contain the specified revisions: %s.", len(missingRepoRevs), strings.Join(revs, ", "))
	}
	return &searchAlert{
		title:       "Some repositories could not be searched",
		description: description,
	}
}

func omitQueryFields(r *searchResolver, field string) searchQuery {
	return searchQuery{
		query:      syntax.ExprString(omitQueryExprWithField(&r.query, field)),
		scopeQuery: syntax.ExprString(omitQueryExprWithField(&r.scopeQuery, field)),
	}
}

func omitQueryExprWithField(query *searchquery.Query, field string) []*syntax.Expr {
	expr2 := make([]*syntax.Expr, 0, len(query.Syntax.Expr))
	for _, e := range query.Syntax.Expr {
		if e.Field == field {
			continue
		}
		expr2 = append(expr2, e)
	}
	return expr2
}

// pathParentsByFrequency returns the most common path parents of the given paths.
// For example, given paths [a/b a/c x/y], it would return [a x] because "a"
// is a parent to 2 paths and "x" is a parent to 1 path.
func pathParentsByFrequency(paths []string) []string {
	var parents []string
	parentFreq := map[string]int{}
	for _, p := range paths {
		parent := path.Dir(p)
		if _, seen := parentFreq[parent]; !seen {
			parents = append(parents, parent)
		}
		parentFreq[parent]++
	}

	sort.Slice(parents, func(i, j int) bool {
		pi, pj := parents[i], parents[j]
		fi, fj := parentFreq[pi], parentFreq[pj]
		return fi > fj || (fi == fj && pi < pj) // freq desc, alpha asc
	})
	return parents
}

// addQueryRegexpField adds a new expr to the query with the given field
// and pattern value. The field is assumed to be a regexp.
//
// It tries to simplify (avoid redundancy in) the result. For example, given
// a query like "x:foo", if given a field "x" with pattern "foobar" to add,
// it will return a query "x:foobar" instead of "x:foo x:foobar". It is not
// guaranteed to always return the simplest query.
func addQueryRegexpField(query *searchquery.Query, field, pattern string) []*syntax.Expr {
	// Copy query expressions.
	expr := make([]*syntax.Expr, len(query.Syntax.Expr))
	for i, e := range query.Syntax.Expr {
		tmp := *e
		expr[i] = &tmp
	}

	var added bool
	for i, e := range expr {
		if e.Field == field && strings.Contains(pattern, e.Value) {
			expr[i].Value = pattern
			added = true
			break
		}
	}

	if !added {
		expr = append(expr, &syntax.Expr{
			Field:     field,
			Value:     pattern,
			ValueType: syntax.TokenLiteral,
		})
	}
	return expr
}
