package query

// VisitNode calls f on all nodes rooted at node.
func VisitNode(node Node, f func(node Node)) {
	switch v := node.(type) {
	case Parameter:
		f(v)
	case Operator:
		f(v)
		Visit(v.Operands, f)
	}
}

// Visit calls f on all nodes rooted at nodes.
func Visit(nodes []Node, f func(node Node)) {
	for _, node := range nodes {
		VisitNode(node, f)
	}
}

// VisitParameter calls f on all parameter nodes. f supplies the node's field,
// value, and whether the value is negated.
func VisitParameter(nodes []Node, f func(field, value string, negated bool)) {
	visitor := func(node Node) {
		if v, ok := node.(Parameter); ok {
			f(v.Field, v.Value, v.Negated)
		}
	}
	Visit(nodes, visitor)
}

// VisitField calls f on all parameter nodes whose field matches the field
// argument. f supplies the node's value and whether the value is negated.
func VisitField(nodes []Node, field string, f func(value string, negated bool)) {
	visitor := func(visitedField, value string, negated bool) {
		if field == visitedField {
			f(value, negated)
		}
	}
	VisitParameter(nodes, visitor)
}
