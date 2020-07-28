package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/src-d/enry/v2"
)

// exists traverses every node in nodes and returns early as soon as fn is satisfied.
func exists(nodes []Node, fn func(node Node) bool) bool {
	found := false
	for _, node := range nodes {
		if fn(node) {
			return true
		}
		if operator, ok := node.(Operator); ok {
			return exists(operator.Operands, fn)
		}
	}
	return found
}

// forAll traverses every node in nodes and returns whether all nodes satisfy fn.
func forAll(nodes []Node, fn func(node Node) bool) bool {
	sat := true
	for _, node := range nodes {
		if !fn(node) {
			return false
		}
		if operator, ok := node.(Operator); ok {
			return forAll(operator.Operands, fn)
		}
	}
	return sat
}

// isPatternExpression returns true if every leaf node in nodes is a search
// pattern expression.
func isPatternExpression(nodes []Node) bool {
	return !exists(nodes, func(node Node) bool {
		// Any non-pattern leaf, i.e., Parameter, falsifies the condition.
		_, ok := node.(Parameter)
		return ok
	})
}

// containsPattern returns true if any descendent of nodes is a search pattern.
func containsPattern(node Node) bool {
	return exists([]Node{node}, func(node Node) bool {
		_, ok := node.(Pattern)
		return ok
	})
}

// returns true if descendent of node contains and/or expressions.
func containsAndOrExpression(nodes []Node) bool {
	return exists(nodes, func(node Node) bool {
		term, ok := node.(Operator)
		return ok && (term.Kind == And || term.Kind == Or)
	})
}

// ContainsAndOrKeyword returns true if this query contains or- or and-
// keywords. It is a temporary signal to determine whether we can fallback to
// the older existing search functionality.
func ContainsAndOrKeyword(input string) bool {
	lower := strings.ToLower(input)
	return strings.Contains(lower, " and ") || strings.Contains(lower, " or ")
}

// ContainsRegexpMetasyntax returns true if a string is a valid regular
// expression and contains regex metasyntax (i.e., it is not a literal).
func ContainsRegexpMetasyntax(input string) bool {
	_, err := regexp.Compile(input)
	if err == nil {
		// It is a regexp. But does it contain metasyntax, or is it literal?
		if len(regexp.QuoteMeta(input)) != len(input) {
			return true
		}
	}
	return false
}

// processTopLevel processes the top level of a query. It validates that we can
// process the query with respect to and/or expressions on file content, but not
// otherwise for nested parameters.
func processTopLevel(nodes []Node) ([]Node, error) {
	if term, ok := nodes[0].(Operator); ok {
		if term.Kind == And && isPatternExpression([]Node{term}) {
			return nodes, nil
		} else if term.Kind == Or && isPatternExpression([]Node{term}) {
			return nodes, nil
		} else if term.Kind == And {
			return term.Operands, nil
		} else if term.Kind == Concat {
			return nodes, nil
		} else {
			return nil, &UnsupportedError{Msg: "cannot evaluate: unable to partition pure search pattern"}
		}
	}
	return nodes, nil
}

// PartitionSearchPattern partitions an and/or query into (1) a single search
// pattern expression and (2) other parameters that scope the evaluation of
// search patterns (e.g., to repos, files, etc.). It validates that a query
// contains at most one search pattern expression and that scope parameters do
// not contain nested expressions.
func PartitionSearchPattern(nodes []Node) (parameters []Node, pattern Node, err error) {
	if len(nodes) == 1 {
		nodes, err = processTopLevel(nodes)
		if err != nil {
			return nil, nil, err
		}
	}

	var patterns []Node
	for _, node := range nodes {
		if isPatternExpression([]Node{node}) {
			patterns = append(patterns, node)
		} else if term, ok := node.(Parameter); ok {
			parameters = append(parameters, term)
		} else {
			return nil, nil, &UnsupportedError{Msg: "cannot evaluate: unable to partition pure search pattern"}
		}
	}
	if len(patterns) > 1 {
		pattern = Operator{Kind: And, Operands: patterns}
	} else if len(patterns) == 1 {
		pattern = patterns[0]
	}

	return parameters, pattern, nil
}

// isPureSearchPattern implements a heuristic that returns true if buf, possibly
// containing whitespace or balanced parentheses, can be treated as a search
// pattern in the and/or grammar.
func isPureSearchPattern(buf []byte) bool {
	// Check if the balanced string we scanned is perhaps an and/or expression by parsing without the parensAsPatterns heuristic.
	try := &parser{buf: buf}
	result, err := try.parseOr()
	if err != nil {
		// This is not an and/or expression, but it is balanced. It
		// could be, e.g., (foo or). Reject this sort of pattern for now.
		return false
	}
	if try.balanced != 0 {
		return false
	}
	if containsAndOrExpression(result) {
		// The balanced string is an and/or expression in our grammar,
		// so it cannot be interpreted as a search pattern.
		return false
	}
	if !isPatternExpression(newOperator(result, Concat)) {
		// The balanced string contains other parameters, like
		// "repo:foo", which are not search patterns.
		return false
	}
	return true
}

// parseBool is like strconv.ParseBool except that it also accepts y, Y, yes,
// YES, Yes, n, N, no, NO, No.
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		b, err := strconv.ParseBool(s)
		if err != nil {
			err = fmt.Errorf("invalid boolean %q", s)
		}
		return b, err
	}
}

func validateField(field, value string, negated bool, seen map[string]struct{}) error {
	isNotNegated := func() error {
		if negated {
			return fmt.Errorf("field %q does not support negation", field)
		}
		return nil
	}

	isSingular := func() error {
		if _, notSingular := seen[field]; notSingular {
			return fmt.Errorf("field %q may not be used more than once", field)
		}
		return nil
	}

	isValidRegexp := func() error {
		if _, err := regexp.Compile(value); err != nil {
			return err
		}
		return nil
	}

	isBoolean := func() error {
		if _, err := parseBool(value); err != nil {
			return err
		}
		return nil
	}

	isNumber := func() error {
		count, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				return fmt.Errorf("field %s has a value that is out of range, try making it smaller", field)
			}
			return fmt.Errorf("field %s has value %[2]s, %[2]s is not a number", field, value)
		}
		if count <= 0 {
			return fmt.Errorf("field %s requires a positive number", field)
		}
		return nil
	}

	isLanguage := func() error {
		_, ok := enry.GetLanguageByAlias(value)
		if !ok {
			return fmt.Errorf("unknown language: %q", value)
		}
		return nil
	}

	isUnrecognizedField := func() error {
		return fmt.Errorf("unrecognized field %q", field)
	}

	satisfies := func(fns ...func() error) error {
		for _, fn := range fns {
			if err := fn(); err != nil {
				return err
			}
		}
		return nil
	}

	switch field {
	case
		FieldDefault:
		// Search patterns are not validated here, as it depends on the search type.
	case
		FieldCase:
		return satisfies(isSingular, isBoolean, isNotNegated)
	case
		FieldRepo:
		return satisfies(isValidRegexp)
	case
		FieldRepoGroup:
		return satisfies(isSingular, isNotNegated)
	case
		FieldFile:
		return satisfies(isValidRegexp)
	case
		FieldFork,
		FieldArchived:
		return satisfies(isSingular, isNotNegated)
	case
		FieldLang:
		return satisfies(isLanguage)
	case
		FieldType:
		return satisfies(isNotNegated)
	case
		FieldPatternType,
		FieldContent:
		return satisfies(isSingular, isNotNegated)
	case
		FieldRepoHasFile:
		return satisfies(isValidRegexp)
	case
		FieldRepoHasCommitAfter:
		return satisfies(isSingular, isNotNegated)
	case
		FieldBefore,
		FieldAfter:
		return satisfies(isNotNegated)
	case
		FieldAuthor,
		FieldCommitter,
		FieldMessage:
		return satisfies(isValidRegexp)
	case
		FieldIndex:
		return satisfies(isSingular, isNotNegated)
	case
		FieldCount:
		return satisfies(isSingular, isNumber, isNotNegated)
	case
		FieldStable:
		return satisfies(isSingular, isBoolean, isNotNegated)
	case
		FieldMax,
		FieldTimeout,
		FieldCombyRule:
		return satisfies(isSingular, isNotNegated)
	default:
		return isUnrecognizedField()
	}
	return nil
}

func validate(nodes []Node) error {
	var err error
	seen := map[string]struct{}{}
	VisitParameter(nodes, func(field, value string, negated bool) {
		if err != nil {
			return
		}
		err = validateField(field, value, negated, seen)
		seen[field] = struct{}{}
	})
	VisitPattern(nodes, func(value string, _ bool, annotation Annotation) {
		if annotation.Labels.isSet(Regexp) {
			if err != nil {
				return
			}
			_, err = regexp.Compile(value)
		}
	})
	return err
}
