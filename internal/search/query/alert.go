package query

import (
	"strings"
)

// addRegexpField adds a new expr to the query with the given field and pattern
// value. The nonnegated field is assumed to associate with a regexp value. The
// pattern value is assumed to be unquoted.
//
// It tries to remove redundancy in the result. For example, given
// a query like "x:foo", if given a field "x" with pattern "foobar" to add,
// it will return a query "x:foobar" instead of "x:foo x:foobar". It is not
// guaranteed to always return the simplest query.
func AddRegexpField(q Q, field, pattern string) string {
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
