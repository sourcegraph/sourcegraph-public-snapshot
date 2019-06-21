package graphqlbackend

import (
	"context"
	"fmt"
	rxsyntax "regexp/syntax"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/types"
)

type didYouMeanQuotedResolver struct {
	query string
	err   error
}

func (r *didYouMeanQuotedResolver) Results(context.Context) (*searchResultsResolver, error) {
	sqds := proposedQuotedQueries(r.query)
	switch e := r.err.(type) {
	case *types.TypeError:
		switch e := e.Err.(type) {
		case *rxsyntax.Error:
			srr := &searchResultsResolver{
				alert: &searchAlert{
					title:           makeTitle(e.Error()),
					description:     "Quoting the query may help if you want a literal match instead of a regular expression match.",
					proposedQueries: sqds,
				},
			}
			return srr, nil
		default:
			return nil, r.err
		}
	case *syntax.ParseError:
		srr := &searchResultsResolver{
			alert: &searchAlert{
				title:           makeTitle(e.Msg),
				description:     "Quoting the query may help if you want a literal match.",
				proposedQueries: sqds,
			},
		}
		return srr, nil
	default:
		return nil, r.err
	}
}

func (r *didYouMeanQuotedResolver) Suggestions(context.Context, *searchSuggestionsArgs) ([]*searchSuggestionResolver, error) {
	return nil, r.err
}

func (r *didYouMeanQuotedResolver) Stats(context.Context) (*searchResultsStats, error) {
	return nil, r.err
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

func makeTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
