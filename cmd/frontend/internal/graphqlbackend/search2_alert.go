package graphqlbackend

import (
	"context"
	"fmt"

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

		repos1, _, _, err := resolveRepositories(ctx, repoFilters.Values(), minusRepoFilters.Values(), nil)
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
		repos2, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters.Values(), repoGroupFilters.Values())
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchFieldRepo)
			query.query += " " + search2.FormatToken(searchFieldRepo, unionRepoFilter)
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

		repos1, _, _, err := resolveRepositories(ctx, repoFilters.Values(), minusRepoFilters.Values(), nil)
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
		repos2, _, _, err := resolveRepositories(ctx, []string{unionRepoFilter}, minusRepoFilters.Values(), repoGroupFilters.Values())
		if err != nil {
			return nil, err
		}
		if len(repos2) > 0 {
			query := omitQueryFields(r, searchFieldRepo)
			query.query += " " + search2.FormatToken(searchFieldRepo, unionRepoFilter)
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
	// TODO(sqs): add repo filters based on the result set so far
	return &searchAlert{
		title:       "Too many matching repositories",
		description: "Narrow your search with repo: filters to see results.",
	}, nil
}

func omitQueryFields(r *searchResolver2, field search2.Field) searchQuery {
	return searchQuery{
		query:      omitQueryTokensWithField(r.query.tokens, field),
		scopeQuery: omitQueryTokensWithField(r.scopeQuery.tokens, field),
	}
}

func omitQueryTokensWithField(tokens search2.Tokens, field search2.Field) string {
	tokens2 := make(search2.Tokens, 0, len(tokens))
	for _, t := range tokens {
		if t.Field == field {
			continue
		}
		tokens2 = append(tokens2, t)
	}
	return tokens2.String()
}
