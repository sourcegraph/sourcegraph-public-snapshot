package search

import (
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

// ParseChangesetSearch parses the given search string into a set of options
// that can be given to ListChangesets().
//
// At present, the only field that will be set in the options is TextSearch.
// This will change in the future as we start to support field operators.
func ParseChangesetSearch(search string) (*campaigns.ListChangesetsOpts, error) {
	tree, err := syntax.Parse(search)
	if err != nil {
		return nil, errors.Wrap(err, "parsing search string")
	}

	opts := campaigns.ListChangesetsOpts{
		TextSearch: make([]campaigns.ListChangesetsTextSearchExpr, 0),
	}
	var errs *multierror.Error
	for _, expr := range tree {
		if expr.Field != "" {
			// Eventually, we'll support some field types and these will set
			// other options in the result. For now, though, this is an error.
			errs = multierror.Append(errs, ErrUnsupportedField{
				ErrExpr: createErrExpr(search, expr),
				Field:   expr.Field,
			})
			continue
		}

		switch expr.ValueType {
		case syntax.TokenLiteral:
			opts.TextSearch = append(opts.TextSearch, campaigns.ListChangesetsTextSearchExpr{
				Term: expr.Value,
				Not:  expr.Not,
			})
		case syntax.TokenQuoted:
			opts.TextSearch = append(opts.TextSearch, campaigns.ListChangesetsTextSearchExpr{
				Term: strings.Trim(expr.Value, `"`),
				Not:  expr.Not,
			})
		// If we ever want to support regex patterns, this would be where we'd
		// hook it in (by matching TokenPattern).
		default:
			errs = multierror.Append(errs, ErrUnsupportedValueType{
				ErrExpr:   createErrExpr(search, expr),
				ValueType: expr.ValueType,
			})
		}
	}

	return &opts, errs.ErrorOrNil()
}
