package query

import (
	"strings"
)

// LowercaseFieldNames performs strings.ToLower on every field name.
func LowercaseFieldNames(nodes []Node) []Node {
	return MapParameter(nodes, func(field, value string, negated bool) Node {
		return Parameter{Field: strings.ToLower(field), Value: value, Negated: negated}
	})
}
