package graphqlbackend

import (
	"context"
	"fmt"
	rxsyntax "regexp/syntax"
	"sort"
	"strings"
	"unicode"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/types"
)

type didYouMeanQuotedResolver struct {
	query string
	err   error
}

func (r *didYouMeanQuotedResolver) Results(context.Context) (*SearchResultsResolver, error) {
	sqds := proposedQuotedQueries(r.query)
	switch e := r.err.(type) {
	case *types.TypeError:
		switch e := e.Err.(type) {
		case *rxsyntax.Error:
			srr := &SearchResultsResolver{
				alert: &searchAlert{
					title:           capFirst(e.Error()),
					description:     "Quoting the query may help if you want a literal match instead of a regular expression match.",
					proposedQueries: sqds,
				},
			}
			return srr, nil
		default:
			return nil, r.err
		}
	case *syntax.ParseError:
		srr := &SearchResultsResolver{
			alert: &searchAlert{
				title:           capFirst(e.Msg),
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
const partsMsg = "treat the errored parts as literals"
const wholeMsg = "treat the whole query as a literal"

func proposedQuotedQueries(rawQuery string) []*searchQueryDescription {
	q := syntax.ParseAllowingErrors(rawQuery)
	// Make a map from various quotings of the query to their descriptions.
	// This should take care of deduplicating them.
	// The descriptions are in a particular order to make the simpler descriptions take precedence.
	qq2d := make(map[string]string)
	qq2d[q.WithErrorsQuoted().String()] = partsMsg
	qq2d[fmt.Sprintf("%q", rawQuery)] = wholeMsg
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

// capFirst capitalizes the first rune in the given string. It can be safely
// used with UTF-8 strings.
func capFirst(s string) string {
	i := 0
	return strings.Map(func(r rune) rune {
		i++
		if i == 1 {
			return unicode.ToTitle(r)
		}
		return r
	}, s)
}
