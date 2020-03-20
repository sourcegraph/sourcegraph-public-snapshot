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

type QueryInfo interface {
	RegexpPatterns(field string) (values, negatedValues []string)
	StringValues(field string) (values, negatedValues []string)
	StringValue(field string) (value, negatedValue string)
	Values(field string) []*types.Value
	Fields() map[string][]*types.Value
	IsCaseSensitive() bool
	ParseTree() syntax.ParseTree
}

func (q OrdinaryQuery) RegexpPatterns(field string) (values, negatedValues []string) {
	return q.query.RegexpPatterns(field)
}
func (q OrdinaryQuery) StringValues(field string) (values, negatedValues []string) {
	return q.query.StringValues(field)
}
func (q OrdinaryQuery) StringValue(field string) (value, negatedValue string) {
	return q.query.StringValue(field)
}
func (q OrdinaryQuery) Values(field string) []*types.Value {
	return q.query.Values(field)
}
func (q OrdinaryQuery) Fields() map[string][]*types.Value {
	return q.query.Fields
}
func (q OrdinaryQuery) ParseTree() syntax.ParseTree {
	return q.parseTree
}
func (q OrdinaryQuery) IsCaseSensitive() bool {
	return q.query.IsCaseSensitive()
}

func (AndOrQuery) RegexpPatterns(field string) (values, negatedValues []string) {
	panic("")
}
func (AndOrQuery) StringValues(field string) (values, negatedValues []string) {
	panic("")
}
func (AndOrQuery) StringValue(field string) (value, negatedValue string) {
	panic("")
}
func (AndOrQuery) Values(field string) []*types.Value {
	panic("")
}
func (AndOrQuery) Fields() map[string][]*types.Value {
	panic("")
}
func (AndOrQuery) ParseTree() syntax.ParseTree {
	panic("")
}
func (AndOrQuery) IsCaseSensitive() bool {
	panic("")
}

// An ordinary query, corresponding to a single search operation.
type OrdinaryQuery struct {
	query     *Query           // the validated search query
	parseTree syntax.ParseTree // the parsed search query
}

// A query containing and/or expressions.
type AndOrQuery struct {
	query []Node
}
