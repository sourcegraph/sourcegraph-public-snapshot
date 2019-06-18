package syntax

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// A Query contains the parse tree of a query.
type Query struct {
	Input string  // the original input query string
	Expr  []*Expr // expressions in this query
}

func (q *Query) String() string {
	return ExprString(q.Expr)
}

// WithErrorsQuoted converts a query like `f:foo b(ar` to `f:foo "b(ar"`.
func (q *Query) WithErrorsQuoted() *Query {
	q2 := &Query{}
	for _, e := range q.Expr {
		e2 := *e
		switch e.ValueType {
		case TokenError:
			e2.Value = fmt.Sprintf("%q", e.Value)
			e2.ValueType = TokenQuoted
		case TokenPattern, TokenLiteral:
			_, err := regexp.Compile(e2.Value)
			if err != nil {
				e2.Value = fmt.Sprintf("%q", e.Value)
				e2.ValueType = TokenQuoted
			}
		}
		q2.Expr = append(q2.Expr, &e2)
	}
	return q2
}

// WithPartsQuoted converts a query like `f:foo b(ar) ba+z` to one like `"f:foo" "b(ar)" "ba+z"`.
func (q *Query) WithPartsQuoted() *Query {
	q2 := &Query{}
	for _, e := range q.Expr {
		e2 := *e
		switch e.Field {
		case "":
			e2.Value = fmt.Sprintf("%q", e.Value)
		default:
			e2.Value = fmt.Sprintf("%q", fmt.Sprintf("%s:%s", e.Field, e.Value))
		}
		e2.Field = ""
		e2.ValueType = TokenQuoted
		q2.Expr = append(q2.Expr, &e2)
	}
	return q2
}

// WithNonFieldPartsQuoted converts a query like `f:foo b(ar) ba+z` to one like `f:foo "b(ar)" "ba+z"`.
func (q *Query) WithNonFieldPartsQuoted() *Query {
	q2 := &Query{}
	for _, e := range q.Expr {
		e2 := *e
		if e2.Field == "" {
			e2.Value = fmt.Sprintf("%q", e2.Value)
			e2.ValueType = TokenQuoted
		}
		q2.Expr = append(q2.Expr, &e2)
	}
	return q2
}

// WithNonFieldsQuoted converts a query like `f:foo b(ar) ba+z` to one like `f:foo "b(ar) ba+z"`.
func (q *Query) WithNonFieldsQuoted() *Query {
	q2 := &Query{}
	var vals []string
	for _, e := range q.Expr {
		e2 := *e
		if e2.Field == "" {
			// It's not a field. Add it to the big expression of non-fields.
			vals = append(vals, e2.Value)
		} else {
			// It's a field. Just append it.
			q2.Expr = append(q2.Expr, &e2)
		}
	}
	// Add the big expression of non-fields to the end, quoted.
	q2.Expr = append(q2.Expr, &Expr{
		ValueType: TokenQuoted,
		Value:     fmt.Sprintf("%q", strings.Join(vals, " ")),
	})
	return q2
}

// An Expr describes an expression in a query.
type Expr struct {
	Pos       int       // the starting character position of the query expression
	Not       bool      // the expression is negated (e.g., -term or -field:term)
	Field     string    // the field that this expression applies to
	Value     string    // the raw field value
	ValueType TokenType // the type of the value
}

func (e Expr) String() string {
	var buf bytes.Buffer
	if e.Not {
		buf.WriteByte('-')
	}
	if e.Field != "" {
		buf.WriteString(e.Field)
		buf.WriteByte(':')
	}
	if e.ValueType == TokenPattern {
		buf.WriteByte('/')
	}
	buf.WriteString(e.Value)
	if e.ValueType == TokenPattern {
		buf.WriteByte('/')
	}
	return buf.String()
}

// ExprString returns the query string that parses to expr.
func ExprString(expr []*Expr) string {
	s := make([]string, len(expr))
	for i, e := range expr {
		s[i] = e.String()
	}
	return strings.Join(s, " ")
}
