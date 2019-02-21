package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
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
func (c *Config) Check(query *syntax.Query) (*Query, error) {
	checkedQuery := Query{
		Syntax: query,
		Fields: map[string][]*Value{},
	}
	for _, expr := range query.Expr {
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
		p, err := compileRegexp(valueString)
		if err != nil {
			return err
		}
		dst.Regexp = p
	case BoolType:
		b, err := ParseBool(valueString)
		if err != nil {
			return err
		}
		dst.Bool = &b
	default:
		return errors.New("no type for literal")
	}
	return nil
}

func compileRegexp(value string) (*regexp.Regexp, error) {
	var r *regexp.Regexp
	var err error

	v := preprocessRegexpQuery(value)

	r, err = regexp.Compile(v)
	if err != nil {
		return fixupRegexpCompileError(v, err)
	}

	return r, err
}

// preprocessRegexpQuery looks for common mistakes in regexp search queries that
// don't cause regexp compile errors and fix them beforehand.
func preprocessRegexpQuery(value string) string {
	return escapeDollarSigns(value)
}

func escapeDollarSigns(value string) string {
	out := ""

	for i, r := range value {
		fmt.Println(r, r == '$', i == len(value)-1)

		if r == '$' && i != len(value)-1 {
			if i == 0 || (i != 0 && value[i-1] != '\\') {
				out += string('\\')
			}
		}

		out += string(r)
	}

	return out
}

var escapeErrorMessages = []string{
	"missing argument to repetition operator: `",
	"missing closing ",
}

var unmatchedOpeningRuneRegexps = map[rune]*regexp.Regexp{
	'(': regexp.MustCompile(`\([^\)]*$`),
	'[': regexp.MustCompile(`\[[^\]]*$`),
}

func fixupRegexpCompileError(value string, err error) (*regexp.Regexp, error) {
	msg := err.Error()
	index := -1

	for _, errorMsg := range escapeErrorMessages {
		index = strings.Index(msg, errorMsg)
		if index > -1 {
			index = len(errorMsg) + index
			break
		}
	}

	if index == -1 {
		return nil, err
	}

	runeToEscape := flipRune(rune(msg[index]))
	if runeToEscape == '*' {
		toEscape := string(runeToEscape)
		escaper := strings.NewReplacer(toEscape, `\`+toEscape)

		correctedValue := escaper.Replace(value)

		return regexp.Compile(correctedValue)
	}

	if r := unmatchedOpeningRuneRegexps[runeToEscape]; r != nil {
		fmt.Println(string(runeToEscape), r)
		match := r.FindStringIndex(value)
		correctedValue := fmt.Sprintf(`%s\%s`, value[:match[0]], value[match[0]:])

		return regexp.Compile(correctedValue)
	}

	return nil, err
}

// flipRune maps opening block characters (e.g. ), ]) to their opening
// counterparts. If the rune provided is not one of those, this func returns
// the identity of the rune.
func flipRune(r rune) rune {
	switch r {
	case 41: // )
		return r - 1
	case 93, 125: // ]
		return r - 2
	}

	return r
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
func ParseBool(s string) (bool, error) {
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
