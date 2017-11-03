package search2

import "strconv"

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

// Tokens is a list of tokens parsed from a query.
type Tokens []Token

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

// Extract extracts field values and terms from the tokens list. The fieldAliases
// argument specifies each valid field as a map key, and an optional list of its
// aliases as the corresponding value. If an alias is ambiguous, it panics.
//
// To support terms (tokens without a field), include a "" key in fieldAliases (and
// a "-" key to support negations of terms).
//
// For example, if "x:foo" is shorthand for "expr:foo", then "x" is a field alias
// of "expr".
func (ts Tokens) Extract(fieldAliases map[Field][]Field) (fieldValues map[Field]Values, unknownFields []Field) {
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

	fieldValues = map[Field]Values{}

	for _, t := range ts {
		field, ok := fieldNames[t.Field]
		if !ok {
			unknownFields = append(unknownFields, t.Field)
			continue
		}

		values := fieldValues[field]
		values = append(values, t.Value)
		fieldValues[field] = values
	}

	return
}
