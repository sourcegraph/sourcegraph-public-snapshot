package search

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/batches/search/syntax"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TextSearchTerm represents a single term within a search string.
type TextSearchTerm struct {
	Term string
	Not  bool
}

// ParseTextSearch parses a free-form text search string into a slice of
// expressions, respecting quoted strings and negation.
func ParseTextSearch(search string) ([]TextSearchTerm, error) {
	tree, err := syntax.Parse(search)
	if err != nil {
		return nil, errors.Wrap(err, "parsing search string")
	}

	var errs error
	terms := []TextSearchTerm{}
	for _, expr := range tree {
		if expr.Field != "" {
			// In the future, we may choose to support field types in batch changes
			// text search queries. When that happens, we should extend this
			// function to accept an additional parameter defining field types
			// and what behaviour should be implemented when they are set. Until
			// then, we'll just error and keep this function simple.
			errs = errors.Append(errs, ErrUnsupportedField{
				ErrExpr: createErrExpr(search, expr),
				Field:   expr.Field,
			})
			continue
		}

		switch expr.ValueType {
		case syntax.TokenLiteral:
			terms = append(terms, TextSearchTerm{
				Term: expr.Value,
				Not:  expr.Not,
			})
		case syntax.TokenQuoted:
			terms = append(terms, TextSearchTerm{
				Term: strings.Trim(expr.Value, `"`),
				Not:  expr.Not,
			})
		// If we ever want to support regex patterns, this would be where we'd
		// hook it in (by matching TokenPattern).
		default:
			errs = errors.Append(errs, ErrUnsupportedValueType{
				ErrExpr:   createErrExpr(search, expr),
				ValueType: expr.ValueType,
			})
		}
	}

	return terms, errs
}
