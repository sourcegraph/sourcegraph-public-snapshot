package types

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

// A Query is the typechecked representation of a search query.
type Query struct {
	ParseTree syntax.ParseTree    // the query parse tree
	Fields    map[string][]*Value // map of field name -> values
}

func (q *Query) String() string {
	fields := []string{}
	for key, values := range q.Fields {
		for _, v := range values {
			switch s := v.Value().(type) {
			case string:
				fields = append(fields, fmt.Sprintf("%s:%q", key, s))
			case *regexp.Regexp:
				fields = append(fields, fmt.Sprintf("%s~%q", key, s))
			case bool:
				fields = append(fields, fmt.Sprintf("%s:%v", key, s))
			default:
				fields = append(fields, fmt.Sprintf("(UNKNOWN TYPE %s:%v)", key, s))
			}
		}
	}
	sort.Strings(fields)
	return strings.Join(fields, " ")
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
