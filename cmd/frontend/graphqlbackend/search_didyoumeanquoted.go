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
	qq2d[q.WithNonFieldPartsQuoted().String()] = "query with parts quoted, except for fields"
	qq2d[q.WithNonFieldsQuoted().String()] = "query quoted, except for fields"
	qq2d[q.WithPartsQuoted().String()] = "query with parts quoted"
	qq2d[fmt.Sprintf("%q", rawQuery)] = "query quoted entirely"
	var sqds []*searchQueryDescription
	for qq, desc := range qq2d {
		sqds = append(sqds, &searchQueryDescription{
			description: desc,
			query:       qq,
		})
	}
	sort.Slice(sqds, func(i, j int) bool { return sqds[i].description < sqds[j].description })
	return sqds
}
