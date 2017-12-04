// Package searchquery provides facilities for parsing and extracting
// information from search queries.
package searchquery

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/syntax"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/types"
)

// All field names.
const (
	FieldDefault   = ""
	FieldCase      = "case"
	FieldRepo      = "repo"
	FieldRepoGroup = "repogroup"
	FieldFile      = "file"
	FieldType      = "type"

	// For diff and commit search only:
	FieldBefore    = "before"
	FieldAfter     = "after"
	FieldAuthor    = "author"
	FieldCommitter = "committer"
	FieldMessage   = "message"
)

var (
	regexpFieldType          = types.FieldType{Literal: types.RegexpType, Quoted: types.RegexpType}
	regexpNegatableFieldType = types.FieldType{Literal: types.RegexpType, Quoted: types.RegexpType, Negatable: true}
	stringFieldType          = types.FieldType{Literal: types.StringType, Quoted: types.StringType}

	conf = types.Config{
		FieldTypes: map[string]types.FieldType{
			FieldDefault:   {Literal: types.RegexpType, Quoted: types.StringType},
			FieldCase:      {Literal: types.BoolType, Quoted: types.BoolType, Singular: true},
			FieldRepo:      regexpNegatableFieldType,
			FieldRepoGroup: types.FieldType{Literal: types.StringType, Quoted: types.StringType, Singular: true},
			FieldFile:      regexpNegatableFieldType,
			FieldType:      stringFieldType,
			FieldBefore:    stringFieldType,
			FieldAfter:     stringFieldType,
			FieldAuthor:    regexpNegatableFieldType,
			FieldCommitter: regexpNegatableFieldType,
			FieldMessage:   regexpNegatableFieldType,
		},
		FieldAliases: map[string]string{
			"r":     FieldRepo,
			"f":     FieldFile,
			"since": FieldAfter,
			"until": FieldBefore,
			"m":     FieldMessage,
			"msg":   FieldMessage,
		},
	}
)

// A Query is the parsed representation of a search query.
type Query struct {
	conf *types.Config // the typechecker config used to produce this query

	*types.Query // the underlying query
}

// ParseAndCheck parses and typechecks a search query using the default
// query type configuration.
func ParseAndCheck(input string) (*Query, error) {
	return parseAndCheck(&conf, input)
}

func parseAndCheck(conf *types.Config, input string) (*Query, error) {
	syntaxQuery, err := syntax.Parse(input)
	if err != nil {
		return nil, err
	}
	checkedQuery, err := conf.Check(syntaxQuery)
	if err != nil {
		return nil, err
	}
	return &Query{conf: conf, Query: checkedQuery}, nil
}

// IsCaseSensitive reports whether the query's expressions are matched
// case sensitively.
func (q *Query) IsCaseSensitive() bool {
	for _, v := range q.Fields[FieldCase] {
		if v.Bool != nil {
			return *v.Bool
		}
	}
	return false // default
}

// Values returns the values for the given field.
func (q *Query) Values(field string) []*types.Value {
	if _, ok := q.conf.FieldTypes[field]; !ok {
		panic("no such field: " + field)
	}
	return q.Fields[field]
}

// RegexpPatterns returns the regexp pattern source strings for the given field.
// If the field is not recognized or it is not always regexp-typed, it panics.
func (q *Query) RegexpPatterns(field string) (values, negatedValues []string) {
	fieldType, ok := q.conf.FieldTypes[field]
	if !ok {
		panic("no such field: " + field)
	}
	if fieldType.Literal != types.RegexpType || fieldType.Quoted != types.RegexpType {
		panic("field is not always regexp-typed: " + field)
	}

	for _, v := range q.Fields[field] {
		s := v.Regexp.String()
		if v.Not() {
			negatedValues = append(negatedValues, s)
		} else {
			values = append(values, s)
		}
	}
	return
}

// StringValues returns the string values for the given field. If the field is
// not recognized or it is not always string-typed, it panics.
func (q *Query) StringValues(field string) (values, negatedValues []string) {
	fieldType, ok := q.conf.FieldTypes[field]
	if !ok {
		panic("no such field: " + field)
	}
	if fieldType.Literal != types.StringType || fieldType.Quoted != types.StringType {
		panic("field is not always string-typed: " + field)
	}

	for _, v := range q.Fields[field] {
		if v.Not() {
			negatedValues = append(negatedValues, *v.String)
		} else {
			values = append(values, *v.String)
		}
	}
	return
}
