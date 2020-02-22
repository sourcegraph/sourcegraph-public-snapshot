package graphqlbackend

import (
	"context"
	"fmt"
	"path"
	"regexp"
	rxsyntax "regexp/syntax"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
	querytypes "github.com/sourcegraph/sourcegraph/internal/search/query/types"
)

type searchAlert struct {
	prometheusType  string
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

// alertForQuery converts errors in the query to search alerts.
func alertForQuery(queryString string, err error) *searchAlert {
	switch e := err.(type) {
	case *syntax.ParseError:
		return &searchAlert{
			prometheusType:  "parse_syntax_error",
			title:           capFirst(e.Msg),
			description:     "Quoting the query may help if you want a literal match.",
			proposedQueries: proposedQuotedQueries(queryString),
		}
	case *query.ValidationError:
		return &searchAlert{
			prometheusType: "validation_error",
			title:          "Invalid Query",
			description:    capFirst(e.Msg),
		}
	case *querytypes.TypeError:
		switch e := e.Err.(type) {
		case *rxsyntax.Error:
			return &searchAlert{
				prometheusType:  "typecheck_regex_syntax_error",
				title:           capFirst(e.Error()),
				description:     "Quoting the query may help if you want a literal match instead of a regular expression match.",
				proposedQueries: proposedQuotedQueries(queryString),
			}
		}
	}
	return &searchAlert{
		prometheusType: "generic_invalid_query",
		title:          "Unable To Process Query",
		description:    capFirst(err.Error()),
	}
}

func alertForTimeout(usedTime time.Duration, suggestTime time.Duration, r *searchResolver) *searchAlert {
	return &searchAlert{
		prometheusType: "timed_out",
		title:          "Timed out while searching",
		description:    fmt.Sprintf("We weren't able to find any results in %s.", roundStr(usedTime.String())),
		proposedQueries: []*searchQueryDescription{
			{
				description: "query with longer timeout",
				query:       fmt.Sprintf("timeout:%v %s", suggestTime, omitQueryField(r.query.ParseTree, query.FieldTimeout)),
				patternType: r.patternType,
			},
		},
	}
}

func alertForStalePermissions() *searchAlert {
	return &searchAlert{
		prometheusType: "no_resolved_repos__stale_permissions",
		title:          "Permissions syncing in progress",
		description:    "Permissions are being synced from your code host, please wait for a minute and try again.",
	}
}

func alertForQuotesInQueryInLiteralMode(p syntax.ParseTree) *searchAlert {
	return &searchAlert{
		prometheusType: "no_results__suggest_quotes",
		title:          "No results. Did you mean to use quotes?",
		description:    "Your search is interpreted literally and contains quotes. Did you mean to search for quotes?",
		proposedQueries: []*searchQueryDescription{{
			description: "Remove quotes",
			query:       omitQuotes(p),
			patternType: query.SearchTypeLiteral,
		}},
	}
}

func (r *searchResolver) alertForNoResolvedRepos(ctx context.Context) (*searchAlert, error) {
	repoFilters, minusRepoFilters := r.query.RegexpPatterns(query.FieldRepo)
	repoGroupFilters, _ := r.query.StringValues(query.FieldRepoGroup)
	fork, _ := r.query.StringValue(query.FieldFork)
	onlyForks, noForks := fork == "only", fork == "no"

	// Handle repogroup-only scenarios.
	if len(repoFilters) == 0 && len(repoGroupFilters) == 0 {
		return &searchAlert{
			prometheusType: "no_resolved_repos__no_repositories",
			title:          "Add repositories or connect repository hosts",
			description:    "There are no repositories to search. Add an external service connection to your code host.",
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) == 1 {
		return &searchAlert{
			prometheusType: "no_resolved_repos__repogroup_empty",
			title:          fmt.Sprintf("Add repositories to repogroup:%s to see results", repoGroupFilters[0]),
			description:    fmt.Sprintf("The repository group %q is empty. See the documentation for configuration and troubleshooting.", repoGroupFilters[0]),
		}, nil
	}
	if len(repoFilters) == 0 && len(repoGroupFilters) > 1 {
		return &searchAlert{
			prometheusType: "no_resolved_repos__repogroup_none_in_common",
			title:          "Repository groups have no repositories in common",
			description:    "No repository exists in all of the specified repository groups.",
		}, nil
	}

	// TODO(sqs): handle -repo:foo fields.

	withoutRepoFields := omitQueryField(r.query.ParseTree, query.FieldRepo)

	var a searchAlert
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
				query:       omitQueryField(r.query.ParseTree, query.FieldRepoGroup),
				patternType: r.patternType,
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
				patternType: r.patternType,
			})
		} else {
			// Fall back to removing repo filters.
			a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
				description: "remove repo: filters",
				query:       withoutRepoFields,
				patternType: r.patternType,
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
				query:       omitQueryField(r.query.ParseTree, query.FieldRepoGroup),
				patternType: r.patternType,
			})
		}

		a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
			description: "remove repo: filters",
			query:       withoutRepoFields,
			patternType: r.patternType,
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
				patternType: r.patternType,
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

		if a.title == "" {
			a.title = "No repositories satisfied your repo: filter"
			a.description = "Change your repo: filter to see results"
			if proposeQueries && strings.TrimSpace(withoutRepoFields) != "" {
				a.proposedQueries = append(a.proposedQueries, &searchQueryDescription{
					description: "remove repo: filter",
					query:       withoutRepoFields,
					patternType: r.patternType,
				})
			}
		}
	}

	return &a, nil
}

func (r *searchResolver) alertForOverRepoLimit(ctx context.Context) (*searchAlert, error) {
	alert := &searchAlert{
		prometheusType: "over_repo_limit",
		title:          "Too many matching repositories",
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
			query:       newExpr.String(),
			patternType: r.patternType,
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
				query:       newExpr.String(),
				patternType: r.patternType,
			})
		}
	}

	return alert, nil
}

// alertForStructuralSearch filters certain errors from multiErr and converts
// them to an alert. We surface one alert at a time, so for multiple errors only
// the last converted error will be surfaced in the alert.
func alertForStructuralSearch(multiErr *multierror.Error) (newMultiErr *multierror.Error, alert *searchAlert) {
	if multiErr != nil {
		for _, err := range multiErr.Errors {
			if strings.Contains(err.Error(), "Worker_oomed") || strings.Contains(err.Error(), "Worker_exited_abnormally") {
				alert = &searchAlert{
					prometheusType: "structural_search_needs_more_memory",
					title:          "Structural search needs more memory",
					description:    "Running your structural search may require more memory. If you are running the query on many repositories, try reducing the number of repositories with the `repo:` filter.",
				}
			} else if strings.Contains(err.Error(), "no indexed repositories for structural search") {
				var msg string
				if envvar.SourcegraphDotComMode() {
					msg = "The good news is you can index any repository you like in a self-install. It takes less than 5 minutes to set up: https://docs.sourcegraph.com/#quickstart"
				} else {
					msg = "Learn more about managing indexed repositories in our documentation: https://docs.sourcegraph.com/admin/search#indexed-search."
				}
				alert = &searchAlert{
					prometheusType: "structural_search_on_zero_indexed_repos",
					title:          "Unindexed repositories with structural search",
					description:    fmt.Sprintf("Structural search currently only works on indexed repositories. Some of the repositories to search are not indexed, so we can't return results for them. %s", msg),
				}
			} else {
				newMultiErr = multierror.Append(newMultiErr, err)
			}
		}
	}
	return newMultiErr, alert
}

func alertForMissingRepoRevs(patternType query.SearchType, missingRepoRevs []*search.RepositoryRevisions) *searchAlert {
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
		prometheusType: "missing_repo_revs",
		title:          "Some repositories could not be searched",
		description:    description,
	}
}

func omitQueryField(p syntax.ParseTree, field string) string {
	omitField := func(e syntax.Expr) *syntax.Expr {
		if e.Field == field {
			return nil
		}
		return &e
	}
	return syntax.Map(p, omitField).String()
}

func omitQuotes(p syntax.ParseTree) string {
	omitQuotes := func(e syntax.Expr) *syntax.Expr {

		if e.Field == "" && strings.HasPrefix(e.Value, `"\"`) && strings.HasSuffix(e.Value, `\""`) {
			e.Value = strings.TrimSuffix(strings.TrimPrefix(e.Value, `"\"`), `\""`)
			return &e
		}
		return &e
	}
	return syntax.Map(p, omitQuotes).String()
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

func (a searchAlert) Results(context.Context) (*SearchResultsResolver, error) {
	alert := &searchAlert{
		prometheusType:  a.prometheusType,
		title:           a.title,
		description:     a.description,
		proposedQueries: a.proposedQueries,
	}
	return &SearchResultsResolver{alert: alert}, nil
}

func (searchAlert) Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error) {
	return nil, nil
}
func (searchAlert) Stats(context.Context) (*searchResultsStats, error) { return nil, nil }
