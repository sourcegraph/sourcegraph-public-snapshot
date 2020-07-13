package query

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SubstituteAliases substitutes field name aliases for their canonical names.
func SubstituteAliases(nodes []Node) []Node {
	aliases := map[string]string{
		"r":        FieldRepo,
		"g":        FieldRepoGroup,
		"f":        FieldFile,
		"l":        FieldLang,
		"language": FieldLang,
		"since":    FieldAfter,
		"until":    FieldBefore,
		"m":        FieldMessage,
		"msg":      FieldMessage,
	}
	return MapParameter(nodes, func(field, value string, negated bool, annotation Annotation) Node {
		if field == "content" {
			return Pattern{Value: value, Negated: negated, Annotation: annotation}
		}
		if canonical, ok := aliases[field]; ok {
			field = canonical
		}
		return Parameter{Field: field, Value: value, Negated: negated, Annotation: annotation}
	})
}

// LowercaseFieldNames performs strings.ToLower on every field name.
func LowercaseFieldNames(nodes []Node) []Node {
	return MapParameter(nodes, func(field, value string, negated bool, annotation Annotation) Node {
		return Parameter{Field: strings.ToLower(field), Value: value, Negated: negated, Annotation: annotation}
	})
}

// translateRange translates character classes and ranges
func translateRange(r []rune, ix int) (int, string) {
	sb := strings.Builder{}

	// the first character of a range or character set
	// is special because
	//   * ranges or character sets cannot be empty
	//   * we have to translate negation from ^ to !
	if r[ix] == '^' {
		sb.WriteRune('!')
	} else {
		// since character sets cannot be empty,
		// a [ followed directly by a ] means ] has to
		// be interpreted literally.
		sb.WriteRune(r[ix])
	}
	ix++
	for ix < len(r) && r[ix] != ']' {
		sb.WriteRune(r[ix])
		ix++
	}

	return len(sb.String()), sb.String()
}

// translateGlobToRegex converts a globbing string to a regex
// we support: *, ?, character classes [...], ranges [A-F]
func translateGlobToRegex(value string) string {
	r := []rune(value)
	l := len(r)
	sb := strings.Builder{}

	i := 0
	// add regex anchor "^" if glob does not start with *
	if r[i] != '*' {
		sb.WriteRune('^')
	}

	for i = 0; i < l; i++ {
		switch r[i] {
		case '*':
			sb.WriteString(".*?")
			// skip repeated '*'
			for i < l-1 && r[i+1] == '*' {
				i++
			}
		case '?':
			sb.WriteRune('.')
		case '\\':
			// handle escaped special characters
			sb.WriteRune('\\')
			i++
			sb.WriteRune(r[i])
		case '[':
			sb.WriteRune('[')
			i++

			advanced, s := translateRange(r, i)
			i += advanced

			sb.WriteString(s)
			sb.WriteRune(']')
		default:
			sb.WriteString(regexp.QuoteMeta(string(r[i])))
		}
	}

	// add regex anchor "$" if glob doesn't end with *
	if r[len(r)-1] != '*' {
		sb.WriteRune('$')
	}
	return sb.String()
}

func isValidRegexp(value string) bool {
	if _, err := regexp.Compile(value); err != nil {
		return false
	}
	return true
}

// globToRegex substitutes glob with regexp for fields supporting regexp
// the value is left unchanged if
// * is a valid regexp
// * the translated value is not a valid regexp
func globToRegex(nodes []Node) []Node {
	return MapParameter(nodes, func(field, value string, negated bool, annotation Annotation) Node {
		if field == FieldRepo || field == FieldFile || field == FieldRepoHasFile {
			tempValue := translateGlobToRegex(value)
			if isValidRegexp(tempValue) {
				value = tempValue
			}
		}
		return Parameter{Field: field, Value: value, Negated: negated, Annotation: annotation}
	})
}

// Hoist is a heuristic that rewrites simple but possibly ambiguous queries. It
// changes certain expressions in a way that some consider to be more natural.
// For example, the following query without parentheses is interpreted as
// follows in the grammar:
//
// repo:foo a or b and c => (repo:foo a) or ((b) and (c))
//
// This function rewrites the above expression as follows:
//
// repo:foo a or b and c => repo:foo (a or b and c)
//
// Any number of field:value parameters may occur before and after the pattern
// expression separated by or- or and-operators, and these are hoisted out. The
// pattern expression must be contiguous. If not, we want to preserve the
// default interpretation, which corresponds more naturally to groupings with
// field parameters, i.e.,
//
// repo:foo a or b or repo:bar c => (repo:foo a) or (b) or (repo:bar c)
func Hoist(nodes []Node) ([]Node, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("heuristic requires one top-level expression")
	}

	expression, ok := nodes[0].(Operator)
	if !ok || expression.Kind == Concat {
		return nil, fmt.Errorf("heuristic requires top-level and- or or-expression")
	}

	n := len(expression.Operands)
	var pattern []Node
	var scopeParameters []Node
	for i, node := range expression.Operands {
		if i == 0 || i == n-1 {
			scopePart, patternPart, err := PartitionSearchPattern([]Node{node})
			if err != nil || patternPart == nil {
				return nil, errors.New("could not partition first or last expression")
			}
			pattern = append(pattern, patternPart)
			scopeParameters = append(scopeParameters, scopePart...)
			continue
		}
		if !isPatternExpression([]Node{node}) {
			return nil, fmt.Errorf("inner expression %s is not a pure pattern expression", node.String())
		}
		pattern = append(pattern, node)
	}
	pattern = MapPattern(pattern, func(value string, negated bool, annotation Annotation) Node {
		annotation.Labels |= HeuristicHoisted
		return Pattern{Value: value, Negated: negated, Annotation: annotation}
	})
	return append(scopeParameters, newOperator(pattern, expression.Kind)...), nil
}

// SearchUppercase adds case:yes to queries if any pattern is mixed-case.
func SearchUppercase(nodes []Node) []Node {
	var foundMixedCase bool
	VisitPattern(nodes, func(value string, _ bool, _ Annotation) {
		if match := containsUppercase(value); match {
			foundMixedCase = true
		}
	})
	if foundMixedCase {
		nodes = append(nodes, Parameter{Field: "case", Value: "yes"})
		return newOperator(nodes, And)
	}
	return nodes
}

func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// partition partitions nodes into left and right groups. A node is put in the
// left group if fn evaluates to true, or in the right group if fn evaluates to false.
func partition(nodes []Node, fn func(node Node) bool) (left, right []Node) {
	for _, node := range nodes {
		if fn(node) {
			left = append(left, node)
		} else {
			right = append(right, node)
		}
	}
	return left, right
}

func substituteOrForRegexp(nodes []Node) []Node {
	isPattern := func(node Node) bool {
		if pattern, ok := node.(Pattern); ok && !pattern.Negated {
			return true
		}
		return false
	}
	new := []Node{}
	for _, node := range nodes {
		switch v := node.(type) {
		case Operator:
			if v.Kind == Or {
				patterns, rest := partition(v.Operands, isPattern)
				var values []string
				for _, node := range patterns {
					values = append(values, node.(Pattern).Value)
				}
				valueString := "(" + strings.Join(values, ")|(") + ")"
				new = append(new, Pattern{Value: valueString})
				if len(rest) > 0 {
					rest = substituteOrForRegexp(rest)
					new = newOperator(append(new, rest...), Or)
				}
			} else {
				new = append(new, newOperator(substituteOrForRegexp(v.Operands), v.Kind)...)
			}
		case Parameter, Pattern:
			new = append(new, node)
		}
	}
	return new
}

// substituteConcat reduces a concatenation of patterns to a separator-separated string.
func substituteConcat(nodes []Node, separator string) []Node {
	isPattern := func(node Node) bool {
		if pattern, ok := node.(Pattern); ok && !pattern.Negated {
			return true
		}
		return false
	}
	new := []Node{}
	for _, node := range nodes {
		switch v := node.(type) {
		case Parameter, Pattern:
			new = append(new, node)
		case Operator:
			if v.Kind == Concat {
				// Merge consecutive patterns.
				previous := v.Operands[0]
				merged := Pattern{}
				if p, ok := previous.(Pattern); ok {
					merged = p
				}
				for _, node := range v.Operands[1:] {
					if isPattern(node) && isPattern(previous) {
						p := node.(Pattern)
						if merged.Value != "" {
							merged.Annotation.Labels |= p.Annotation.Labels
							merged = Pattern{
								Value:      merged.Value + separator + p.Value,
								Annotation: merged.Annotation,
							}
						} else {
							// Base case.
							merged = Pattern{Value: p.Value}
						}
						previous = node
						continue
					}
					if merged.Value != "" {
						new = append(new, merged)
						merged = Pattern{}
					}
					new = append(new, substituteConcat([]Node{node}, separator)...)
				}
				if merged.Value != "" {
					new = append(new, merged)
					merged = Pattern{}
				}
			} else {
				new = append(new, newOperator(substituteConcat(v.Operands, separator), v.Kind)...)
			}
		}
	}
	return new
}

// EmptyGroupsToLiteral is a heuristic used in the context of regular expression
// search. It labels any pattern containing "()" as a literal pattern since in
// regex it implies the empty string, which is meaningless as a search query and
// probably not what the user intended.
func EmptyGroupsToLiteral(nodes []Node) []Node {
	return MapPattern(nodes, func(value string, negated bool, annotation Annotation) Node {
		if ok, _ := regexp.MatchString(`\(\)`, value); ok {
			annotation.Labels.set(Literal)
			annotation.Labels.unset(Regexp)
		}
		return Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})
}

// Map pipes query through one or more query transformer functions.
func Map(query []Node, fns ...func([]Node) []Node) []Node {
	for _, fn := range fns {
		query = fn(query)
	}
	return query
}
