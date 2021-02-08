package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

// addRegexpField adds a new expr to the query with the given field and pattern
// value. The nonnegated field is assumed to associate with a regexp value. The
// pattern value is assumed to be unquoted.
//
// It tries to remove redundancy in the result. For example, given
// a query like "x:foo", if given a field "x" with pattern "foobar" to add,
// it will return a query "x:foobar" instead of "x:foo x:foobar". It is not
// guaranteed to always return the simplest query.
func AddRegexpField(q Query, field, pattern string) string {
	var modified bool
	q = MapParameter(q, func(gotField, value string, negated bool, annotation Annotation) Node {
		if field == gotField && strings.Contains(pattern, value) {
			value = pattern
			modified = true
		}
		return Parameter{
			Field:      gotField,
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !modified {
		// use newOperator to reduce And nodes when adding a parameter to the query toplevel.
		q = newOperator(append(q, Parameter{Field: field, Value: pattern}), And)
	}
	return StringHuman(q)
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
