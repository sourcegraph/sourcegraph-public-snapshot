package promql

import (
	promqlparser "github.com/prometheus/prometheus/promql/parser"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Validate(expression string, vars VariableApplier) error {
	_, err := replaceAndParse(expression, vars)
	return err
}

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
