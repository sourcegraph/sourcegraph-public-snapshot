package query

import (
	"encoding/json"
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
			if n.Annotation.Labels.IsSet(Regexp) {
				v = Delimit(v, '/')
			}
			if _, _, ok := ScanBalancedPattern([]byte(v)); !ok && !n.Annotation.Labels.IsSet(IsContent) && n.Annotation.Labels.IsSet(Literal) {
				v = fmt.Sprintf(`content:%s`, strconv.Quote(v))
				if n.Negated {
					v = "-" + v
				}
			} else if n.Annotation.Labels.IsSet(IsContent) {
				v = fmt.Sprintf("content:%s", v)
				if n.Negated {
					v = "-" + v
				}
			} else if n.Negated {
				v = fmt.Sprintf("(NOT %s)", v)
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
				separator = " OR "
			case And:
				separator = " AND "
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
		field := p.Field
		if p.Annotation.Labels.IsSet(IsAlias) {
			// Preserve alias for fields in the query for fields
			// with only one alias. We don't know which alias was in
			// the original query for fields that have multiple
			// aliases.
			switch p.Field {
			case FieldRepo:
				field = "r"
			case FieldAfter:
				field = "since"
			case FieldBefore:
				field = "until"
			case FieldRev:
				field = "revision"
			}
		}
		if p.Negated {
			result = append(result, fmt.Sprintf("-%s:%s", field, v))
		} else {
			result = append(result, fmt.Sprintf("%s:%s", field, v))
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
					v = append(v, "("+strings.Join(s, " OR ")+")")
				} else if term.Kind == And {
					v = append(v, "("+strings.Join(s, " AND ")+")")
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

func nodeToJSON(node Node) any {
	switch n := node.(type) {
	case Operator:
		var jsons []any
		for _, o := range n.Operands {
			jsons = append(jsons, nodeToJSON(o))
		}

		switch n.Kind {
		case And:
			return struct {
				And []any `json:"and"`
			}{
				And: jsons,
			}
		case Or:
			return struct {
				Or []any `json:"or"`
			}{
				Or: jsons,
			}
		case Concat:
			// Concat should already be processed at this point, or
			// the original query expresses something that is not
			// supported. We just return the parse tree anyway.
			return struct {
				Concat []any `json:"concat"`
			}{
				Concat: jsons,
			}
		}
	case Parameter:
		return struct {
			Field   string   `json:"field"`
			Value   string   `json:"value"`
			Negated bool     `json:"negated"`
			Labels  []string `json:"labels"`
			Range   Range    `json:"range"`
		}{
			Field:   n.Field,
			Value:   n.Value,
			Negated: n.Negated,
			Labels:  n.Annotation.Labels.String(),
			Range:   n.Annotation.Range,
		}
	case Pattern:
		return struct {
			Value   string   `json:"value"`
			Negated bool     `json:"negated"`
			Labels  []string `json:"labels"`
			Range   Range    `json:"range"`
		}{
			Value:   n.Value,
			Negated: n.Negated,
			Labels:  n.Annotation.Labels.String(),
			Range:   n.Annotation.Range,
		}
	}
	// unreachable.
	return struct{}{}
}

func nodesToJSON(q Q) []any {
	var jsons []any
	for _, node := range q {
		jsons = append(jsons, nodeToJSON(node))
	}
	return jsons
}

func ToJSON(q Q) (string, error) {
	j, err := json.Marshal(nodesToJSON(q))
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func PrettyJSON(q Q) (string, error) {
	j, err := json.MarshalIndent(nodesToJSON(q), "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}
