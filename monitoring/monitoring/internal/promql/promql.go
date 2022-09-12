package promql

import (
	"github.com/prometheus/prometheus/model/labels"
	promqlparser "github.com/prometheus/prometheus/promql/parser"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// https://github.com/cherti/promql-labelinjector/blob/main/promql-labelinjector.go

// VariableApplier converts Prometheus expressions with template variables into valid
// Prometheus expressions, and vice versa.
type VariableApplier interface {
	// ApplyDefaults applies default variable values to the expression, such that the
	// expression is a valid Prometheus query.
	ApplyDefaults(expression string) string
	// RevertDefaults returns the expression that has been modified through ApplyDefaults
	// and revert any defaults applied to it.
	RevertDefaults(applied string) string
}

func Validate(expression string, vars VariableApplier) error {
	_, err := parse(expression, vars)
	return err
}

func Inject(expression string, matchers []*labels.Matcher, vars VariableApplier) (string, error) {
	expr, err := parse(expression, vars)
	if err != nil {
		return expression, err
	}

	promqlparser.Inspect(expr, func(n promqlparser.Node, path []promqlparser.Node) error {
		if vec, ok := n.(*promqlparser.VectorSelector); ok {
			vec.LabelMatchers = append(vec.LabelMatchers, matchers...)
		}
		return nil
	})

	injected := expr.String()
	if vars != nil {
		return vars.RevertDefaults(injected), nil
	}
	return injected, nil
}

func parse(expression string, vars VariableApplier) (promqlparser.Expr, error) {
	if vars != nil {
		expression = vars.ApplyDefaults(expression)
	}
	expr, err := promqlparser.ParseExpr(expression)
	if err != nil {
		return nil, errors.Wrapf(err, "%q", expression)
	}
	return expr, nil
}
