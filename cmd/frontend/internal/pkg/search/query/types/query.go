package types

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
)

// A Query is the typechecked representation of a search query.
type Query struct {
	Syntax *syntax.Query       // the query syntax
	Fields map[string][]*Value // map of field name -> values
}

// ValueType is the set of types of values in queries.
type ValueType int

// All ValueType values.
const (
	StringType ValueType = 1 << iota
	RegexpType
	BoolType
)

// A Value is a field value in a query.
type Value struct {
	syntax *syntax.Expr // the underlying query expression

	String *string        // if a string value, the string value (with escape sequences interpreted)
	Regexp *regexp.Regexp // if a regexp pattern, the compiled regular expression (call its String method to get source pattern string)
	Bool   *bool          // if a bool value, the bool value
}

func (v *Value) SyntaxValue() string {
	return v.syntax.Value
}


// Not returns whether the value is negated in the query (e.g., -value or -field:value).
func (v *Value) Not() bool {
	return v.syntax.Not
}

// Value returns the value as an interface{}.
func (v *Value) Value() interface{} {
	switch {
	case v.String != nil:
		return *v.String
	case v.Regexp != nil:
		return v.Regexp
	case v.Bool != nil:
		return *v.Bool
	default:
		panic("no value")
	}
}
