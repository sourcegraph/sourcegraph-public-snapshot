package promql

import (
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

func parse(expression string, vars VariableApplier) (promqlparser.Expr, error) {
	applied := vars.ApplyDefaults(expression)
	expr, err := promqlparser.ParseExpr(applied)
	if err != nil {
		return nil, errors.Wrapf(err, "%q", expression)
	}
	return expr, nil
}
