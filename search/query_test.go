package search

import (
	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Q is an alias for newQueryWithInsertionPoint.
func Q(s string) sourcegraph.RawQuery { return newQueryWithInsertionPoint(s) }

// newQueryWithInsertionPoint creates a new query with no Type field
// set. The query string may contain a '•' character that marks the
// insertion point. If it is present, it is removed from the string
// and its character offset is used as the query's
// InsertionPoint. Otherwise the InsertionPoint is set to -1
// (essentially making no tokens active).
func newQueryWithInsertionPoint(s string) sourcegraph.RawQuery {
	const bullet = "•"
	return sourcegraph.RawQuery{
		Text:           strings.Replace(s, bullet, "", 1),
		InsertionPoint: int32(strings.Index(s, bullet)),
	}
}
