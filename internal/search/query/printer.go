package query

import (
	"fmt"
	"strconv"
	"strings"
)

func stringHumanPattern(nodes []Node) string {
	var result []string
	for _, node := range nodes {
		switch n := node.(type) {
		case Pattern:
			v := n.Value
			if n.Annotation.Labels.IsSet(Quoted) {
				v = strconv.Quote(v)
			}
			if n.Negated {
				v = fmt.Sprintf("(not %s)", v)
			}
			result = append(result, v)
		case Operator:
			var nested []string
			for _, operand := range n.Operands {
				nested = append(nested, stringHumanPattern([]Node{operand}))
			}
			var separator string
			switch n.Kind {
			case Or:
				separator = " or "
			case And:
				separator = " and "
			}
			result = append(result, "("+strings.Join(nested, separator)+")")
		}
	}
	return strings.Join(result, "")
}

func stringHumanParameters(parameters []Parameter) string {
	var result []string
	for _, p := range parameters {
		v := p.Value
		if p.Annotation.Labels.IsSet(Quoted) {
			v = strconv.Quote(v)
		}
		if p.Negated {
			result = append(result, fmt.Sprintf("-%s:%s", p.Field, v))
		} else {
			result = append(result, fmt.Sprintf("%s:%s", p.Field, v))
		}
	}
	return strings.Join(result, " ")
}

// StringHuman creates a valid query string from a parsed query. It is used in
// contexts like query suggestions where we take the original query string of a
// user, parse it to a tree, modify the tree, and return a valid string
// representation. To faithfully preserve the meaning of the original tree,
// we need to consider whether to add operators like "and" contextually and must
// process the tree as a whole:
//
// repo:foo file:bar a and b -> preserve 'and', but do not insert 'and' between 'repo:foo file:bar'.
// repo:foo file:bar a b     -> do not insert any 'and', especially not between 'a b'.
//
// It strives to be syntax preserving, but may in some cases affect whitespace,
// operator capitalization, or parenthesized groupings. In very complex queries,
// additional 'and' operators may be inserted to segment parameters
// from patterns to preserve the original meaning.
func StringHuman(nodes []Node) string {
	parameters, pattern, err := PartitionSearchPattern(nodes)
	if err != nil {
		// We couldn't partition at this level in the tree, so recurse on operators until we can.
		var v []string
		for _, node := range nodes {
			if term, ok := node.(Operator); ok {
				var s []string
				for _, operand := range term.Operands {
					s = append(s, StringHuman([]Node{operand}))
				}
				if term.Kind == Or {
					v = append(v, "("+strings.Join(s, " or ")+")")
				} else if term.Kind == And {
					v = append(v, "("+strings.Join(s, " and ")+")")
				}
			}
		}
		return strings.Join(v, "")
	}
	if pattern == nil {
		return stringHumanParameters(parameters)
	}
	if len(parameters) == 0 {
		return stringHumanPattern([]Node{pattern})
	}
	return stringHumanParameters(parameters) + " " + stringHumanPattern([]Node{pattern})
}

// toString returns a string representation of a query's structure.
func toString(nodes []Node) string {
	var result []string
	for _, node := range nodes {
		result = append(result, node.String())
	}
	return strings.Join(result, " ")
}
