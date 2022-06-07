package query

type Visitor struct {
	Operator  func(kind OperatorKind, operands []Node)
	Parameter func(field, value string, negated bool, annotation Annotation)
	Pattern   func(value string, negated bool, annotation Annotation)
}

// Visit recursively visits each node in a query. Need a visitor that
// returns early or doesn't recurse? Use this function as a template and
// customize it for your task!
func (v *Visitor) Visit(node Node) {
	switch n := node.(type) {
	case Operator:
		if v.Operator != nil {
			v.Operator(n.Kind, n.Operands)
		}
		for _, child := range n.Operands {
			v.Visit(child)
		}

	case Parameter:
		if v.Parameter != nil {
			v.Parameter(n.Field, n.Value, n.Negated, n.Annotation)
		}

	case Pattern:
		if v.Pattern != nil {
			v.Pattern(n.Value, n.Negated, n.Annotation)
		}

	default:
		panic("unreachable")
	}
}

// VisitOperator is a convenience function that calls `f` on all operators `f`
// supplies the node's kind and operands.
func VisitOperator(nodes []Node, f func(kind OperatorKind, operands []Node)) {
	v := &Visitor{Operator: f}
	for _, n := range nodes {
		v.Visit(n)
	}
}

// VisitParameter is a convenience function that calls `f` on all parameters.
// `f` supplies the node's field, value, and whether the value is negated.
func VisitParameter(nodes []Node, f func(field, value string, negated bool, annotation Annotation)) {
	v := &Visitor{Parameter: f}
	for _, n := range nodes {
		v.Visit(n)
	}
}

// VisitPattern is a convenience function that calls `f` on all pattern nodes.
// `f` supplies the node's value, and whether the value is negated or quoted.
func VisitPattern(nodes []Node, f func(value string, negated bool, annotation Annotation)) {
	v := &Visitor{Pattern: f}
	for _, n := range nodes {
		v.Visit(n)
	}
}

// VisitField convenience function that calls `f` on all parameters whose field
// matches `field` argument. `f` supplies the node's value and whether the value
// is negated.
func VisitField(nodes []Node, field string, f func(value string, negated bool, annotation Annotation)) {
	VisitParameter(nodes, func(gotField, value string, negated bool, annotation Annotation) {
		if field == gotField {
			f(value, negated, annotation)
		}
	})
}

// VisitPredicate convenience function that calls `f` on all query predicates,
// supplying the node's field and predicate info.
func VisitPredicate(nodes []Node, f func(field, name, value string)) {
	VisitParameter(nodes, func(gotField, value string, _ bool, annotation Annotation) {
		if annotation.Labels.IsSet(IsPredicate) {
			name, predValue := ParseAsPredicate(value)
			f(gotField, name, predValue)
		}
	})
}
