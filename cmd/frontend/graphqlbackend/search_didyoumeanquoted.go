package graphqlbackend

import (
	"context"
	"fmt"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
)

type didYouMeanQuotedResolver struct {
	query string
	err   error
}

func (r *didYouMeanQuotedResolver) Results(context.Context) (*searchResultsResolver, error) {
	sqds := proposedQuotedQueries(r.query)
	srr := &searchResultsResolver{
		alert: &searchAlert{
			title:           "Quoting the query may help if you meant a literal search.",
			proposedQueries: sqds,
		},
	}
	if r.err != nil {
		srr.alert.description = r.err.Error()
	}
	return srr, nil
}

func (r *didYouMeanQuotedResolver) Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error) {
	return nil, nil
}

func (r *didYouMeanQuotedResolver) Stats(context.Context) (*searchResultsStats, error) {
	srs := &searchResultsStats{}
	return srs, nil
}

// proposedQuotedQueries generates various ways of quoting the given query,
// with descriptions, removing duplicates.
func proposedQuotedQueries(rawQuery string) []*searchQueryDescription {
	q := syntax.ParseAllowingErrors(rawQuery)
	// Make a map from various quotings of the query to their descriptions.
	// This should take care of deduplicating them.
	// The descriptions are in a particular order to make the simpler descriptions take precedence.
	qq2d := make(map[string]string)
	qq2d[q.WithErrorsQuoted().String()] = "quote just the errored parts"
	qq2d[fmt.Sprintf("%q", rawQuery)] = "quote the whole thing"
	var sqds []*searchQueryDescription
	for qq, desc := range qq2d {
		if qq == rawQuery {
			continue
		}
		sqds = append(sqds, &searchQueryDescription{
			description: desc,
			query:       qq,
		})
	}
	sort.Slice(sqds, func(i, j int) bool { return sqds[i].description < sqds[j].description })
	return sqds
}
