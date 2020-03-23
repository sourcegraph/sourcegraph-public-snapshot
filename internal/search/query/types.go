package query

import (
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
	"github.com/sourcegraph/sourcegraph/internal/search/query/types"
)

type SearchType int

const (
	SearchTypeRegex SearchType = iota
	SearchTypeLiteral
	SearchTypeStructural
)

// QueryInfo is an intermediate type for an interface of both ordinary queries
// and and/or query processing. The and/or query processing will become the
// canonical query form and the QueryInfo type will be removed.
type QueryInfo interface {
	RegexpPatterns(field string) (values, negatedValues []string)
	StringValues(field string) (values, negatedValues []string)
	StringValue(field string) (value, negatedValue string)
	Values(field string) []*types.Value
	Fields() map[string][]*types.Value
	IsCaseSensitive() bool
	ParseTree() syntax.ParseTree
}

// An ordinary query (not containing and/or expressions).
type OrdinaryQuery struct {
	Query     *Query           // the validated search query
	parseTree syntax.ParseTree // the parsed search query
}

// A query containing and/or expressions.
type AndOrQuery struct {
	query []Node
}

func (q OrdinaryQuery) RegexpPatterns(field string) (values, negatedValues []string) {
	return q.Query.RegexpPatterns(field)
}
func (q OrdinaryQuery) StringValues(field string) (values, negatedValues []string) {
	return q.Query.StringValues(field)
}
func (q OrdinaryQuery) StringValue(field string) (value, negatedValue string) {
	return q.Query.StringValue(field)
}
func (q OrdinaryQuery) Values(field string) []*types.Value {
	return q.Query.Values(field)
}
func (q OrdinaryQuery) Fields() map[string][]*types.Value {
	return q.Query.Fields
}
func (q OrdinaryQuery) ParseTree() syntax.ParseTree {
	return q.parseTree
}
func (q OrdinaryQuery) IsCaseSensitive() bool {
	return q.Query.IsCaseSensitive()
}

// AndOrQuery satisfies the interface for QueryInfo with empty values. Its
// methods are not currently used.
func (AndOrQuery) RegexpPatterns(field string) (values, negatedValues []string) {
	return nil, nil
}
func (AndOrQuery) StringValues(field string) (values, negatedValues []string) {
	return nil, nil
}
func (AndOrQuery) StringValue(field string) (value, negatedValue string) {
	return "", ""
}
func (AndOrQuery) Values(field string) []*types.Value {
	return nil
}
func (AndOrQuery) Fields() map[string][]*types.Value {
	return nil
}
func (AndOrQuery) ParseTree() syntax.ParseTree {
	return nil
}
func (AndOrQuery) IsCaseSensitive() bool {
	return false
}
