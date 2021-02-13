package query

import (
	"regexp"
	"strconv"
)

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
	String *string        // if a string value, the string value (with escape sequences interpreted)
	Regexp *regexp.Regexp // if a regexp pattern, the compiled regular expression (call its String method to get source pattern string)
	Bool   *bool          // if a bool value, the bool value
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

func (v *Value) ToString() string {
	switch {
	case v.String != nil:
		return *v.String
	case v.Regexp != nil:
		return v.Regexp.String()
	case v.Bool != nil:
		return strconv.FormatBool(*v.Bool)
	default:
		return "<unable to get querytypes.Value as string>"
	}
}
