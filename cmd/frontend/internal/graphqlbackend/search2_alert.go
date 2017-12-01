package graphqlbackend

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search2"
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

func (r *searchResolver2) alertForNoResolvedRepos(ctx context.Context) (*searchAlert, error) {
	repoFilters := r.combinedQuery.fieldValues[searchFieldRepo]
	minusRepoFilters := r.combinedQuery.fieldValues[minusField(searchFieldRepo)]
	repoGroupFilters := r.combinedQuery.fieldValues[searchFieldRepoGroup]

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
			description: fmt.Sprintf("The repository group %q is empty. See the documentation for configuration and troubleshooting.", repoGroupFilters[0].Value),
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

		repos1, _, _, _, err := resolveRepositories(ctx, repoFilters.Values(), minusRepoFilters.Values(), nil)
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, searchFieldRepoGroup),
			})
		}

		unionRepoFilter := unionRegExps(repoFilters.Values())
		repos2, _, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters.Values(), repoGroupFilters.Values())
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchFieldRepo)
			query.query += " " + search2.NewToken(searchFieldRepo, unionRepoFilter).String()
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		} else {
			// Fall back to removing repo filters.
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: "remove repo: filters",
				query:       omitQueryFields(r, searchFieldRepo),
			})
		}

	case len(repoGroupFilters) == 1 && len(repoFilters) == 1:
		a.title = "Expand your repository filters to see results"
		a.description = fmt.Sprintf("No repositories in repogroup:%s satisfied your repo: filter.", repoGroupFilters[0])

		repos1, _, _, _, err := resolveRepositories(ctx, repoFilters.Values(), minusRepoFilters.Values(), nil)
		if err != nil {
			return nil, err
		}
		if len(repos1) > 0 {
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories outside of repogroup:%s", repoGroupFilters[0]),
				query:       omitQueryFields(r, searchFieldRepoGroup),
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       omitQueryFields(r, searchFieldRepo),
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) > 1:
		a.title = "Expand your repo: filters to see results"
		a.description = fmt.Sprintf("No repositories satisfied all of your repo: filters.")

		unionRepoFilter := unionRegExps(repoFilters.Values())
		repos2, _, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters.Values(), repoGroupFilters.Values())
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchFieldRepo)
			query.query += " " + search2.NewToken(searchFieldRepo, unionRepoFilter).String()
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: fmt.Sprintf("include repositories satisfying any (not all) of your repo: filters"),
				query:       query,
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       omitQueryFields(r, searchFieldRepo),
		})

	case len(repoGroupFilters) == 0 && len(repoFilters) == 1:
		a.title = "Change your repo: filter to see results"
		a.description = fmt.Sprintf("No repositories satisfied your repo: filter.")

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filter",
			query:       omitQueryFields(r, searchFieldRepo),
		})
	}

	return &a, nil
}

func (r *searchResolver2) alertForOverRepoLimit(ctx context.Context) (*searchAlert, error) {
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
	// (Above we assume MAX_REPOS_TO_SEARCH is 30.)
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
		repoFieldValues := r.combinedQuery.fieldValues[searchFieldRepo].Values()

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
		query := addQueryRegexpField(r.query.tokens, searchFieldRepo, repoParentPattern)
		alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
			description: "in repositories under " + repoParent + more,
			query: searchQuery{
				query:      query.String(),
				scopeQuery: r.scopeQuery.tokens.String(),
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
			query := addQueryRegexpField(r.query.tokens, searchFieldRepo, "^"+regexp.QuoteMeta(pathToPropose)+"$")
			alert.proposedQueries = append(alert.proposedQueries, &searchQueryDescription{
				description: "in the repository " + strings.TrimPrefix(pathToPropose, "github.com/"),
				query: searchQuery{
					query:      query.String(),
					scopeQuery: r.scopeQuery.tokens.String(),
				},
			})
		}
	}

	return alert, nil
}

func (r *searchResolver2) alertForMissingRepoRevs(missingRepoRevs []*repositoryRevision) *searchAlert {
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

func omitQueryFields(r *searchResolver2, field search2.Field) searchQuery {
	return searchQuery{
		query:      omitQueryTokensWithField(r.query.tokens, field).String(),
		scopeQuery: omitQueryTokensWithField(r.scopeQuery.tokens, field).String(),
	}
}

func omitQueryTokensWithField(tokens search2.Tokens, field search2.Field) search2.Tokens {
	tokens2 := make(search2.Tokens, 0, len(tokens))
	for _, t := range tokens {
		if t.Field == field {
			continue
		}
		tokens2 = append(tokens2, t)
	}
	return tokens2
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

// addQueryRegexpField adds a new token to the query with the given field
// and pattern value. The token is assumed to be a regexp.
//
// It tries to simplify (avoid redundancy in) the result. For example, given
// a query like "x:foo", if given a field "x" with pattern "foobar" to add,
// it will return a query "x:foobar" instead of "x:foo x:foobar". It is not
// guaranteed to always return the simplest query.
func addQueryRegexpField(queryTokens search2.Tokens, field search2.Field, pattern string) search2.Tokens {
	queryTokens = queryTokens.Copy()
	var added bool
	for i, token := range queryTokens {
		if token.Field == field && strings.Contains(pattern, token.Value.Value) {
			queryTokens[i] = search2.NewToken(field, pattern)
			added = true
			break
		}
	}

	if !added {
		queryTokens = append(queryTokens, search2.NewToken(field, pattern))
	}
	return queryTokens
}
