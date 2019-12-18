package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

// TypeError describes an error in query typechecking.
type TypeError struct {
	Pos int   // the character position where the error occurred
	Err error // the error
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("type error at character %d: %s", e.Pos, e.Err)
}

// Config specifies configuration for parsing a query.
type Config struct {
	FieldTypes   map[string]FieldType // map of recognized field name (excluding aliases) -> type
	FieldAliases map[string]string    // map of field alias -> field name
}

// FieldType describes the type of a query field.
type FieldType struct {
	Literal   ValueType // interpret literal tokens as being of this type
	Quoted    ValueType // interpret literal tokens as being of this type
	Singular  bool      // whether the field may only be used 0 or 1 times
	Negatable bool      // whether the field can be matched negated (i.e., -field:value)

	// FeatureFlagEnabled returns true if this field is enabled.
	// The field is always enabled if this is nil.
	FeatureFlagEnabled func() bool
}

// Check typechecks the input query for field and type validity.
func (c *Config) Check(parseTree syntax.ParseTree) (*Query, error) {
	checkedQuery := Query{
		ParseTree: parseTree,
		Fields:    map[string][]*Value{},
	}
	for _, expr := range parseTree {
		field, fieldType, value, err := c.checkExpr(expr)
		if err != nil {
			return nil, err
		}
		if fieldType.Singular && len(checkedQuery.Fields[field]) >= 1 {
			return nil, &TypeError{Pos: expr.Pos, Err: fmt.Errorf("field %q may not be used more than once", field)}
		}
		checkedQuery.Fields[field] = append(checkedQuery.Fields[field], value)
	}
	return &checkedQuery, nil
}

func (c *Config) resolveField(field string, not bool) (resolvedField string, typ FieldType, err error) {
	// Resolve field alias, if any.
	if resolvedField, ok := c.FieldAliases[field]; ok {
		field = resolvedField
	}

	// Check that field is recognized.
	var ok bool
	typ, ok = c.FieldTypes[field]
	if !ok {
		err = fmt.Errorf("unrecognized field %q", field)
		return
	}
	if typ.FeatureFlagEnabled != nil && !typ.FeatureFlagEnabled() {
		err = fmt.Errorf("unrecognized field %q; the feature flag for this field is not enabled", field)
		return
	}
	if not && !typ.Negatable {
		if field == "" {
			err = errors.New("negated terms (-term) are not yet supported")
		} else {
			err = fmt.Errorf("field %q does not support negation", field)
		}
		return
	}
	return field, typ, nil
}

func (c *Config) checkExpr(expr *syntax.Expr) (field string, fieldType FieldType, value *Value, err error) {
	// Resolve field name.
	resolvedField, fieldType, err := c.resolveField(expr.Field, expr.Not)
	if err != nil {
		return "", FieldType{}, nil, &TypeError{Pos: expr.Pos, Err: err}
	}

	// Resolve value.
	value = &Value{syntax: expr}
	switch expr.ValueType {
	case syntax.TokenLiteral:
		if err := setValue(value, expr.Value, fieldType.Literal); err != nil {
			return "", FieldType{}, nil, &TypeError{Pos: expr.Pos, Err: err}
		}

	case syntax.TokenQuoted:
		stringValue, err := unquoteString(expr.Value)
		if err != nil {
			return "", FieldType{}, nil, &TypeError{Pos: expr.Pos, Err: err}
		}
		if err := setValue(value, stringValue, fieldType.Quoted); err != nil {
			return "", FieldType{}, nil, &TypeError{Pos: expr.Pos, Err: err}
		}

	case syntax.TokenPattern:
		if err := setValue(value, expr.Value, RegexpType); err != nil {
			return "", FieldType{}, nil, &TypeError{Pos: expr.Pos, Err: err}
		}
	}

	return resolvedField, fieldType, value, nil
}

func setValue(dst *Value, valueString string, valueType ValueType) error {
	switch valueType {
	case StringType:
		dst.String = &valueString
	case RegexpType:
		valueString = autoFix(valueString)
		p, err := regexp.Compile(valueString)
		if err != nil {
			return err
		}
		dst.Regexp = p
	case BoolType:
		b, err := parseBool(valueString)
		if err != nil {
			return err
		}
		dst.Bool = &b
	default:
		return errors.New("no type for literal")
	}
	return nil
}

var leftParenRx = lazyregexp.New(`([^\\]|^)\($`)
var squareBraceRx = lazyregexp.New(`([^\\]|^)\[$`)

// autoFix escapes various patterns that are very likely not meant to be
// treated as regular expressions.
func autoFix(pat string) string {
	pat = leftParenRx.ReplaceAllString(pat, `$1\(`)
	pat = squareBraceRx.ReplaceAllString(pat, `$1\[`)
	pat = escapeAll(pat, "()", `\(\)`)
	return pat
}

// escapeAll replaces all instances of old with new. However, it will not
// replace old if it is already escaped.
func escapeAll(s, old, new string) string {
	var b strings.Builder
	b.Grow(len(s))
	for len(s) > 0 {
		i := strings.Index(s, old)
		if i == -1 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		if i == 0 || s[i-1] != '\\' {
			b.WriteString(new)
		} else {
			b.WriteString(old)
		}
		s = s[i+len(old):]
	}
	return b.String()
}

// unquoteString is like strings.Unquote except that it supports single-quoted
// strings with more than 1 character.
func unquoteString(s string) (string, error) {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = `"` + strings.Replace(s[1:len(s)-1], `"`, `\"`, -1) + `"`
	}
	s2, err := strconv.Unquote(s)
	if err != nil {
		err = fmt.Errorf("invalid quoted string: %s", s)
	}
	return s2, err
}

// parseBool is like strconv.ParseBool except that it also accepts y, Y, yes,
// YES, Yes, n, N, no, NO, No.
func parseBool(s string) (bool, error) {
	switch s {
	case "y", "Y", "yes", "YES", "Yes":
		return true, nil
	case "n", "N", "no", "NO", "No":
		return false, nil
	default:
		b, err := strconv.ParseBool(s)
		if err != nil {
			err = fmt.Errorf("invalid boolean %q", s)
		}
		return b, err
	}
}
