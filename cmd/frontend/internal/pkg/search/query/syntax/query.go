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
		e2 := e.WithErrorsQuoted()
		q2.Expr = append(q2.Expr, &e2)
	}
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

// WithErrorsQuoted returns a new version of the expression,
// quoting in case of TokenError or an invalid regular expression.
func (e Expr) WithErrorsQuoted() Expr {
	e2 := e
	needsQuoting := false
	switch e.ValueType {
	case TokenError:
		needsQuoting = true
	case TokenPattern, TokenLiteral:
		_, err := regexp.Compile(e2.Value)
		if err != nil {
			needsQuoting = true
		}
	}
	if needsQuoting {
		e2.Not = false
		e2.Field = ""
		e2.Value = fmt.Sprintf("%q", e.String())
		e2.ValueType = TokenQuoted
	}
	return e2
}

// ExprString returns the query string that parses to expr.
func ExprString(expr []*Expr) string {
	s := make([]string, len(expr))
	for i, e := range expr {
		s[i] = e.String()
	}
	return strings.Join(s, " ")
}
