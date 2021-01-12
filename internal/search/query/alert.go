package query

import (
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
