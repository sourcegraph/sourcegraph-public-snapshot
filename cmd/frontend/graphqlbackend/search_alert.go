package graphqlbackend

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
)

type searchAlert struct {
	title           string
	description     string
	patternType     SearchType
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
	for _, proposedQuery := range a.proposedQueries {
		if proposedQuery.description != "Remove quotes" {
			switch a.patternType {
			case SearchTypeRegex:
				proposedQuery.query = proposedQuery.query + " patternType:regexp"
			case SearchTypeLiteral:
				proposedQuery.query = proposedQuery.query + " patternType:literal"
			case SearchTypeStructural:
				// Don't append patternType:structural, it is not erased from the query like
				// patterntype:regexp and patterntype:literal.
				// TODO(RVT): Making this consistent requires a change on the UI side.
			default:
				panic("unreachable")
			}
		}
	}
	return &a.proposedQueries
}

func (r *searchResolver) alertForQuotesInQueryInLiteralMode(ctx context.Context) (*searchAlert, error) {
	return &searchAlert{
		title:       "No results. Did you mean to use quotes?",
		description: "Your search is interpreted literally and contains quotes. Did you mean to search for quotes?",
		proposedQueries: []*searchQueryDescription{{
			description: "Remove quotes",
			query:       syntax.ExprString(omitQuotes(r.query)),
		}},
	}, nil
}

func (r *searchResolver) alertForNoResolvedRepos(ctx context.Context) (*searchAlert, error) {
	repoFilters, minusRepoFilters := r.query.RegexpPatterns(query.FieldRepo)
	repoGroupFilters, _ := r.query.StringValues(query.FieldRepoGroup)
	fork, _ := r.query.StringValue(query.FieldFork)
	onlyForks, noForks := fork == "only", fork == "no"

	// Handle repogroup-only scenarios.
	if len(repoFilters) == 0 && len(repoGroupFilters) == 0 {
		return &searchAlert{
			title:       "Add repositories or connect repository hosts",
			description: "There are no repositories to search. Add an external service connection to your code host.",
			patternType: r.patternType,
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) == 1 {
		return &searchAlert{
			title:       fmt.Sprintf("Add repositories to repogroup:%s to see results", repoGroupFilters[0]),
			description: fmt.Sprintf("The repository group %q is empty. See the documentation for configuration and troubleshooting.", repoGroupFilters[0]),
			patternType: r.patternType,
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) > 1 {
		return &searchAlert{
			title:       "Repository groups have no repositories in common",
			description: "No repository exists in all of the specified repository groups.",
			patternType: r.patternType,
		}, nil
	}

	// TODO(sqs): handle -repo:foo fields.

	withoutRepoFields := omitQueryFields(r, query.FieldRepo)

	var a searchAlert
	a.patternType = r.patternType
	switch {
	case len(repoGroupFilters) > 1:
		// This is a rare case, so don't bother proposing queries.
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repository exists in all specified groups and satisfies all of your repo: filters.")

	case len(repoGroupFilters) == 1 && len(repoFilters) > 1:
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repositories in repogroup:%s satisfied all of your repo: filters.", repoGroupFilters[0])

		repos1, _, _, err := resolveRepositories(ctx, resolveRepoOp{repoFilters: repoFilters, minusRepoFilters: minusRepoFilters, onlyForks: onlyForks, noForks: noForks})
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, query.FieldRepoGroup),
			})
		}

		unionRepoFilter := unionRegExps(repoFilters)
		repos2, _, _, err := resolveRepositories(ctx, resolveRepoOp{repoFilters: []string{unionRepoFilter}, minusRepoFilters: minusRepoFilters, repoGroupFilters: repoGroupFilters, onlyForks: onlyForks, noForks: noForks})
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := withoutRepoFields
			query += fmt.Sprintf(" repo:%s", unionRepoFilter)
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		} else {
			// Fall back to removing repo filters.
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: "remove repo: filters",
				query:       withoutRepoFields,
			})
		}

	case len(repoGroupFilters) == 1 && len(repoFilters) == 1:
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repositories in repogroup:%s satisfied your repo: filter.", repoGroupFilters[0])

		repos1, _, _, err := resolveRepositories(ctx, resolveRepoOp{repoFilters: repoFilters, minusRepoFilters: minusRepoFilters, noForks: noForks, onlyForks: onlyForks})
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, query.FieldRepoGroup),
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       withoutRepoFields,
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) > 1:
		a.title = "Expand your repo: filters to see results"
		a.description = fmt.Sprintf("No repositories satisfied all of your repo: filters.")

		unionRepoFilter := unionRegExps(repoFilters)
		repos2, _, _, err := resolveRepositories(ctx, resolveRepoOp{repoFilters: []string{unionRepoFilter}, minusRepoFilters: minusRepoFilters, repoGroupFilters: repoGroupFilters, noForks: noForks, onlyForks: onlyForks})
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := withoutRepoFields
			query += fmt.Sprintf(" repo:%s", unionRepoFilter)
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       withoutRepoFields,
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) == 1:
		isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx) == nil
		proposeQueries := true
		if !envvar.SourcegraphDotComMode() {
			if needsRepoConfig, err := needsRepositoryConfiguration(ctx); err == nil && needsRepoConfig {
				proposeQueries = false
				a.title = "No repositories or code hosts configured"
				a.description = "To start searching code, "
				if isSiteAdmin {
					a.description += "first go to site admin to configure repositories and code hosts."
				} else {
					a.description = "ask the site admin to configure and enable repositories."
				}
			}
		}

		suggestEnablingRepos := false
		if a.title == "" && !envvar.SourcegraphDotComMode() {
			repoPattern, _ := search.ParseRepositoryRevisions(repoFilters[0])
			repos, err := db.Repos.List(ctx, db.ReposListOptions{
				Enabled:         false,
				Disabled:        true,
				IncludePatterns: []string{optimizeRepoPatternWithHeuristics(string(repoPattern))},
				LimitOffset:     &db.LimitOffset{Limit: 1},
			})
			if err == nil && len(repos) > 0 && isSiteAdmin {
				suggestEnablingRepos = true
			}
		}

		if a.title == "" {
			if suggestEnablingRepos {
				a.title = "Your repo: filter matched only disabled repositories"
				a.description = "Go to site admin to enable more repositories, or broaden your repo: scope."
			} else {
				a.title = "No repositories satisfied your repo: filter"
				a.description = "Change your repo: filter to see results"
			}
			if proposeQueries && strings.TrimSpace(withoutRepoFields) != "" {
				a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
					description: "remove repo: filter",
					query:       withoutRepoFields,
				})
			}
		}
	}

	return &a, nil
}

func (r *searchResolver) alertForOverRepoLimit(ctx context.Context) (*searchAlert, error) {
	alert := &searchAlert{
		title:       "Too many matching repositories",
		patternType: r.patternType,
	}

	if envvar.SourcegraphDotComMode() {
		alert.description = "Use a 'repo:' or 'repogroup:' filter to narrow your search and see results or set up a self-hosted Sourcegraph instance to search an unlimited number of repositories."
	} else {
		alert.description = "Use a 'repo:' or 'repogroup:' filter to narrow your search and see results."
	}

	isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx) == nil
	if isSiteAdmin {
		alert.description += " As a site admin, you can increase the limit by changing maxReposToSearch in site config."
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
	repos, _, _, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(repos))
	pathPatterns := make([]string, len(repos))
	for i, repo := range repos {
		paths[i] = string(repo.Repo.Name)
		pathPatterns[i] = "^" + regexp.QuoteMeta(string(repo.Repo.Name)) + "$"
	}

	// See if we can narrow it down by using filters like
	// repo:github.com/myorg/.
	const maxParentsToPropose = 4
	ctx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
outer:
	for i, repoParent := range pathParentsByFrequency(paths) {
		if i >= maxParentsToPropose || ctx.Err() == nil {
			break
		}
		repoParentPattern := "^" + regexp.QuoteMeta(repoParent) + "/"
		repoFieldValues, _ := r.query.RegexpPatterns(query.FieldRepo)

		for _, v := range repoFieldValues {
			if strings.HasPrefix(v, strings.TrimSuffix(repoParentPattern, "/")) {
				continue outer // this repo: filter is already applied
			}
		}

		repoFieldValues = append(repoFieldValues, repoParentPattern)
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		_, _, overLimit, err := r.resolveRepositories(ctx, repoFieldValues)
		if ctx.Err() != nil {
			continue
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
		newExpr := addQueryRegexpField(r.query, query.FieldRepo, repoParentPattern)
		alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
			description: "in repositories under " + repoParent + more,
			query:       syntax.ExprString(newExpr),
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
			newExpr := addQueryRegexpField(r.query, query.FieldRepo, "^"+regexp.QuoteMeta(pathToPropose)+"$")
			alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
				description: "in the repository " + strings.TrimPrefix(pathToPropose, "github.com/"),
				query:       syntax.ExprString(newExpr),
			})
		}
	}

	return alert, nil
}

func (r *searchResolver) alertForMissingRepoRevs(missingRepoRevs []*search.RepositoryRevisions) *searchAlert {
	var description string
	if len(missingRepoRevs) == 1 {
		if len(missingRepoRevs[0].RevSpecs()) == 1 {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it does not contain the revision %q.", missingRepoRevs[0].Repo.Name, missingRepoRevs[0].RevSpecs()[0])
		} else {
			description = fmt.Sprintf("The repository %s matched by your repo: filter could not be searched because it has multiple specified revisions: @%s.", missingRepoRevs[0].Repo.Name, strings.Join(missingRepoRevs[0].RevSpecs(), ","))
		}
	} else {
		repoRevs := make([]string, 0, len(missingRepoRevs))
		for _, r := range missingRepoRevs {
			repoRevs = append(repoRevs, string(r.Repo.Name)+"@"+strings.Join(r.RevSpecs(), ","))
		}
		description = fmt.Sprintf("%d repositories matched by your repo: filter could not be searched because the following revisions do not exist, or differ but were specified for the same repository: %s.", len(missingRepoRevs), strings.Join(repoRevs, ", "))
	}
	return &searchAlert{
		title:       "Some repositories could not be searched",
		description: description,
		patternType: r.patternType,
	}
}

func omitQueryFields(r *searchResolver, field string) string {
	return syntax.ExprString(omitQueryExprWithField(r.query, field))
}

func omitQueryExprWithField(query *query.Query, field string) syntax.ParseTree {
	expr2 := make(syntax.ParseTree, 0, len(query.ParseTree))
	for _, e := range query.ParseTree {
		if e.Field == field {
			continue
		}
		expr2 = append(expr2, e)
	}
	return expr2
}

func omitQuotes(query *query.Query) syntax.ParseTree {
	result := make(syntax.ParseTree, 0, len(query.ParseTree))
	for _, e := range query.ParseTree {
		cpy := *e
		e = &cpy
		if e.Field == "" && strings.HasPrefix(e.Value, `"\"`) && strings.HasSuffix(e.Value, `\""`) {
			e.Value = strings.TrimSuffix(strings.TrimPrefix(e.Value, `"\"`), `\""`)
		}
		result = append(result, e)
	}
	return result
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
func addQueryRegexpField(query *query.Query, field, pattern string) syntax.ParseTree {
	// Copy query expressions.
	expr := make(syntax.ParseTree, len(query.ParseTree))
	for i, e := range query.ParseTree {
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
