package protocol

import (
	"sort"
)

var defaultReducers = []pass{
	propagateBoolean,
	rewriteConjunctive,
	flatten,
	mergeOrRegexp,
	sortAndByCost,
}

// Reduce simplifies and optimizes a query using the default reducers
func Reduce(n Node) Node {
	return ReduceWith(n, defaultReducers...)
}

// ReduceWith simplifies and optimizes a query using the provided reducers.
// It visits nodes in a depth-first manner.
func ReduceWith(n Node, reducers ...pass) Node {
	switch v := n.(type) {
	case *Operator:
		for i, operand := range v.Operands {
			v.Operands[i] = ReduceWith(operand, reducers...)
		}
	}

	for _, f := range reducers {
		n = f(n)
	}
	return n
}

type pass func(Node) Node

// propagateBoolean simplifies any nodes containing constant nodes
func propagateBoolean(n Node) Node {
	operator, ok := n.(*Operator)
	if !ok {
		return n
	}

	switch operator.Kind {
	case Not:
		// Negate the constant and propagate it upwards
		if c, ok := operator.Operands[0].(*Boolean); ok {
			// Not(false) => true
			return &Boolean{!c.Value}
		}
		return n
	case And:
		filteredOperands := operator.Operands[:0]
		for _, operand := range operator.Operands {
			if c, ok := operand.(*Boolean); ok {
				if !c.Value {
					// And(x, y, false) => false
					return operand
				}
				// And(x, y, true) => And(x, y)
			} else {
				filteredOperands = append(filteredOperands, operand)
			}
		}
		return newOperator(And, filteredOperands...)
	case Or:
		filteredOperands := operator.Operands[:0]
		for _, operand := range operator.Operands {
			if c, ok := operand.(*Boolean); ok {
				if c.Value {
					// Or(x, y, true) => true
					return operand
				}
				// Or(x, y, false) => Or(x, y)
			} else {
				filteredOperands = append(filteredOperands, operand)
			}
		}
		return newOperator(Or, filteredOperands...)
	default:
		panic("unknown operator kind")
	}
}

// rewriteConjunctive does a best-effort attempt at rewriting a node from a top-level disjunctive
// to a conjunctive. For example, Or(And(x, y), z) => And(Or(x, z), Or(y, z)). This allows
// us to short-circuit more quickly. This does not necessarily get us to conjunctive normal form
// because we do not distribute Not operators due to super-exponential query size.
func rewriteConjunctive(n Node) Node {
	operator, ok := n.(*Operator)
	if !ok || operator.Kind != Or {
		return n
	}

	var andOperands [][]Node
	siblings := operator.Operands[:0]
	for _, operand := range operator.Operands {
		if o, ok := operand.(*Operator); ok && o.Kind == And {
			andOperands = append(andOperands, o.Operands)
		} else {
			siblings = append(siblings, operand)
		}
	}

	if len(andOperands) == 0 {
		// No nested and operands, so don't modify the node
		return n
	}

	distributed := distribute(andOperands, siblings)
	newAndOperands := make([]Node, 0, len(distributed))
	for _, d := range distributed {
		newAndOperands = append(newAndOperands, newOperator(Or, d...))
	}
	return newOperator(And, newAndOperands...)
}

// distribute will expand prefixes into every choice of one node
// from each prefix, then append that set to each of the nodes.
func distribute(prefixes [][]Node, nodes []Node) [][]Node {
	if len(prefixes) == 0 {
		return [][]Node{nodes}
	}

	distributed := distribute(prefixes[1:], nodes)
	res := make([][]Node, 0, len(distributed)*len(prefixes[0]))
	for _, p := range prefixes[0] {
		for _, d := range distributed {
			newRow := make([]Node, len(d), len(d)+1)
			copy(newRow, d)
			res = append(res, append(newRow, p))
		}
	}
	return res
}

// flatten will flatten And children of And operators and Or children of Or operators
func flatten(n Node) Node {
	operator, ok := n.(*Operator)
	if !ok || operator.Kind == Not {
		return n
	}

	flattened := make([]Node, 0, len(operator.Operands))
	for _, operand := range operator.Operands {
		if nestedOperator, ok := operand.(*Operator); ok && nestedOperator.Kind == operator.Kind {
			flattened = append(flattened, nestedOperator.Operands...)
		} else {
			flattened = append(flattened, operand)
		}
	}

	return newOperator(operator.Kind, flattened...)
}

// mergeOrRegexp will merge regexp nodes of the same type in an Or operand,
// allowing us to run only a single regex search over the field rather than multiple.
func mergeOrRegexp(n Node) Node {
	operator, ok := n.(*Operator)
	if !ok || operator.Kind != Or {
		return n
	}

	union := func(left, right string) string {
		return "(?:" + left + ")|(?:" + right + ")"
	}

	unmergeable := operator.Operands[:0]
	mergeable := map[any]Node{}
	for _, operand := range operator.Operands {
		switch v := operand.(type) {
		case *AuthorMatches:
			key := AuthorMatches{IgnoreCase: v.IgnoreCase}
			if prev, ok := mergeable[key]; ok {
				mergeable[key] = &AuthorMatches{
					Expr:       union(prev.(*AuthorMatches).Expr, v.Expr),
					IgnoreCase: v.IgnoreCase,
				}
			} else {
				mergeable[key] = v
			}
		case *CommitterMatches:
			key := CommitterMatches{IgnoreCase: v.IgnoreCase}
			if prev, ok := mergeable[key]; ok {
				mergeable[key] = &CommitterMatches{
					Expr:       union(prev.(*CommitterMatches).Expr, v.Expr),
					IgnoreCase: v.IgnoreCase,
				}
			} else {
				mergeable[key] = v
			}
		case *MessageMatches:
			key := MessageMatches{IgnoreCase: v.IgnoreCase}
			if prev, ok := mergeable[key]; ok {
				mergeable[key] = &MessageMatches{
					Expr:       union(prev.(*MessageMatches).Expr, v.Expr),
					IgnoreCase: v.IgnoreCase,
				}
			} else {
				mergeable[key] = v
			}
		case *DiffMatches:
			key := DiffMatches{IgnoreCase: v.IgnoreCase}
			if prev, ok := mergeable[key]; ok {
				mergeable[key] = &DiffMatches{
					Expr:       union(prev.(*DiffMatches).Expr, v.Expr),
					IgnoreCase: v.IgnoreCase,
				}
			} else {
				mergeable[key] = v
			}
		case *DiffModifiesFile:
			key := DiffModifiesFile{IgnoreCase: v.IgnoreCase}
			if prev, ok := mergeable[key]; ok {
				mergeable[key] = &DiffModifiesFile{
					Expr:       union(prev.(*DiffModifiesFile).Expr, v.Expr),
					IgnoreCase: v.IgnoreCase,
				}
			} else {
				mergeable[key] = v
			}
		default:
			unmergeable = append(unmergeable, operand)
		}
	}

	// Re-combine the merged operands into our unmerged operands
	res := unmergeable
	for _, m := range mergeable {
		res = append(res, m)
	}
	return newOperator(Or, res...)
}

// estimatedNodeCost estimates node cost in a completely unscientific way.
// The numbers are largely educated speculation, but it doesn't matter that much
// since we mostly care about nodes that generate diffs being put last.
func estimatedNodeCost(n Node) float64 {
	switch v := n.(type) {
	case *Operator:
		sum := 0.0
		for _, operand := range v.Operands {
			sum += estimatedNodeCost(operand)
		}
		return sum
	case *Boolean:
		return 0
	case *CommitBefore, *CommitAfter:
		return 1
	case *AuthorMatches, *CommitterMatches:
		return 5
	case *MessageMatches:
		return 10
	case *DiffModifiesFile:
		return 1000
	case *DiffMatches:
		return 10000
	default:
		return 1
	}
}

// sortAndByCost sorts the operands of And nodes by their estimated cost
// so more expensive nodes are excluded by short-circuiting when possible.
// Or nodes are not short-circuited, so this does not sort Or nodes.
func sortAndByCost(n Node) Node {
	operator, ok := n.(*Operator)
	if !ok || operator.Kind != And {
		return n
	}

	sort.Slice(operator.Operands, func(i, j int) bool {
		return estimatedNodeCost(operator.Operands[i]) < estimatedNodeCost(operator.Operands[j])
	})
	return operator
}
