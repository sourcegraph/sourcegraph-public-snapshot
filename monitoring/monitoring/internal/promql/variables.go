package promql

import (
	"fmt"
	"strings"
)

// VariableApplier converts Prometheus expressions with template variables into valid
// Prometheus expressions, and vice versa.
type VariableApplier map[string]string

// ApplySentinelValues applies default sentinel variable values to the expression, such
// that the expression is a valid Prometheus query.
func (vars VariableApplier) ApplySentinelValues(expression string) string {
	for name, sentinelValue := range vars {
		varKey := "$" + name

		// If the expression uses the variable in a quoted context ("$var") then it's=
		// interpreted as valid PromQL, we don't need to replace it!
		if strings.Contains(expression, fmt.Sprintf("%q", varKey)) {
			continue
		}

		// Otherwise replace all occurrences.
		expression = strings.ReplaceAll(expression, varKey, sentinelValue)
	}
	return expression
}
