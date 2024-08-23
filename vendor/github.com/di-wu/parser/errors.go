package parser

import (
	"fmt"
	"github.com/di-wu/parser/op"
	"reflect"
	"strings"
)

// InitError is an error that occurs on instantiating new structures.
type InitError struct {
	// The error message. This should provide an intuitive message or advice on
	// how to solve this error.
	Message string
}

func (e *InitError) Error() string {
	return fmt.Sprintf("parser: %s", e.Message)
}

// ExpectError is an error that occurs on when an invalid/unsupported value is
// passed to the Parser.Expect function.
type ExpectError struct {
	Message string
}

func (e *ExpectError) Error() string {
	return fmt.Sprintf("expect: %s", e.Message)
}

// ExpectedParseError creates an ExpectedParseError error based on the given
// start and end cursor. Resets the parser tot the start cursor.
func (p *Parser) ExpectedParseError(expected interface{}, start, end *Cursor) *ExpectedParseError {
	if end == nil {
		end = start
	}
	defer p.Jump(start)
	return &ExpectedParseError{
		Expected: expected,
		String:   p.Slice(start, end),
		Conflict: *end,
	}
}

// ExpectedParseError indicates that the parser Expected a different value than
// the Actual value present in the buffer.
type ExpectedParseError struct {
	// The value that was expected.
	Expected interface{}
	// The value it actually got.
	String string
	// The position of the conflicting value.
	Conflict Cursor
}

func Stringer(i interface{}) string {
	i = ConvertAliases(i)
	if reflect.TypeOf(i).Kind() == reflect.Func {
		return "func"
	}

	switch v := i.(type) {
	case rune:
		return fmt.Sprintf("'%s'", string(v))
	case string:
		return fmt.Sprintf("%q", v)
	case op.Not:
		return fmt.Sprintf("!%s", Stringer(v.Value))
	case op.Ensure:
		return fmt.Sprintf("?%s", Stringer(v.Value))
	case op.And:
		and := make([]string, len(v))
		for i, v := range v {
			and[i] = Stringer(v)
		}
		return fmt.Sprintf("and[%s]", strings.Join(and, " "))
	case op.Or:
		or := make([]string, len(v))
		for i, v := range v {
			or[i] = Stringer(v)
		}
		return fmt.Sprintf("or[%s]", strings.Join(or, " "))
	case op.XOr:
		xor := make([]string, len(v))
		for i, v := range v {
			xor[i] = Stringer(v)
		}
		return fmt.Sprintf("xor[%s]", strings.Join(xor, " "))
	case op.Range:
		if v.Max == -1 {
			switch v.Min {
			case 0:
				return fmt.Sprintf("%s*", Stringer(v.Value))
			case 1:
				return fmt.Sprintf("%s+", Stringer(v.Value))
			}
		}
		return fmt.Sprintf("%s{%d:%d}", Stringer(v.Value), v.Min, v.Max)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (e *ExpectedParseError) Error() string {
	got := e.String
	if len(e.String) == 1 {
		got = fmt.Sprintf("'%s'", string([]rune(e.String)[0]))
	} else {
		got = fmt.Sprintf("%q", e.String)
	}

	return fmt.Sprintf(
		"parse conflict [%02d:%03d]: expected %T %s but got %s",
		e.Conflict.row, e.Conflict.column, e.Expected, Stringer(e.Expected), got,
	)
}

// UnsupportedType indicates the type of the value is unsupported.
type UnsupportedType struct {
	Value interface{}
}

func (e *UnsupportedType) Error() string {
	return fmt.Sprintf("parse: value of type %T are not supported", e.Value)
}
