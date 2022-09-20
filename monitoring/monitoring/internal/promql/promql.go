package promql

import (
	"github.com/prometheus/prometheus/model/labels"
	promqlparser "github.com/prometheus/prometheus/promql/parser"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Validate applies vars to the expression and asserts that the result is a valid PromQL
// expression.
func Validate(expression string, vars VariableApplier) error {
	_, err := replaceAndParse(expression, vars)
	return err
}

// Inject applies vars to the expression, parses the result into a PromQL AST, walks it
// to inject matchers, and renders it back to a string, using vars again to revert any
// replacements that occur.
func Inject(expression string, matchers []*labels.Matcher, vars VariableApplier) (string, error) {
	// Generate AST
	expr, err := replaceAndParse(expression, vars)
	if err != nil {
		return expression, err // return original
	}

	// Undo replacements if there are any
	revertExpr := func(e promqlparser.Expr) (string, error) {
		// Convert back to string, and revert injection of default values
		injected := expr.String()
		if vars != nil {
			return vars.RevertDefaults(expression, injected), nil
		}
		return injected, nil
	}

	if len(matchers) == 0 {
		return revertExpr(expr) // return formatted regardless, for consistency
	}

	// Inject matchers into selectors
	promqlparser.Inspect(expr, func(n promqlparser.Node, path []promqlparser.Node) error {
		if vec, ok := n.(*promqlparser.VectorSelector); ok {
			vec.LabelMatchers = append(vec.LabelMatchers, matchers...)
		}
		return nil
	})

	return revertExpr(expr)
}

// replaceAndParse applies vars to the expression and parses the result into a PromQL AST.
func replaceAndParse(expression string, vars VariableApplier) (promqlparser.Expr, error) {
	if vars != nil {
		expression = vars.ApplySentinelValues(expression)
	}
	expr, err := promqlparser.ParseExpr(expression)
	if err != nil {
		return nil, errors.Wrapf(err, "%q", expression)
	}
	return expr, nil
}
