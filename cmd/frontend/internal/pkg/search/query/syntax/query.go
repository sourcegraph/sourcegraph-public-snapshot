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

// EscapeImpossibleCaretsDollars puts a backslash in front of ^ if it occurs after the
// beginning and $ if it appears before the end of the query.
var unescapedCaretRx = regexp.MustCompile(`([^\\])\^`)
var unescapedDollarRx = regexp.MustCompile(`(^|[^\\])\$(.)`)
var initialCaretRx = regexp.MustCompile(`^\^`)
var finalUnescapedDollarRx = regexp.MustCompile(`(^|[^\\])\$$`)

func (q *Query) EscapeImpossibleCaretsDollars() {
	for i, e := range q.NonFieldExprs() {
		e2 := *e
		escape := func(s string) string {
			s = unescapedCaretRx.ReplaceAllString(s, `$1\^`)
			s = unescapedDollarRx.ReplaceAllString(s, `$1\$$$2`)
			if i > 0 {
				s = initialCaretRx.ReplaceAllString(s, `\^`)
			}
			if i+1 < len(q.Expr) {
				s = finalUnescapedDollarRx.ReplaceAllString(s, `$1\$$`)
			}
			return s
		}
		// escape is called twice to handle for example `^^^` which would
		// otherwise end up as `^\^^` whereas we want it to be escaped as
		// `^\^\^`.
		e2.Value = escape(escape(e2.Value))
		*e = e2
	}
	q.Input = q.String()
}

func (q *Query) FieldExprs() []*Expr {
	var es []*Expr
	for _, e := range q.Expr {
		if e.Field != "" {
			es = append(es, e)
		}
	}
	return es
}

func (q *Query) NonFieldExprs() []*Expr {
	var es []*Expr
	for _, e := range q.Expr {
		if e.Field == "" {
			es = append(es, e)
		}
	}
	return es
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
