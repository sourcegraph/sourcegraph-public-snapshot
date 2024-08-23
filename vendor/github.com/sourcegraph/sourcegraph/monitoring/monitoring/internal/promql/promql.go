package promql

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"

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

// InjectMatchers applies vars to the expression, parses the result into a PromQL AST,
// walks it to inject matchers, and renders it back to a string, using vars again to
// revert any replacements that occur.
func InjectMatchers(expression string, matchers []*labels.Matcher, vars VariableApplier) (string, error) {
	// Generate AST
	expr, err := replaceAndParse(expression, vars)
	if err != nil {
		return expression, err // return original
	}

	// Undo replacements if there are any
	revertExpr := func() (string, error) {
		// Convert back to string, and revert injection of default values
		injected := expr.String()
		if vars != nil {
			return vars.RevertDefaults(expression, injected), nil
		}
		return injected, nil
	}

	if len(matchers) == 0 {
		return revertExpr() // return formatted regardless, for consistency
	}

	// Inject matchers into selectors
	promqlparser.Inspect(expr, func(n promqlparser.Node, path []promqlparser.Node) error {
		if vec, ok := n.(*promqlparser.VectorSelector); ok {
			vec.LabelMatchers = append(vec.LabelMatchers, matchers...)
		}
		return nil
	})

	return revertExpr()
}

type inspector func(promqlparser.Node, []promqlparser.Node) error

func (f inspector) Visit(node promqlparser.Node, path []promqlparser.Node) (promqlparser.Visitor, error) {
	if err := f(node, path); err != nil {
		return nil, err
	}
	return f, nil
}

// InjectAsAlert does the same thing as Inject, but also converts expression into a valid
// query that can be used for alerting by removing selectors with variable values, or an
// error if it can't.
func InjectAsAlert(expression string, matchers []*labels.Matcher, vars VariableApplier) (string, error) {
	// Generate AST
	expr, err := replaceAndParse(expression, vars)
	if err != nil {
		return expression, err // return original
	}

	// Inject matchers into selectors, but also remove selectors that have variables in
	// them.
	err = promqlparser.Walk(inspector(func(n promqlparser.Node, path []promqlparser.Node) error {
		if vec, ok := n.(*promqlparser.VectorSelector); ok {
			validMatchers := make([]*labels.Matcher, 0, len(vec.LabelMatchers)+len(matchers))
			for _, lm := range vec.LabelMatchers {
				// vars.ApplySentinelValues does not replace vars that are used in string
				// values, so we will find them here in the value intact
				var hasVar bool
				for varName, sentinelValue := range vars {
					// We use regexp here because we want to be stricter than
					// VariableApplier - we need to catch any possible usage of this var.
					varKey, err := newVarKeyRegexp(varName)
					if err != nil {
						return errors.Wrapf(err, "generating regexp for variable %q", varName)
					}
					reValue := lm.GetRegexString()
					if varKey.MatchString(lm.Value) || varKey.MatchString(reValue) {
						hasVar = true
						break
					}
					// If the regexp match value contains this variable's sentinel value,
					// it means this variable was used in a regexp match, and should use
					// Grafana's '${variable:regex}' instead.
					if strings.Contains(reValue, sentinelValue) {
						return errors.Newf("unexpected sentinel value found in value of %q - you may want to use '${variable:regex}' instead", lm.String())
					}
				}
				if !hasVar {
					validMatchers = append(validMatchers, lm)
				}
			}

			vec.LabelMatchers = append(validMatchers, matchers...)
		}
		return nil
	}), expr, nil)
	if err != nil {
		return expression, errors.Wrap(err, "walk promql") // return original
	}

	// Revert any remaining variables
	rendered := expr.String()
	if vars != nil {
		rendered = vars.RevertDefaults(expression, rendered)
	}

	// Validate that the result is a valid query for use in alerting
	if _, err := promqlparser.ParseExpr(rendered); err != nil {
		return rendered, errors.Wrap(err, "invalid alert expression")
	}

	return rendered, nil
}

// Prometheus histograms require all 3 metrics in the set: https://prometheus.io/docs/practices/histograms/
//
// This map maps suffixes to the other 2 metrics in a set. If one is used, they must
// all be listed.
var histogramSuffixes = map[string][]string{
	"_count":  {"_sum", "_bucket"},
	"_sum":    {"_count", "_bucket"},
	"_bucket": {"_count", "_sum"},
}

// ListMetrics returns all unique metrics used in the expression.
func ListMetrics(expression string, vars VariableApplier) ([]string, error) {
	// Generate AST
	expr, err := replaceAndParse(expression, vars)
	if err != nil {
		return nil, err // return original
	}

	// Collect all metrics mentioned in the expression
	foundMetrics := make(map[string]struct{})
	var metrics []string
	addMetric := func(m string) {
		if _, exists := foundMetrics[m]; !exists {
			metrics = append(metrics, m)
			foundMetrics[m] = struct{}{}
		}
	}

	promqlparser.Inspect(expr, func(n promqlparser.Node, path []promqlparser.Node) error {
		if vec, ok := n.(*promqlparser.VectorSelector); ok {
			// Handle '{__name__=~"..."}' selectors
			if vec.Name == "" {
				for _, matcher := range vec.LabelMatchers {
					if matcher.Name == "__name__" {
						// This may be an arbitrary regex or something, but oh well
						addMetric(matcher.Value)
					}
				}
			} else {
				// Otherwise just add the vector
				addMetric(vec.Name)

				// If vector is part of a histogram set, add all the other metrics in the
				// set.
				for suffix, otherSuffixes := range histogramSuffixes {
					if strings.HasSuffix(vec.Name, suffix) {
						root := strings.TrimSuffix(vec.Name, suffix)
						for _, s := range otherSuffixes {
							addMetric(root + s)
						}
					}
				}
			}
		}
		return nil
	})
	return metrics, nil
}

// InjectGroupings applies vars to the expression, parses the result into a PromQL AST,
// walks it to add the provided groupings to all aggregation expressions, and renders it
// back to a string, using vars again to revert any replacements that occur.
func InjectGroupings(expression string, groupings []string, vars VariableApplier) (string, error) {
	// Generate AST
	expr, err := replaceAndParse(expression, vars)
	if err != nil {
		return expression, err // return original
	}

	// Undo replacements if there are any
	revertExpr := func() (string, error) {
		// Convert back to string, and revert injection of default values
		injected := expr.String()
		if vars != nil {
			return vars.RevertDefaults(expression, injected), nil
		}
		return injected, nil
	}

	if len(groupings) == 0 {
		return revertExpr() // return formatted regardless, for consistency
	}

	// Inject aggregators into selectors
	promqlparser.Inspect(expr, func(n promqlparser.Node, path []promqlparser.Node) error {
		if agg, ok := n.(*promqlparser.AggregateExpr); ok {
			agg.Grouping = append(agg.Grouping, groupings...)
		}

		return nil
	})

	return revertExpr()
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

const varKeyRegexpFormat = `(\$%[1]s|\${%[1]s}|\${%[1]s:[^}]*})`

func newVarKeyRegexp(name string) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf(varKeyRegexpFormat, name))
}
