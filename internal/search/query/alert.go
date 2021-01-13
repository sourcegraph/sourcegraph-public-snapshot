package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

func OmitQueryField(p syntax.ParseTree, field string) string {
	omitField := func(e syntax.Expr) *syntax.Expr {
		if e.Field == field {
			return nil
		}
		return &e
	}
	return syntax.Map(p, omitField).String()
}

// addRegexpField adds a new expr to the query with the given field
// and pattern value. The field is assumed to be a regexp.
//
// It tries to remove redundancy in the result. For example, given
// a query like "x:foo", if given a field "x" with pattern "foobar" to add,
// it will return a query "x:foobar" instead of "x:foo x:foobar". It is not
// guaranteed to always return the simplest query.
func AddRegexpField(p syntax.ParseTree, field, pattern string) string {
	var added bool
	addRegexpField := func(e syntax.Expr) *syntax.Expr {
		if e.Field == field && strings.Contains(pattern, e.Value) {
			e.Value = pattern
			added = true
			return &e
		}
		return &e
	}
	modified := syntax.Map(p, addRegexpField)
	if !added {
		p = append(p, &syntax.Expr{
			Field:     field,
			Value:     pattern,
			ValueType: syntax.TokenLiteral,
		})
		return p.String()
	}
	return modified.String()
}

type ProposedQuery struct {
	Description string
	Query       string
}

// ProposedQuotedQueries generates various ways of quoting the given query,
// with descriptions, removing duplicates.
const partsMsg = "treat the errored parts as literals"
const wholeMsg = "treat the whole query as a literal"

func ProposedQuotedQueries(rawQuery string) []*ProposedQuery {
	q := syntax.ParseAllowingErrors(rawQuery)
	// Make a map from various quotings of the query to their descriptions.
	// This should take care of deduplicating them.
	// The descriptions are in a particular order to make the simpler descriptions take precedence.
	qq2d := make(map[string]string)
	qq2d[q.WithErrorsQuoted().String()] = partsMsg
	qq2d[fmt.Sprintf("%q", rawQuery)] = wholeMsg
	var sqds []*ProposedQuery
	for qq, desc := range qq2d {
		if qq == rawQuery {
			continue
		}
		sqds = append(sqds, &ProposedQuery{
			Description: desc,
			Query:       qq,
		})
	}
	sort.Slice(sqds, func(i, j int) bool { return sqds[i].Description < sqds[j].Description })
	return sqds
}
