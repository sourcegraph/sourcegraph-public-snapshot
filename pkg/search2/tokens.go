package search2

import (
	"strconv"
	"strings"
)

// Field is the name of a field (e.g., "x" in the token "x:foo").
//
// A field prefixed with "-" conventionally means that it is negated.
type Field string

// Token is the smallest unit parsed from a query.
type Token struct {
	// Field is the name of the field that the value applies to (e.g,
	// "x" in the token "x:foo"). If the token is a string token, then
	// Field is empty (or "-" if negated).
	Field

	// Value is the value of the field.
	Value
}

func (t Token) String() string {
	switch t.Field {
	case "":
		return t.Value.String()
	case "-":
		return "-" + t.Value.String()
	default:
		return string(t.Field) + ":" + t.Value.String()
	}
}

// Value represents the value of a token.
type Value struct {
	// Value is the value of the token (e.g. "foo" in the token "x:foo").
	Value string

	// Quoted is whether the value was double-quoted (e.g., `x:"foo"` vs.
	// `x:foo`).
	Quoted bool
}

func (v Value) String() string {
	if v.Quoted {
		return strconv.Quote(v.Value)
	}
	return v.Value
}

// Values is a list of values.
type Values []Value

// Values returns a slice of the string value of each item in vs.
func (vs Values) Values() []string {
	if vs == nil {
		return nil
	}
	ss := make([]string, len(vs))
	for i, v := range vs {
		ss[i] = v.Value
	}
	return ss
}

// Tokens is a list of tokens parsed from a query.
type Tokens []Token

func (ts Tokens) String() string {
	ss := make([]string, len(ts))
	for i, t := range ts {
		ss[i] = t.String()
	}
	return strings.Join(ss, " ")
}

// UnmarshalText implements encoding.TextMarshaler.
func (ts *Tokens) UnmarshalText(text []byte) error {
	tokens, err := Parse(string(text))
	if err != nil {
		return err
	}
	*ts = tokens
	return nil
}

// Normalize modifies ts, dealiasing the field of each token based on fieldAliases.
// The fieldAliases argument specifies each valid field as a map key, and an optional
// list of its aliases as the corresponding value. If an alias is ambiguous, it panics.
//
// For example, if "x:foo" is shorthand for "expr:foo", then "x" is a field alias
// of "expr".
func (ts Tokens) Normalize(fieldAliases map[Field][]Field) {
	fieldNames := map[Field]Field{}
	for name := range fieldAliases {
		fieldNames[name] = name
	}
	for name, aliases := range fieldAliases {
		for _, alias := range aliases {
			if alias == "" {
				panic("field alias must be non-empty string")
			}
			if _, present := fieldNames[alias]; present {
				panic("field alias " + alias + " is ambiguous")
			}
			fieldNames[alias] = name
		}
	}

	for i, t := range ts {
		field, ok := fieldNames[t.Field]
		if ok {
			ts[i].Field = field
		}
	}
}

// Extract returns field values grouped by the field name.
func (ts Tokens) Extract() (fieldValues map[Field]Values) {
	fieldValues = map[Field]Values{}
	for _, t := range ts {
		values := fieldValues[t.Field]
		values = append(values, t.Value)
		fieldValues[t.Field] = values
	}
	return
}
